package common

import (
	"math/rand"
	"sync"
	"container/list"
)

// popTimes	中存放了已经Pop的次数
// items	中存放着任务id和任务内容
// ids 		正好是反过来存放着
// count 	是当前的数量
// mutex	是锁
type Queue struct {
	items    map[int64]interface{}
	popTimes map[int64]int64
	ids      map[interface{}]int64
	buf      *list.List
	count    int
	mutex    *sync.Mutex
}

// 添加一个新的Queue,每个都会生成一个同步锁
// 然后这个queue中,会存放
func NewQueue() *Queue {
	q := &Queue{
		items:    make(map[int64]interface{}),
		popTimes: make(map[int64]int64),
		ids:      make(map[interface{}]int64),
		buf:      list.New(),
		mutex:    &sync.Mutex{},
	}

	return q
}

// Removes all elements from queue
func (q *Queue) Clean() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.items = make(map[int64]interface{})
	q.ids = make(map[interface{}]int64)
	q.popTimes = make(map[int64]int64)
	q.buf = list.New()
	q.count = 0
}

// Returns the number of elements in queue
func (q *Queue) Length() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.items)
}

// 为这个queue添加一个新的元素
func (q *Queue) Append(id int64, poptimes int64, elem interface{}, ) int64 {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if id == 0 {
		id = q.newId()
	}
	q.items[id] = elem
	q.ids[elem] = id
	q.popTimes[id] = poptimes
	q.buf.PushBack(id)

	q.count++
	return id
}

// Adds one element at the front of queue
func (q *Queue) Prepend(elem interface{}, poptimes int64) int64 {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	id := q.newId()
	q.items[id] = elem
	q.ids[elem] = id
	q.popTimes[id] = poptimes
	q.buf.PushFront(id)
	q.count++

	return id
}

// Previews element at the front of queue
// 查看这个queue中最前面的元素
func (q *Queue) Front() interface{} {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	id := q.buf.Front()
	if id != nil {
		return q.items[id.Value.(int64)]
	}
	return nil
}

// Previews element at the back of queue
// 查看这个queue中最后面的元素
func (q *Queue) Back() interface{} {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	id := q.buf.Back()

	if id != nil {
		return q.items[id.Value.(int64)]
	}
	return nil
}

// get pop
func (q *Queue) pop() int64 {
	if q.count <= 0 {
		return 0
	}

	id := q.buf.Front()

	q.count--
	q.buf.Remove(id)

	return id.Value.(int64)
}

// Pop removes and returns the element from the front of the queue.
// If the queue is empty, it will block
func (q *Queue) Pop() (int64, int64, interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	id := q.pop()

	if id == 0 {
		return 0, 0, nil
	}

	item, ok := q.items[id]
	poptimes, Tok := q.popTimes[id]

	if ok && Tok {
		delete(q.ids, item)
		delete(q.items, id)
		delete(q.popTimes, id)
		return id, poptimes, item
	} else {
		return 0, 0, nil
	}
}

// Removes one element from the queue
// 删除一个元素从队列中
// 通过id的形式
func (q *Queue) Remove(id int64) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	item, ok := q.items[id]
	if !ok {
		return false
	}
	delete(q.items, id)
	delete(q.ids, item)
	delete(q.popTimes, id)

	for e := q.buf.Front(); e != nil; e = e.Next() {
		if e.Value == id {
			q.buf.Remove(e)
		}
	}

	q.count--
	return true
}

// 查询一个元素是否在队列中
func (q *Queue) Exist(elem interface{}) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	_, ok := q.ids[elem]
	if !ok {
		return false
	}
	return true
}

// 生成一个新的id,这个id不存在已有的queue中
func (q *Queue) newId() int64 {
	for {
		id := rand.Int63()
		_, ok := q.items[id]
		if id != 0 && !ok {
			return id
		}
	}
}

// 查询这个queue是否为空
func (q *Queue) Empty() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.count <= 0 {
		return true
	}
	return false
}
