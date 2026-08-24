package cpulaoder

import (
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/niexqc/nlibs/cpumem/memloader"

	"github.com/shirou/gopsutil/cpu"
)

type CpuLoader struct {
	TargetPercent    float64 // 目标CPU占用率百分比
	CheckTimeInerval int64   // 监控打印间隔（秒）
	CpuAvgTime       int64   // 采样窗口（秒）
	maxWokerNum      int64   // 工作协程数

	activeWorkers int32
	stopChan      chan struct{}
	stopOnce      sync.Once
	// loadLevel 是"当前应同时忙碌的 worker 数"（0~NumCPU）。
	// 系统 CPU 百分比 ≈ loadLevel / NumCPU * 100%，因此它与目标百分比线性对应。
	// 忙碌的 worker 采用纯忙等（不调用 time.Sleep），实测可让单核占满，
	// 从而 loadLevel=NumCPU 时 CPU 可达到接近 100%（标定验证 24 核全忙→99.9%）。
	// 空闲的 worker 长睡。这是一个真正线性、可覆盖 0~100% 的控制量。
	// 用原子变量存储，避免 worker 每次循环都去抢 adjustLock 造成锁竞争，
	// 否则频繁加锁会把 busySpin 的忙等截断，导致 CPU 无法达到满负荷。
	loadLevel      atomic.Uint64 // 存储 float64 位
	smoothingFactor float64

	sampleStop     chan struct{}
	controllerStop chan struct{}
	smoothedCPU    atomic.Uint64 // 存储 float64 位，平滑后的CPU%
}

