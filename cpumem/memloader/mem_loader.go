package memloader

import (
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
)

type MemLoader struct {
	TargetPercent    float64 // 目标内存占用百分比
	CheckInterval    int64   // 监控打印间隔（秒）
	active           int32
	stopChan         chan struct{}
	allocatedBytes   uint64 // 已分配字节数（原子）
	minBlockSize     uint64
	maxBlockSize     uint64
	adjustLock       sync.Mutex
	smoothingFactor  float64
	allocationRate   float64 // 每秒分配/释放的比例（正=分配，负=释放）
	protectionFactor float64 // 保护阈值（超过即强制释放）
	blocksMutex      sync.Mutex
	blocks           [][]byte // 保持已分配内存块的引用，防止GC回收
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
		protectionFactor: 0.90, // 系统内存使用率达到90%即触发保护
		active:           1,    // 初始即为激活状态，保证 Stop 的 CAS 能生效
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

	// 初始分配速率为0，由控制器按需调节
	vloader.allocationRate = 0.0

	return vloader
}

func MenPercent() float64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil || memInfo == nil {
		return 0
	}
	return memInfo.UsedPercent
}

// Start 启动内存负载调节器（阻塞）。
func (loader *MemLoader) Start() {
	ticker := time.NewTicker(time.Duration(loader.CheckInterval) * time.Second)
	defer ticker.Stop()

	// 单独的调节协程：始终运行（不再因 active 为0而停止），
	// 这样超过目标时也能持续释放内存，避免内存卡在峰值。
	go loader.memoryAdjuster()

	for {
		select {
		case <-ticker.C:
			currentPercent := MenPercent()
			loaderMb := atomic.LoadUint64(&loader.allocatedBytes) / (1024 * 1024)
			loader.blocksMutex.Lock()
			blockCount := len(loader.blocks)
			loader.blocksMutex.Unlock()
			slog.Info(fmt.Sprintf("内存状态,当前:%02f,目标:%02f,分配(Mb):%d,块数量:%d", currentPercent, loader.TargetPercent, loaderMb, blockCount))

			// 内存保护机制
			if currentPercent > loader.protectionFactor*100 {
				loader.emergencyFree()
			}
		case <-loader.stopChan:
			loader.freeAllMemory()
			return
		}
	}
}

// memoryAdjuster 内存调节协程：依据PID实时分配/释放，逼近目标并避免抖动。
func (loader *MemLoader) memoryAdjuster() {
	const adjustmentInterval = 2 * time.Second
	ticker := time.NewTicker(adjustmentInterval)
	defer ticker.Stop()

	prevDiff := 0.0
	integral := 0.0

	for {
		select {
		case <-ticker.C:
			// 仅当存在已分配内存时才触发GC，避免每周期强制GC造成抖动
			if atomic.LoadUint64(&loader.allocatedBytes) > 0 {
				runtime.GC()
			}

			currentPercent := MenPercent()
			if math.IsNaN(currentPercent) || currentPercent <= 0 {
				continue
			}

			// 自我保护：系统内存过高时直接释放，不再按速率调节
			if currentPercent > loader.protectionFactor*100 {
				continue
			}

			diff := loader.TargetPercent - currentPercent

			// PID
			Kp := 0.5
			Ki := 0.05
			Kd := 0.1

			if math.Abs(diff) < 8 {
				integral += diff
			} else {
				integral = 0
			}

			adjustment := Kp*diff + Ki*integral - Kd*(diff-prevDiff)

			// 应用平滑变换，将调节量映射为每秒分配/释放速率（百分比）
			loader.adjustLock.Lock()
			loader.allocationRate = loader.smoothingFactor*loader.allocationRate +
				(1-loader.smoothingFactor)*clamp(adjustment/100, -0.10, 0.10)
			loader.adjustLock.Unlock()

			prevDiff = diff

			if loader.allocationRate > 0 {
				allocSize := uint64(loader.allocationRate * float64(getTotalMemory()))
				allocSize = clampSize(allocSize, loader.minBlockSize/8, loader.maxBlockSize)
				loader.allocateMemory(allocSize)
			} else if loader.allocationRate < 0 {
				freeSize := uint64(math.Abs(loader.allocationRate) * float64(atomic.LoadUint64(&loader.allocatedBytes)))
				freeSize = clampSize(freeSize, loader.minBlockSize/8, loader.maxBlockSize)
				loader.freeMemory(freeSize)
			}

		case <-loader.stopChan:
			return
		}
	}
}

// allocateMemory 分配指定大小的内存并保持引用。
func (loader *MemLoader) allocateMemory(size uint64) {
	if size == 0 {
		return
	}
	data := make([]byte, size)
	// 只写入首尾字节以保证内存真实被占用（中段为0），避免逐字节填充带来的昂贵开销。
	data[0] = 1
	data[len(data)-1] = 1

	loader.blocksMutex.Lock()
	loader.blocks = append(loader.blocks, data)
	atomic.AddUint64(&loader.allocatedBytes, size)
	loader.blocksMutex.Unlock()
}

// freeMemory 释放指定大小的内存。
func (loader *MemLoader) freeMemory(size uint64) {
	loader.blocksMutex.Lock()
	defer loader.blocksMutex.Unlock()
	if len(loader.blocks) == 0 {
		return
	}
	freed := uint64(0)
	for freed < size && len(loader.blocks) > 0 {
		blockSize := uint64(len(loader.blocks[0]))
		if freed+blockSize > size && len(loader.blocks) > 1 {
			break
		}
		loader.blocks = loader.blocks[1:]
		freed += blockSize
	}
	atomic.AddUint64(&loader.allocatedBytes, -freed)
	if freed > 0 {
		runtime.GC()
	}
}

// freeAllMemory 释放所有内存。
func (loader *MemLoader) freeAllMemory() {
	loader.blocksMutex.Lock()
	loader.blocks = nil
	loader.blocksMutex.Unlock()
	atomic.StoreUint64(&loader.allocatedBytes, 0)
	runtime.GC()
}

// emergencyFree 系统内存超过保护阈值时紧急释放。
func (loader *MemLoader) emergencyFree() {
	slog.Warn("内存超过保护阈值，执行紧急释放")
	targetFreed := atomic.LoadUint64(&loader.allocatedBytes) * 3 / 4
	loader.freeMemory(targetFreed)

	currentPercent := MenPercent()
	if currentPercent > loader.protectionFactor*100 {
		slog.Warn("内存仍然过高，尝试完全释放")
		loader.freeAllMemory()
	}
}

// Stop 停止内存负载。
func (loader *MemLoader) Stop() {
	if atomic.CompareAndSwapInt32(&loader.active, 1, 0) {
		close(loader.stopChan)
		loader.stopChan = make(chan struct{})
		loader.freeAllMemory()
		slog.Info("🛑 停止内存负载生成")
	}
}

// getTotalMemory 获取系统总内存（字节）。
func getTotalMemory() uint64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil || memInfo.Total == 0 {
		return 32 * 1024 * 1024 * 1024
	}
	return memInfo.Total
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampSize(value, min, max uint64) uint64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
