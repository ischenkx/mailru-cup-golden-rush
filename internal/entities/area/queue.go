package area

import (
	"container/heap"
	"sync"
)

type sortableArea []Area

func (s *sortableArea) Swap(i, j int) {
	slc := *s
	t := slc[i]
	slc[i] = slc[j]
	slc[j] = t
}

func (s *sortableArea) Less(i, j int) bool {
	slc := *s
	return slc[j].Treasures < slc[i].Treasures
}

func (s *sortableArea) Len() int {
	slc := *s
	return len(slc)
}

func (s *sortableArea) Push(data interface{}) {
	*s = append(*s, data.(Area))
}

func (s *sortableArea) Pop() interface{} {
	slc := *s
	n := len(slc)
	el := slc[n-1]
	*s = slc[:n-1]
	return el
}

type Queue struct {
	sortedAreas []Area
	blockingBufferSize int
	counter int64
	popCond, pushCond *sync.Cond
	mu sync.Mutex
}

func (q *Queue) sortable() *sortableArea {
	return (*sortableArea)(&q.sortedAreas)
}

func (q *Queue) PushWithoutBlocking(area Area) {
	q.mu.Lock()
	defer q.mu.Unlock()
	prevLen := len(q.sortedAreas)
	heap.Push(q.sortable(), area)
	if prevLen == 0 && len(q.sortedAreas) > 0 {
		q.popCond.Signal()
	}
}

func (q *Queue) Push(area Area) {
	q.mu.Lock()
	if len(q.sortedAreas) > q.blockingBufferSize {
		q.pushCond.Wait()
	}
	prevLen := len(q.sortedAreas)
	heap.Push(q.sortable(), area)
	if prevLen == 0 {
		q.popCond.Signal()
	}
	q.mu.Unlock()
}

func (q *Queue) Pop() Area {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.sortedAreas) == 0 {
		q.popCond.Wait()
	}
	srt := q.sortable()
	prevLen := srt.Len()
	area := heap.Pop(srt).(Area)
	if prevLen >= q.blockingBufferSize && len(q.sortedAreas) < q.blockingBufferSize {
		q.pushCond.Signal()
	}
	return area
}

func (q *Queue) Size() int {
	q.mu.Lock()
	l := len(q.sortedAreas)
	q.mu.Unlock()
	return l
}

func NewQueue(bsize int) *Queue {
	q := &Queue{
		sortedAreas:        nil,
		blockingBufferSize: bsize,
		counter:            0,
		mu:                 sync.Mutex{},
	}
	q.popCond = sync.NewCond(&q.mu)
	q.pushCond = sync.NewCond(&q.mu)
	return q
}