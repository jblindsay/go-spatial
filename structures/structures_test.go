package structures

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

var testKD = false
var testPQ = true

func TestKDTree(t *testing.T) {
	// Make a K-D tree of random points.

	if testKD {
		const N = 1000
		nodes := make([]*T, N)
		for i := range nodes {
			nodes[i] = new(T)
			for j := range nodes[i].Point {
				nodes[i].Point[j] = rand.Float64() * 10000000
			}
			nodes[i].Data = fmt.Sprintf("Point %v", i)
		}
		tree := New(nodes)

		nodes = tree.InRange(nodes[500].Point, 0.25, make([]*T, 0, N))
		fmt.Println(nodes)

		// Reuse the nodes slice from the previous call.
		//nodes = tree.InRange(Point{0, 0}, 0.5, nodes[:0])
		//fmt.Println(nodes)
	} else {
		t.SkipNow()
	}
}

func TestPQTree(t *testing.T) {
	if testPQ {
		letters := [10]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		q := NewPQueue(MAXPQ)
		//q := NewPriorityQueue()
		val := ""
		//Pushes
		start1 := time.Now()

		q.Push("Jim", 1)
		q.Push("Bob", 3)
		q.Push("Mary", 4)
		q.Push("Larry", 5)
		q.Push("Sally", 2)

		j := 0
		for j < 1500000 {
			name := ""
			for i := 0; i < 6; i++ {
				name += letters[int(rand.Float32()*10)]
			}
			q.Push(name, rand.Int())
			j++
		}

		elapsed := time.Since(start1)

		// Pop time
		start2 := time.Now()
		for q.Len() > 0 {
			//(q.Poll()).(string)
			val = (q.Pop()).(string)
			if strings.HasPrefix(val, "aaaaa") {
				println(val)
			}
			if q.Len() < 6 {
				println(val)
			}
		}

		value := fmt.Sprintf("Elapsed Push time: %s", elapsed)
		println(value)

		elapsed = time.Since(start2)
		value = fmt.Sprintf("Elapsed Pop time: %s", elapsed)
		println(value)
	} else {
		t.SkipNow()
	}
}
