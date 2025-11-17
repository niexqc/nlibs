package memloader

import (
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
)

type MemLoader struct {
	TargetPercent    float64       // 目标内存占用百分比
	CheckInterval    int64         // 检查间隔（秒）
	active           int32         // 负载激活状态
	stopChan         chan struct{} // 停止信号
	allocatedBytes   uint64        // 已分配字节数
	minBlockSize     uint64        // 最小内存块大小
	maxBlockSize     uint64        // 最大内存块大小
	adjustLock       sync.Mutex    // 内存调整锁
	smoothingFactor  float64       // 平滑因子
	allocationRate   float64       // 当前分配速率
	protectionFactor float64       // 内存保护因子
	blocksMutex      sync.Mutex
	blocks           [][]byte // 保持分配的内存块引用
}

func NewMemLoader(targetPercent float64, checkInterval int64) *MemLoader {
	const (
		defaultMinBlockSize = 32 * 1024 * 1024  // 32MB
		defaultMaxBlockSize = 128 * 1024 * 1024 // 128MB
	)

	vloader := &MemLoader{
		TargetPercent:    targetPercent,
		stopChan:         make(chan struct{}),
		CheckInterval:    checkInterval,
		minBlockSize:     defaultMinBlockSize,
		maxBlockSize:     defaultMaxBlockSize,
		smoothingFactor:  0.7,
		protectionFactor: 0.90, // 保护阈值（默认95%）
	}

	// 根据系统内存自动调整块大小
	totalMem := getTotalMemory()
	if totalMem > 0 {
		vloader.minBlockSize = totalMem / 200 // 0.5%总内存
		if vloader.minBlockSize < defaultMinBlockSize {
			vloader.minBlockSize = defaultMinBlockSize
		}

		vloader.maxBlockSize = totalMem / 40 // 2.5%总内存
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
			loaderMb := atomic.LoadUint64(&loader.allocatedBytes) / (1024 * 1024)
			slog.Info(fmt.Sprintf("内存状态,当前:%02f,目标:%02f,分配(Mb):%d,块数量:%d", currentPercent, loader.TargetPercent, loaderMb, len(loader.blocks)))

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
			runtime.GC()

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
			Kp := 1.2  // 增加比例系数
			Ki := 0.08 // 适度积分
			Kd := 0.02 // 适度微分

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
	data := make([]byte, size)
	// 初始化数据
	for i := range data {
		data[i] = byte(i % 256)
	}
	// 保持内存块引用，防止GC回收
	loader.blocksMutex.Lock()
	loader.blocks = append(loader.blocks, data)
	loader.blocksMutex.Unlock()

	atomic.AddUint64(&loader.allocatedBytes, size)
}

// freeMemory 释放指定大小的内存
func (loader *MemLoader) freeMemory(size uint64) {
	loader.blocksMutex.Lock()
	defer loader.blocksMutex.Unlock()
	if len(loader.blocks) == 0 {
		return
	}
	// 释放部分内存块（如释放前N个块直到满足size）
	freed := uint64(0)
	for freed < size && len(loader.blocks) > 0 {
		blockSize := uint64(len(loader.blocks[0]))
		if freed+blockSize > size && len(loader.blocks) > 1 {
			// 保留部分块，不全部释放
			break
		}
		// 移除块引用，允许GC回收
		loader.blocks = loader.blocks[1:]
		freed += blockSize
	}
	atomic.AddUint64(&loader.allocatedBytes, -freed)
	// 显式触发GC
	runtime.GC()
}

// freeAllMemory 释放所有内存
func (loader *MemLoader) freeAllMemory() {
	loader.blocksMutex.Lock()
	loader.blocks = nil // 直接置空，放弃所有内存块的引用
	loader.blocksMutex.Unlock()

	atomic.StoreUint64(&loader.allocatedBytes, 0)
	runtime.GC()
	debug.FreeOSMemory() // 同样建议在这里强制归还
}

// emergencyFree 内存紧急释放
func (loader *MemLoader) emergencyFree() {
	slog.Warn("内存超过保护阈值，执行紧急释放")

	// 1. 释放大部分已分配内存（例如75%）
	targetFreed := atomic.LoadUint64(&loader.allocatedBytes) * 3 / 4
	loader.freeMemory(targetFreed)

	// 2. 强制进行垃圾回收
	runtime.GC()

	// 3. 强制将内存归还给操作系统
	debug.FreeOSMemory() // 需要导入 "runtime/debug"

	slog.Info("紧急释放操作完成", "目标释放量(MB)", targetFreed/(1024*1024))

	// 4. 延长等待时间，让操作系统有足够时间处理（例如5-10秒）
	time.Sleep(8 * time.Second)

	// 5. 重新检查，使用更长的间隔或基于进程RSS判断
	currentPercent := MenPercent()
	if currentPercent > loader.protectionFactor*100 {
		slog.Warn("内存仍然过高，尝试完全释放")
		loader.freeAllMemory()
		runtime.GC()
		debug.FreeOSMemory()
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