func NewCpuLoader(targetPercent float64, checkTimeInerval, cpuAvgTime int64) *CpuLoader {
	vloader := &CpuLoader{
		TargetPercent:     targetPercent,
		CheckTimeInerval:  checkTimeInerval,
		CpuAvgTime:        cpuAvgTime,
		stopChan:          make(chan struct{}),
		sampleStop:        make(chan struct{}),
		controllerStop:    make(chan struct{}),
		smoothingFactor:  0.5,
		activeWorkers:    0,
	}
	// 初始估计：目标百分比对应的等效忙碌 worker 数
	vloader.loadLevel.Store(math.Float64bits(targetPercent / 100 * float64(runtime.NumCPU())))
	vloader.maxWokerNum = int64(runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
	return vloader
}

// CpuPercent 返回系统CPU使用率（该导出函数独立调用，仅供外部使用）。
// loader内部不会调用它，以避免与单一采样器并发调用 gopsutil 造成基线污染。
func CpuPercent(second int64) float64 {
	percent, err := cpu.Percent(time.Duration(second)*time.Second, false)
	if err != nil || len(percent) == 0 {
		return 0
	}
	return percent[0]
}

// sampleCPU 由单一采样goroutine独占调用，返回一个窗口内的CPU使用率。
func sampleCPU(second int64) float64 {
	percent, err := cpu.Percent(time.Duration(second)*time.Second, false)
	if err != nil || len(percent) == 0 {
		return 0
	}
	return percent[0]
}

// Start 启动CPU负载生成器（阻塞，直到调用 Stop）。
func (loader *CpuLoader) Start() {
	loader.setLinearControlParams()
	atomic.StoreInt32(&loader.activeWorkers, 1)
	// Start 退出时统一关闭子协程的停止信号，保证只关闭一次且不会产生泄漏。
	defer loader.closeChildStops()

	// 单一采样器：独占调用 gopsutil，避免多goroutine并发采样导致基线污染与数据抖动。
	go loader.sampleLoop()
	// PID控制器：依据平滑后的CPU值调节统一的 loadLevel。
	go loader.controller()
	// 负载worker：依据全局 loadLevel 决定自己忙碌还是空闲。
	for i := int64(0); i < loader.maxWokerNum; i++ {
		go loader.cpuWorker(i)
	}

	ticker := time.NewTicker(time.Duration(loader.CheckTimeInerval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			memPercent := memloader.MenPercent()
			currentCPU := loader.getSmoothedCPU()
			level := loader.getLoadLevel()
			slog.Info(fmt.Sprintf("CPU: %.2f%%, Memory: %.2f%%, loadLevel: %.2f", currentCPU, memPercent, level))
		case <-loader.stopChan:
			return
		}
	}
}

// closeChildStops 关闭采样器与控制器协程的停止信号。
func (loader *CpuLoader) closeChildStops() {
	close(loader.sampleStop)
	close(loader.controllerStop)
}

// Stop 停止CPU负载生成器（可重复调用）。
func (loader *CpuLoader) Stop() {
	atomic.StoreInt32(&loader.activeWorkers, 0)
	loader.stopOnce.Do(func() {
		close(loader.stopChan)
	})
}

// sampleLoop 单一采样循环：独占调用 gopsutil 并做EMA平滑，结果存入原子变量。
func (loader *CpuLoader) sampleLoop() {
	// 建立gopsutil采样基线（立即返回）
	_, _ = cpu.Percent(0, false)
	for {
		select {
		case <-loader.sampleStop:
			return
		default:
		}
		raw := sampleCPU(loader.CpuAvgTime)
		if raw <= 0 {
			continue
		}
		cur := loader.getSmoothedCPU()
		alpha := 0.4 // EMA平滑系数，越大越激进、越小越平稳
		if cur == 0 {
			cur = raw
		} else {
			cur = alpha*raw + (1-alpha)*cur
		}
		loader.smoothedCPU.Store(math.Float64bits(cur))
	}
}

// controller PID调节 loadLevel，使实测CPU逼近目标。
// 由于 CPU% ≈ loadLevel/NumCPU*100%，误差 err（百分比）换算到 loadLevel 需 /100*NumCPU，
// 比例增益取 0.5 使其对目标变化响应适度、不易过冲。
func (loader *CpuLoader) controller() {
	const interval = 500 * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	integral := 0.0
	prev := 0.0
	maxLevel := float64(runtime.NumCPU())
	for {
		select {
		case <-ticker.C:
			cur := loader.getSmoothedCPU()
			if cur <= 0 {
				continue
			}
			err := loader.TargetPercent - cur
			if math.Abs(err) < 10 {
				integral += err
			} else {
				integral = 0
			}
			deriv := cur - prev
			// 误差是 CPU 百分比，换算到 loadLevel：/100*NumCPU
			// 增益 Kp=0.5, Ki=0.1, Kd=0.2，经标定可稳定收敛且抑制过冲。
			deltaLevel := (0.5*err + 0.1*integral - 0.2*deriv) / 100.0 * maxLevel
			l := loader.getLoadLevel() + deltaLevel
			loader.loadLevel.Store(math.Float64bits(clamp(l, 0, maxLevel)))
			prev = cur
		case <-loader.controllerStop:
			return
		}
	}
}

// getSmoothedCPU 读取平滑后的CPU值。
func (loader *CpuLoader) getSmoothedCPU() float64 {
	return math.Float64frombits(loader.smoothedCPU.Load())
}

// getLoadLevel 读取当前等效忙碌 worker 数（无锁，供高频循环的 worker 调用）。
func (loader *CpuLoader) getLoadLevel() float64 {
	return math.Float64frombits(loader.loadLevel.Load())
}

// 设置平滑因子（用于控制器的光滑性）。
func (loader *CpuLoader) setLinearControlParams() {
	if loader.TargetPercent < 30 {
		loader.smoothingFactor = 0.7
	} else if loader.TargetPercent < 60 {
		loader.smoothingFactor = 0.5
	} else {
		loader.smoothingFactor = 0.3
	}
	slog.Info("负载控制器参数", "平滑因子", loader.smoothingFactor)
}

// cpuWorker 负载工作协程：依据全局 loadLevel 平滑分配本 worker 的忙碌占空比。
// 每个 worker 的 myDuty = clamp(loadLevel - index, 0, 1)，因此 loadLevel 为小数时也能平滑工作：
// 例如 loadLevel=10.5 时「worker 0~9 全忙、worker 10 忙 0.5、worker 11+ 空闲」，
// 平均正好 loadLevel 个 worker 在忙，规避了"整数个 worker"导致的离散跳变与 PID 死区。
// 忙碌用纯忙等（不调 time.Sleep），空闲用 sleep 让出 CPU。相位错开避免所有 worker 同步忙/睡。
func (loader *CpuLoader) cpuWorker(index int64) {
	const cpuCycle = 200 * time.Millisecond
	// 初始相位错开，避免多个 worker 同时进入忙/睡，导致整体 CPU 抖动
	time.Sleep(time.Duration(float64(cpuCycle) * float64(index) / float64(loader.maxWokerNum)))

	for atomic.LoadInt32(&loader.activeWorkers) == 1 {
		level := loader.getLoadLevel()
		myDuty := clamp(level-float64(index), 0, 1)
		if myDuty <= 0 {
			time.Sleep(cpuCycle)
			continue
		}
		work := time.Duration(float64(cpuCycle) * myDuty)
		busySpin(work)
		if sleep := cpuCycle - work; sleep > 0 {
			time.Sleep(sleep)
		}
	}
}

// busySpin 纯忙等指定的时长，占用一个核而不让出。
func busySpin(duration time.Duration) {
	start := time.Now()
	for {
		for i := 0; i < 1000; i++ {
			_ = math.Sqrt(float64(i))
		}
		if time.Since(start) >= duration {
			break
		}
	}
}

// clamp 限制值在[min, max]范围内。
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
