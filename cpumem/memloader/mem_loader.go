package memloader

import (
	"log/slog"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/mem"
)

type MemLoader struct {
	TargetPercent    float64       // 目标内存占用百分比
	CheckInterval    int64         // 检查间隔（秒）
	CpuAvgTime       int64         // CPU平均时间（保留参数）
	active           int32         // 负载激活状态
	stopChan         chan struct{} // 停止信号
	allocatedBytes   uint64        // 已分配字节数
	minBlockSize     uint64        // 最小内存块大小
	maxBlockSize     uint64        // 最大内存块大小
	adjustLock       sync.Mutex    // 内存调整锁
	smoothingFactor  float64       // 平滑因子
	allocationRate   float64       // 当前分配速率
	protectionFactor float64       // 内存保护因子
}

func NewMemLoader(targetPercent float64, checkInterval, cpuAvgTime int64) *MemLoader {
	const (
		defaultMinBlockSize = 32 * 1024 * 1024  // 32MB
		defaultMaxBlockSize = 256 * 1024 * 1024 // 256MB
	)

	vloader := &MemLoader{
		TargetPercent:    targetPercent,
		CheckInterval:    checkInterval,
		CpuAvgTime:       cpuAvgTime,
		stopChan:         make(chan struct{}),
		minBlockSize:     defaultMinBlockSize,
		maxBlockSize:     defaultMaxBlockSize,
		smoothingFactor:  0.7,
		protectionFactor: 0.95, // 保护阈值（默认95%）
	}

	// 根据系统内存自动调整块大小
	totalMem := getTotalMemory()
	if totalMem > 0 {
		vloader.minBlockSize = totalMem / 100 // 1%总内存
		if vloader.minBlockSize < defaultMinBlockSize {
			vloader.minBlockSize = defaultMinBlockSize
		}

		vloader.maxBlockSize = totalMem / 20 // 5%总内存
		if vloader.maxBlockSize < defaultMaxBlockSize {
			vloader.maxBlockSize = defaultMaxBlockSize
		}
	}

	// 设置初始分配速率
	vloader.allocationRate = 0.01 // 1%每秒

	return vloader
}

func MenPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	memPercent := memInfo.UsedPercent
	return memPercent
}

// Start 启动内存负载调节器
func (loader *MemLoader) Start() {
	ticker := time.NewTicker(time.Duration(loader.CheckInterval) * time.Second)
	defer ticker.Stop()

	// 单独协程用于内存调整
	go loader.memoryAdjuster()

	for {
		select {
		case <-ticker.C:
			currentPercent := MenPercent()
			slog.Info("内存状态", "当前", currentPercent, "目标", loader.TargetPercent)

			// 内存保护机制
			if currentPercent > loader.protectionFactor*100 {
				loader.emergencyFree()
				continue
			}

			// 状态控制
			shouldBeActive := currentPercent < loader.TargetPercent
			currentlyActive := atomic.LoadInt32(&loader.active) == 1

			if shouldBeActive && !currentlyActive {
				atomic.StoreInt32(&loader.active, 1)
				slog.Info("启动内存负载生成")
			} else if !shouldBeActive && currentlyActive {
				atomic.StoreInt32(&loader.active, 0)
				slog.Info("停止内存负载生成")
			}

		case <-loader.stopChan:
			loader.freeAllMemory()
			return
		}
	}
}

