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
	TargetPercent    float64
	CheckTimeInerval int64 //间隔多少秒检查
	CpuAvgTime       int64 // 获取多少秒的平均值
	maxWokerNum      int64 //最大负载的协程数

	activeWorkers   int32         // 原子计数器
	stopChan        chan struct{} // 停止信号通道
	loadFactor      float64       // 动态负载因子
	adjustLock      sync.Mutex    // 调节器锁
	smoothingFactor float64       // 平滑因子
}

func NewCpuLoader(targetPercent float64, checkTimeInerval, cpuAvgTime int64) *CpuLoader {
	vloader := &CpuLoader{
		TargetPercent:    targetPercent,
		CheckTimeInerval: checkTimeInerval,
		CpuAvgTime:       cpuAvgTime,
		stopChan:         make(chan struct{}),
		loadFactor:       1.0,
		smoothingFactor:  0.5, // 默认平滑系数
	}
	vloader.maxWokerNum = int64(runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
	return vloader
}

func CpuPercent(second int64) float64 {
	percent, _ := cpu.Percent(time.Duration(second)*time.Second, false)
	currentCPU := percent[0]
	return currentCPU
}

func (loader *CpuLoader) Start() {
	ticker := time.NewTicker(time.Duration(loader.CheckTimeInerval) * time.Second)
	defer ticker.Stop()
	// 设置线性控制参数
	loader.setLinearControlParams()

	for range ticker.C {
		memPercent := memloader.MenPercent()
		currentCPU := CpuPercent(loader.CpuAvgTime)
		slog.Info(fmt.Sprintf("CPU: %.2f%%, Memory: %.2f%%", currentCPU, memPercent))

		// 更平滑的触发逻辑
		if currentCPU < loader.TargetPercent*0.95 && atomic.LoadInt32(&loader.activeWorkers) == 0 {
			loader.startGenCpuLoad()
		} else if currentCPU > loader.TargetPercent*1.05 && atomic.LoadInt32(&loader.activeWorkers) == 1 {
			loader.stopCpuLoad()
		}
	}
}

func (loader *CpuLoader) startGenCpuLoad() {
	atomic.StoreInt32(&loader.activeWorkers, 1)
	slog.Info("🚀 启动CPU负载生成器")
	// 启动负载生成协程
	for i := int64(0); i < loader.maxWokerNum; i++ {
		go loader.cpuWorker()
	}
	// 启动线性调节器
	go loader.linearAdjuster()
}

func (loader *CpuLoader) stopCpuLoad() {
	if atomic.LoadInt32(&loader.activeWorkers) == 1 {
		atomic.StoreInt32(&loader.activeWorkers, 0)
		close(loader.stopChan)
		loader.stopChan = make(chan struct{}) // 重置通道
		slog.Info("🛑 停止CPU负载生成器")
	}
}

// 设置线性控制参数
func (loader *CpuLoader) setLinearControlParams() {
	// 根据目标负载率设置平滑因子
	// 低负载目标使用更强的平滑效果，高负载目标使用更快的响应
	if loader.TargetPercent < 30 {
		loader.smoothingFactor = 0.7
	} else if loader.TargetPercent < 60 {
		loader.smoothingFactor = 0.5
	} else {
		loader.smoothingFactor = 0.3
	}
	slog.Info("负载控制器参数", "平滑因子", loader.smoothingFactor)
}

// 线性负载工作器
func (loader *CpuLoader) cpuWorker() {
	var (
		workTime   time.Duration
		sleepTime  time.Duration
		cycleCount int
	)

	// 初始化工作/休眠时间比例为1:1
	workTime = 100 * time.Millisecond
	sleepTime = 100 * time.Millisecond

	for atomic.LoadInt32(&loader.activeWorkers) == 1 {
		// 动态调整工作/休眠比例
		factor := loader.loadFactor
		scaledWorkTime := time.Duration(float64(workTime) * factor)
		scaledSleepTime := time.Duration(float64(sleepTime) * (2 - factor))

		// 线性负载周期
		cycleStart := time.Now()
		calculateLinear(scaledWorkTime)
		cycleEnd := time.Now()

		// 精确控制周期时间
		actualWorkTime := cycleEnd.Sub(cycleStart)
		targetDuration := scaledWorkTime + scaledSleepTime
		remainingSleep := targetDuration - actualWorkTime

		if remainingSleep > 0 {
			time.Sleep(remainingSleep)
		}

		// 每10个周期微调参数
		cycleCount++
		if cycleCount%10 == 0 {
			// 根据实际负载情况微调比例
			current := CpuPercent(1)
			if current < loader.TargetPercent*0.95 {
				workTime = time.Duration(float64(workTime) * 1.05)
			} else if current > loader.TargetPercent*1.05 {
				sleepTime = time.Duration(float64(sleepTime) * 1.05)
			}
		}
	}
}

// 线性调节器
func (loader *CpuLoader) linearAdjuster() {
	const adjustmentInterval = 500 * time.Millisecond
	ticker := time.NewTicker(adjustmentInterval)
	defer ticker.Stop()

	prevCPU := 0.0
	integral := 0.0

	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&loader.activeWorkers) == 0 {
				return
			}

			currentCPU := CpuPercent(1)
			if math.IsNaN(currentCPU) || currentCPU <= 0 {
				continue
			}

			// 使用PID算法进行线性调节
			errorVal := loader.TargetPercent - currentCPU
			derivative := (currentCPU - prevCPU) / adjustmentInterval.Seconds()

			// PID参数 (比例、积分、微分)
			Kp := 0.8
			Ki := 0.1
			Kd := 0.01

			// 防止积分饱和
			if math.Abs(errorVal) < 10 {
				integral += errorVal * adjustmentInterval.Seconds()
			} else {
				integral = 0
			}

			// 计算调节量
			adjustment := Kp*errorVal + Ki*integral - Kd*derivative

			// 应用平滑变换
			loader.adjustLock.Lock()
			loader.loadFactor = loader.smoothingFactor*loader.loadFactor + (1-loader.smoothingFactor)*clamp(1+adjustment*0.01, 0.3, 2.0)
			loader.adjustLock.Unlock()

			prevCPU = currentCPU

		case <-loader.stopChan:
			return
		}
	}
}

// 线性计算任务
func calculateLinear(duration time.Duration) {
	start := time.Now()
	for {
		// 固定比例的计算任务
		for i := 0; i < 1000; i++ {
			_ = math.Sqrt(float64(i))
		}

		// 精确的时间控制
		if time.Since(start) >= duration {
			break
		}
	}
}

// 辅助函数：限制值在[min, max]范围内
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
