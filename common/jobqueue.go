package common

import (
	"sync"
)

type JobQueue struct {
	mutex *sync.Mutex
	wait  *Queue
	work  *Queue
}

type Response struct {
	Job interface{}
	Id  int64
}

func NewJobQueue() *JobQueue {
	j := &JobQueue{
		wait:  NewQueue(),
		work:  NewQueue(),
		mutex: &sync.Mutex{},
	}

	return j
}

func (j *JobQueue) Push(elem interface{}, poptimes int64) interface{} {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if j.work.Exist(elem) {
		return false
	}

	if j.wait.Exist(elem) {
		return false
	} else {
		return j.wait.Append(0, poptimes, elem)
	}
}

func (j *JobQueue) Pop() (elem interface{}) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if j.wait.Empty() {
		if j.work.Empty() {
			return nil
		} else {
			id, poptimes, item := j.work.Pop()

			if id != 0 && poptimes != 1 {
				poptimes = poptimes - 1
				id = j.work.Append(id, poptimes, item)
			}

			return &Response{
				Job: item,
				Id:  id,
			}
		}
	} else {
		id, poptimes, item := j.wait.Pop()

		if id != 0 && poptimes != 1 && poptimes != -1 {
			poptimes = poptimes - 1
			id = j.work.Append(id, poptimes, item)
		}

		return &Response{
			Job: item,
			Id:  id,
		}
	}
}

func (j *JobQueue) Remove(id int64) interface{} {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	if j.work.Remove(id) {
		return true
	} else {
		if j.wait.Remove(id) {
			return true
		} else {
			return false
		}
	}
}

func (j *JobQueue) Clean() {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.work.Clean()
	j.wait.Clean()
}

func (j *JobQueue) Length() int {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	return j.work.Length() + j.wait.Length()
}