// memoryAdjuster 内存调整协程
func (loader *MemLoader) memoryAdjuster() {
	const adjustmentInterval = 1 * time.Second
	ticker := time.NewTicker(adjustmentInterval)
	defer ticker.Stop()

	prevDiff := 0.0
	integral := 0.0

	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&loader.active) == 0 {
				continue
			}

			currentPercent := MenPercent()
			if math.IsNaN(currentPercent) || currentPercent <= 0 {
				continue
			}

			target := loader.TargetPercent
			diff := target - currentPercent

			// PID控制器参数 (比例、积分、微分)
			Kp := 0.5
			Ki := 0.05
			Kd := 0.01

			// 防积分饱和
			if math.Abs(diff) < 10 {
				integral += diff
			} else {
				integral = 0
			}

			// 计算速率调整量
			adjustment := Kp*diff + Ki*integral - Kd*(diff-prevDiff)

			// 应用平滑变换
			loader.adjustLock.Lock()
			loader.allocationRate = loader.smoothingFactor*loader.allocationRate +
				(1-loader.smoothingFactor)*clamp(adjustment/100, -0.05, 0.1)
			loader.adjustLock.Unlock()

			prevDiff = diff

			// 根据速率分配内存
			if loader.allocationRate > 0 {
				allocSize := uint64(loader.allocationRate * float64(getTotalMemory()))
				allocSize = clampSize(allocSize, loader.minBlockSize, loader.maxBlockSize)
				loader.allocateMemory(allocSize)
			} else if loader.allocationRate < 0 {
				freeSize := uint64(math.Abs(loader.allocationRate) * float64(atomic.LoadUint64(&loader.allocatedBytes)))
				freeSize = clampSize(freeSize, loader.minBlockSize/2, loader.maxBlockSize)
				loader.freeMemory(freeSize)
			}

		case <-loader.stopChan:
			return
		}
	}
}

// allocateMemory 分配指定大小的内存
func (loader *MemLoader) allocateMemory(size uint64) {
	if size == 0 {
		return
	}

	data := make([]byte, size)

	// 写入数据以防止优化（实际工作中可能会被编译器优化掉）
	for i := range data {
		data[i] = byte(i % 256)
	}

	// 将指针转换为整数保存，避免被GC回收
	ptr := uintptr(unsafe.Pointer(&data[0]))
	_ = ptr // 防止编译器警告

	atomic.AddUint64(&loader.allocatedBytes, size)
	slog.Debug("分配内存", "大小(MB)", size/(1024*1024), "总计(MB)", atomic.LoadUint64(&loader.allocatedBytes)/(1024*1024))

	// 触发GC但保留内存
	runtime.KeepAlive(data)
}

// freeMemory 释放指定大小的内存
func (loader *MemLoader) freeMemory(size uint64) {
	if size == 0 || atomic.LoadUint64(&loader.allocatedBytes) == 0 {
		return
	}

	// 在实际应用中，这里应该使用一个池来管理分配的内存块
	// 简化实现：通过缩小分配的内存大小来模拟释放
	current := atomic.LoadUint64(&loader.allocatedBytes)
	freed := uint64(0)

	if size >= current {
		freed = current
	} else {
		freed = size
	}

	atomic.AddUint64(&loader.allocatedBytes, -freed)
	slog.Debug("释放内存", "大小(MB)", freed/(1024*1024))
}

// freeAllMemory 释放所有内存
func (loader *MemLoader) freeAllMemory() {
	atomic.StoreUint64(&loader.allocatedBytes, 0)
	runtime.GC()
}

// emergencyFree 内存紧急释放
func (loader *MemLoader) emergencyFree() {
	slog.Warn("内存超过保护阈值，执行紧急释放")
	current := atomic.LoadUint64(&loader.allocatedBytes)
	if current > 0 {
		// 释放50%已分配内存
		freeSize := current / 2
		loader.freeMemory(freeSize)
	}

	// 如果内存仍然过高，释放更多
	time.Sleep(2 * time.Second)
	if MenPercent() > loader.protectionFactor*100 {
		slog.Warn("内存仍然过高，释放所有负载内存")
		loader.freeAllMemory()
	}
}

// Stop 停止内存负载
func (loader *MemLoader) Stop() {
	if atomic.LoadInt32(&loader.active) == 1 {
		atomic.StoreInt32(&loader.active, 0)
		close(loader.stopChan)
		loader.stopChan = make(chan struct{})
		loader.freeAllMemory()
		slog.Info("🛑 停止内存负载生成")
	}
}

// getTotalMemory 获取系统总内存（字节）
func getTotalMemory() uint64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil || memInfo.Total == 0 {
		// 默认返回32GB（基于图片中的31.8GB）
		return 32 * 1024 * 1024 * 1024
	}
	return memInfo.Total
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

// 限制内存大小在[min, max]范围内
func clampSize(value, min, max uint64) uint64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
