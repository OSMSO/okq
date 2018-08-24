package common

import (
	"github.com/alex023/clock"
	"sync"
	"time"
	"math/rand"
	"strconv"
	"container/list"
)

type Clock struct {
	Content interface{} `json:"cotent"`
	Job     clock.Job   `json:"job"`
}

type JsonClock struct {
	Content interface{} `json:"content"`
	Id      string      `json:"id"`
}

type ClocksStore struct {
	sync.Mutex
	clocks map[string]Clock
	list   *list.List
}

func NewClocksStore() *ClocksStore {
	return &ClocksStore{
		clocks: make(map[string]Clock),
		list:   list.New(),
	}
}

func (g *ClocksStore) newId() string {
	for {
		id := rand.Int63()
		_, ok := g.clocks[strconv.FormatInt(id, 10)]
		if id != 0 && !ok {
			return strconv.FormatInt(id, 10)
		}
	}
}

func (g *ClocksStore) AddClock(id string, content interface{}, interval time.Duration, actionMax uint64, jobFunc func()) (added bool, err error, tid string) {
	g.Lock()
	defer g.Unlock()

	if len(id) == 0 {
		id = g.newId()
	}
	tid = id
	item, founded := g.clocks[id]
	if founded {
		added = false
	} else {
		job, _ := clock.NewClock().AddJobRepeat(interval, actionMax, jobFunc)

		item = Clock{
			Content: content,
			Job:     job,
		}

		g.list.PushBack(id)
		g.clocks[id] = item
		added = true
	}
	return
}

func (g *ClocksStore) GetClock(id string) bool {
	g.Lock()
	defer g.Unlock()

	_, founded := g.clocks[id]
	return founded
}

func (g *ClocksStore) GetClocks(begin, end int) interface{} {
	g.Lock()
	defer g.Unlock()

	clocks := []JsonClock{}

	if end > len(g.clocks) {
		end = len(g.clocks)
	}

	index := 0
	for e := g.list.Front(); e != nil; e = e.Next() {
		if index > end {
			break
		} else if index >= begin {
			clocks = append(clocks, JsonClock{g.clocks[e.Value.(string)].Content, e.Value.(string)})
		}

		index += 1
	}

	return clocks
}

func (g *ClocksStore) GetClockNum() int {
	g.Lock()
	defer g.Unlock()
	return len(g.clocks)
}

func (g *ClocksStore) RemoveClock(id string) {
	g.Lock()
	defer g.Unlock()

	if work, founded := g.clocks[id]; founded {
		for e := g.list.Front(); e != nil; e = e.Next() {
			if e.Value == id {
				g.list.Remove(e)
			}
		}
		delete(g.clocks, id)
		work.Job.Cancel()
	}
}

func (g *ClocksStore) CleanInsectGod() {
	g.Lock()
	defer g.Unlock()

	g.list = list.New()
	for k, _ := range g.clocks {
		if work, founded := g.clocks[k]; founded {
			delete(g.clocks, k)
			work.Job.Cancel()
		}
	}
}
