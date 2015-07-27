package main

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

type TaskHandler interface {
	TaskName() string
	HandleTask(task interface{}) bool
	TaskToString(task interface{}) string
}

type Engine struct {
	dispatchThreads []*DispatcherThread
	waitGroup       sync.WaitGroup
}

func NewEngine() *Engine {
	var wg sync.WaitGroup
	return &Engine{dispatchThreads: make([]*DispatcherThread, 0), waitGroup: wg}
}

func (this *Engine) AddTask(taskHanlder TaskHandler, mainWorkerCnt int, retryWorkerThread int) {
	this.dispatchThreads = append(this.dispatchThreads, NewDispatcherThread(taskHanlder, mainWorkerCnt, retryWorkerThread))
}

func (this *Engine) Start() {
	seelog.Info("[Engine Start]")
	defer this.waitGroup.Wait()
	for _, dispatchThread := range this.dispatchThreads {
		this.waitGroup.Add(1)
		go dispatchThread.Run(this.waitGroup)
	}
}

type DispatcherThread struct {
	TaskHanlder       TaskHandler
	redisConn         redis.Conn
	mainChanGroup     *ChanGroup
	retryChanGroup    *ChanGroup
	mainWorkerCnt     int
	retryWorkerCnt    int
	mainWorkerThread  []*WorkerThread
	retryWorkerThread []*WorkerThread
}

func NewDispatcherThread(taskHanlder TaskHandler, mainWorkerCnt int, retryWorkerCnt int) *DispatcherThread {
	dispatcherThread := &DispatcherThread{TaskHanlder: taskHanlder,
		redisConn:      RedisPool.Get(),
		mainChanGroup:  NewChanGroup(mainWorkerCnt),
		retryChanGroup: NewChanGroup(retryWorkerCnt),
		mainWorkerCnt:  mainWorkerCnt,
		retryWorkerCnt: retryWorkerCnt}
	dispatcherThread.retryWorkerThread = make([]*WorkerThread, retryWorkerCnt)
	for i := 0; i < retryWorkerCnt; i++ {
		dispatcherThread.retryWorkerThread[i] =
			NewWorkerThread(
				fmt.Sprintf("%v Retry Worker<%v>", dispatcherThread.TaskHanlder.TaskName(), i),
				taskHanlder)
	}
	dispatcherThread.mainWorkerThread = make([]*WorkerThread, mainWorkerCnt)
	for i := 0; i < mainWorkerCnt; i++ {
		dispatcherThread.mainWorkerThread[i] =
			NewWorkerThread(
				fmt.Sprintf("%v Main Worker<%v>", dispatcherThread.TaskHanlder.TaskName(), i),
				taskHanlder)
	}
	return dispatcherThread
}

func (this *DispatcherThread) Run(wg sync.WaitGroup) {
	defer wg.Done()
	seelog.Infof("[%v Dispatcher Thread Start]", this.TaskHanlder.TaskName())
	wg.Add(1)
	for i := 0; i < this.retryWorkerCnt; i++ {
		wg.Add(1)
		go this.retryWorkerThread[i].Run(this.retryChanGroup.NextChan(), nil, wg)
	}
	for i := 0; i < this.mainWorkerCnt; i++ {
		wg.Add(1)
		go this.mainWorkerThread[i].Run(this.mainChanGroup.NextChan(), this.retryChanGroup, wg)
	}

	for {
		emailTasks := FetchEmailTasksFromRedis()
		for _, task := range emailTasks {
			seelog.Debugf("[Prepare To Start %v Task] [Task : %v]", this.TaskHanlder.TaskName(), this.TaskHanlder.TaskToString(task))
			this.mainChanGroup.NextChan() <- task
		}
		time.Sleep(5 * time.Second)
	}
}

type WorkerThread struct {
	WorkerName  string
	TaskHanlder TaskHandler
}

func NewWorkerThread(name string, taskHanlder TaskHandler) *WorkerThread {
	return &WorkerThread{WorkerName: name, TaskHanlder: taskHanlder}
}

func (this *WorkerThread) Run(myChan chan interface{}, retryChanGroup *ChanGroup, wg sync.WaitGroup) {
	seelog.Infof("[%v Start]", this.WorkerName)
	defer wg.Done()
	for {
		task := <-myChan
		if !this.TaskHanlder.HandleTask(task) {
			if retryChanGroup != nil {
				seelog.Debugf("[Prepare To Retry %v Task] [Task : %v]", this.TaskHanlder.TaskName(), this.TaskHanlder.TaskToString(task))
				retryChanGroup.NextChan() <- task
			} else {
				seelog.Debugf("[Abandon %v Task] [Task : %v]", this.TaskHanlder.TaskName(), this.TaskHanlder.TaskToString(task))
			}
		}
	}
}
