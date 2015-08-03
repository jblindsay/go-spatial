// Copyright 2015 the GoSpatial Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// licence that can be found in the LICENCE.txt file.

// File created by John Lindsay, March 2015 based on code originally found at
// https://github.com/oleiade/lane/blob/master/pqueue.go

package structures

import (
	"fmt"
	"sync"
)

// PQType represents a priority queue ordering kind (see MAXPQ and MINPQ)
type PQType int

const (
	MAXPQ PQType = iota
	MINPQ
)

type item struct {
	value    interface{}
	priority int
}

// PQueue is a heap priority queue data structure implementation.
// It can be whether max or min ordered and it is synchronized
// and is safe for concurrent operations.
type PQueue struct {
	sync.RWMutex
	items      []*item
	elemsCount int
	comparator func(int, int) bool
}

func newItem(value interface{}, priority int) *item {
	return &item{
		value:    value,
		priority: priority,
	}
}

func (i *item) String() string {
	return fmt.Sprintf("<item value:%s priority:%d>", i.value, i.priority)
}

// NewPQueue creates a new priority queue with the provided pqtype
// ordering type
func NewPQueue(pqType PQType) *PQueue {
	var cmp func(int, int) bool

	if pqType == MAXPQ {
		cmp = max
	} else {
		cmp = min
	}

	items := make([]*item, 1)
	items[0] = nil // Heap queue first element should always be nil

	return &PQueue{
		items:      items,
		elemsCount: 0,
		comparator: cmp,
	}
}

// Push the value item into the priority queue with provided priority.
func (pq *PQueue) Push(value interface{}, priority int) {
	item := newItem(value, priority)

	pq.Lock()
	//pq.items = append(pq.items, item)
	pq.items = appendItem(pq.items, item)
	pq.elemsCount += 1
	pq.swim(pq.elemsCount)
	pq.Unlock()
}

func appendItem(slice []*item, data *item) []*item {
	m := len(slice)
	n := m + 1
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]*item, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
		//println("Slice capacity:", cap(slice))
	}
	slice = slice[0:n]
	slice[m] = data
	//copy(slice[m:n], data)
	return slice
}

// Pop and returns the highest/lowest priority item (depending on whether
// you're using a MINPQ or MAXPQ) from the priority queue
func (pq *PQueue) Pop() interface{} {
	pq.Lock()
	defer pq.Unlock()

	if pq.elemsCount < 1 {
		return nil
	}

	var max *item = pq.items[1]

	pq.items[1], pq.items[pq.elemsCount] = pq.items[pq.elemsCount], pq.items[1]
	pq.items = pq.items[0:pq.elemsCount]
	pq.elemsCount -= 1
	pq.sink(1)

	return max.value
}

func (pq *PQueue) Len() int {
	pq.RLock()
	defer pq.RUnlock()
	return pq.elemsCount
}

func max(i, j int) bool {
	return i < j
}

func min(i, j int) bool {
	return i > j
}

func (pq *PQueue) swim(k int) {
	for k > 1 && (pq.items[k/2].priority < pq.items[k].priority) {
		pq.items[k/2], pq.items[k] = pq.items[k], pq.items[k/2]
		k = k / 2
	}
}

func (pq *PQueue) sink(k int) {
	var j int
	for 2*k <= pq.elemsCount {
		j = 2 * k

		if j < pq.elemsCount && (pq.items[j].priority < pq.items[j+1].priority) {
			j++
		}

		if !(pq.items[k].priority < pq.items[j].priority) {
			break
		}

		pq.items[k], pq.items[j] = pq.items[j], pq.items[k]
		k = j
	}
}
