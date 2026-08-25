package ntools

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DirScan 递归扫描目录（基于 filepath.WalkDir）。
// 需要并发扫描时使用 DirScanConcurrent（回调语义与本函数一致）。
func DirScan(scanRootPath string, resp func(scanRootPath, relativePath, absPath string, entry fs.DirEntry)) {
	fileInfo, err := os.Stat(scanRootPath)
	if nil != err {
		panic(fmt.Sprintf("获取文件信息:【%s】,异常:%s", scanRootPath, err.Error()))
	}
	if !fileInfo.IsDir() {
		panic(fmt.Sprintf("非目录【%s】", scanRootPath))
	}
	scanRootPathStr, _ := filepath.Abs(scanRootPath)

	filepath.WalkDir(scanRootPathStr, func(absPath string, d fs.DirEntry, err error) error {
		relativePath := strings.ReplaceAll(absPath, scanRootPathStr, "")
		if len(relativePath) > 0 {
			relativePath = relativePath[1:]
		}
		resp(scanRootPathStr, relativePath, absPath, d)
		return err
	})
}

// DirScanConcurrent 并发递归扫描文件夹。
// 与 DirScan 的语义一致（回调参数相同），但采用"目录队列 + 协程池"模型，
// 让目录遍历在多个协程间并行，从而在机械硬盘上也能接近磁盘 stat 的速度上限。
// workerNum 建议 4~8（机械盘上过高会造成磁头抖动，反而更慢）。
// 内部使用 mutex + condition 驱动的无界目录队列，保证无死锁、无忙等。
func DirScanConcurrent(scanRootPath string, workerNum int, resp func(scanRootPath, relativePath, absPath string, entry fs.DirEntry)) {
	fileInfo, err := os.Stat(scanRootPath)
	if nil != err {
		panic(fmt.Sprintf("获取文件信息:【%s】,异常:%s", scanRootPath, err.Error()))
	}
	if !fileInfo.IsDir() {
		panic(fmt.Sprintf("非目录【%s】", scanRootPath))
	}
	scanRootPathStr, _ := filepath.Abs(scanRootPath)

	if workerNum < 1 {
		workerNum = 1
	}

	var (
		mu      sync.Mutex
		cond    = sync.NewCond(&mu)
		pending []string // 待处理目录队列
		active  int      // 在途目录数（已入队 + 正在处理）
	)
	var wg sync.WaitGroup

	enqueue := func(dir string) {
		mu.Lock()
		pending = append(pending, dir)
		active++
		mu.Unlock()
		cond.Signal()
	}

	enqueue(scanRootPathStr)

	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mu.Lock()
				for len(pending) == 0 {
					if active == 0 {
						mu.Unlock()
						return
					}
					cond.Wait()
				}
				dir := pending[0]
				pending = pending[1:]
				mu.Unlock()

				entries, err := os.ReadDir(dir)
				if nil != err {
					panic(fmt.Sprintf("读取子目录:【%s】,异常:%s", dir, err.Error()))
				}
				for _, item := range entries {
					fullPath := filepath.Join(dir, item.Name())
					relativePath := relPath(scanRootPathStr, fullPath)
					if item.IsDir() {
						resp(scanRootPathStr, relativePath, fullPath, item)
						enqueue(fullPath)
					} else {
						resp(scanRootPathStr, relativePath, fullPath, item)
					}
				}

				mu.Lock()
				active--
				if active == 0 && len(pending) == 0 {
					cond.Broadcast()
				}
				mu.Unlock()
				cond.Signal()
			}
		}()
	}

	wg.Wait()
}

// relPath 计算相对路径，与 DirScan 中 strings.ReplaceAll 的语义保持一致。
func relPath(scanRootPathStr, absPath string) string {
	relativePath := strings.ReplaceAll(absPath, scanRootPathStr, "")
	if len(relativePath) > 0 {
		relativePath = relativePath[1:]
	}
	return relativePath
}
