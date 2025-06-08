package nlpworker

import (
	"log/slog"
	"sync"
	"time"

	"github.com/niexqc/nlibs/ntools"
	"github.com/panjf2000/ants/v2"
)

type NlpWorkGroup[T any] struct {
	workGroupName       string
	started             bool
	mutx                *sync.Mutex
	maxGrpNum           int
	maxGrpWorkDoTaskNum int
	nextTimeout         int64 //空数据时候，下次执行的时间,秒
	waitNotifyer        chan struct{}
	workPool            *ants.Pool
	findGrpFun          func(maxResult int) []string               //获取分组
	nextTaskFun         func(grpName string, childRunCount int) *T // 获取分组任务
	taskDoFun           func(task *T, childRunCount int)           //执行任务
}

func NewNlpWorkGroup[T any](workGroupName string, maxGrpNum, maxGrpWorkDoTaskNum int, nextTimeout int64,
	findGrpFun func(maxResult int) []string, nextTaskFun func(grpName string, childRunCount int) *T, taskDoFun func(task *T, childRunCount int),
) *NlpWorkGroup[T] {
	// 不需要关心ants.NewPool返回的err
	// 必须使用ants.WithNonblocking(false)的阻塞提交模式,保证任务的顺序执行
	antsPool, _ := ants.NewPool(maxGrpNum, ants.WithNonblocking(false))
	nlp := &NlpWorkGroup[T]{
		workGroupName:       workGroupName,
		maxGrpNum:           maxGrpNum,
		maxGrpWorkDoTaskNum: maxGrpWorkDoTaskNum,
		nextTimeout:         nextTimeout,
		started:             false,
		mutx:                &sync.Mutex{},
		waitNotifyer:        make(chan struct{}, 1),
		workPool:            antsPool,
		findGrpFun:          findGrpFun,
		nextTaskFun:         nextTaskFun,
		taskDoFun:           taskDoFun,
	}
	return nlp
}

func (nlp *NlpWorkGroup[T]) Start() {
	nlp.mutx.Lock()
	defer nlp.mutx.Unlock()
	if nlp.started {
		return
	}
	nlp.started = true
	go nlp.printNlpWorkGroupStatus()

	go func() {
		slog.Debug("NlpGroup start working", "workGroupName", nlp.workGroupName)
		for {
			nlp.nlpLoopWork()
		}
	}()

}

func (nlp *NlpWorkGroup[T]) printNlpWorkGroupStatus() {
	//实现30秒打印一次,		nlp.workPool转状态
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	ntools.SlogSetTraceId("printNlpWorkGroupStatus")
	for range ticker.C {
		slog.Info("NlpWorkGroup Status", " workGroupName", nlp.workGroupName, " maxGrpNum", nlp.maxGrpNum, " maxGrpWorkDoTaskNum", nlp.maxGrpWorkDoTaskNum, " RuningGrpWorker", nlp.workPool.Running())
	}
}

func (nlp *NlpWorkGroup[T]) nlpLoopWork() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("NlpWorkGroup-nlpLoopWork 发生异常", "err", r)
		}
	}()
	ntools.SlogSetTraceId("nlpLoopWork")
	// 在findGrpFun内部不存在竞态条件
	grpNames := nlp.findGrpFun(nlp.maxGrpNum)
	if len(grpNames) == 0 {
		timer := time.NewTimer(time.Duration(nlp.nextTimeout) * time.Second)
		defer timer.Stop()
		// 如果超时或者收到唤醒通知，再次进入循环
		select {
		case <-timer.C: // Timer超时
		case <-nlp.waitNotifyer: // 唤醒通知
			if !timer.Stop() {
				<-timer.C // 排空通道
			}
		}
	} else {
		waitGrp := &sync.WaitGroup{}
		waitGrp.Add(len(grpNames))
		for _, grpName := range grpNames {
			grpName := grpName // 闭包捕获
			nlp.workPool.Submit(func() {
				defer waitGrp.Done()
				newNlpWorker(grpName, nlp).Start()
			})
		}
		waitGrp.Wait()
	}
}

// 如果循环未在工作中，通知循环可以开始工作： 无阻塞
func (nlp *NlpWorkGroup[T]) NotifyNlpGroup() {
	select {
	case nlp.waitNotifyer <- struct{}{}:
	default: // 避免阻塞
	}
}

type nlpWorker[T any] struct {
	grpWrkName string
	workGroup  *NlpWorkGroup[T]
}

func newNlpWorker[T any](grpWrkName string, workGroup *NlpWorkGroup[T]) *nlpWorker[T] {
	nlp := &nlpWorker[T]{
		grpWrkName: grpWrkName,
		workGroup:  workGroup,
	}
	return nlp
}

func (wrk *nlpWorker[T]) Start() {
	ntools.SlogSetTraceId("GRPN_" + wrk.grpWrkName)
	slog.Debug("NewNlpWorker start working", "grpWrkName", wrk.grpWrkName)
	taskRunCount := 0
	for {
		if taskRunCount >= wrk.workGroup.maxGrpWorkDoTaskNum {
			slog.Debug("NewNlpWorker dotask times over,giveup wait next run ", "maxGrpWorkDoTaskNum", wrk.workGroup.maxGrpWorkDoTaskNum)
			break
		}
		ntools.SlogSetTraceId("GRPN_" + wrk.grpWrkName)
		// 这里需要读取任务直到所有任务完成--
		task := wrk.workGroup.nextTaskFun(wrk.grpWrkName, taskRunCount)
		if task != nil {
			wrk.workGroup.taskDoFun(task, taskRunCount)
		} else {
			break
		}
		taskRunCount++
	}
	slog.Debug("NewNlpWorker stop work", "grpWrkName", wrk.grpWrkName)
}
