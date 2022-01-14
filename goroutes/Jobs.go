package goroutes

import (
	"SecShell/config"
	"github.com/xtaci/kcp-go/v5"
	"sync"
)

var jobID = 0

// 定义作业所需要结构体

var Jobs = &jobs{
	active: map[int]*Job{},
	mutex:  &sync.RWMutex{},
}

type Job struct {
	ID           int    // 作业id
	Name         string // 作业名称
	Description  string // 作业描述
	PersistentID string // 当前作业id
	Conn         *kcp.UDPSession
	Debug        bool
	Info         config.SystemInfo
}

type jobs struct {
	active map[int]*Job  // 活跃的会话
	mutex  *sync.RWMutex // 读写锁
}

// 获取所有的活跃作业

func (j *jobs) All() []*Job {
	j.mutex.RLock()         // 加读锁
	defer j.mutex.RUnlock() // 解读锁
	var all []*Job
	// 遍历所有的活跃作业，并将作业添加到all切片当中
	for _, job := range j.active {
		all = append(all, job)
	}
	return all
}

// 添加一个活跃作业

func (j *jobs) Add(job *Job) {
	j.mutex.Lock()         // 加写锁
	defer j.mutex.Unlock() // 解写锁
	j.active[job.ID] = job // 添加一个作业到active中
	EventBroker.Publish(Event{
		Job:       job,
		EventType: "start-job",
	})
}

func (j *jobs) Remove(job *Job) {
	j.mutex.Lock()         // 加写锁
	defer j.mutex.Unlock() // 解写锁
	delete(j.active, job.ID)
	EventBroker.Publish(Event{
		Job:       job,
		EventType: "stop-job",
	})
}

func (j *jobs) Get(jobID int) *Job {
	if jobID <= 0 {
		return nil
	}
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	return j.active[jobID]
}

func NextJobID() int {
	newID := jobID + 1
	jobID++
	return newID
}
