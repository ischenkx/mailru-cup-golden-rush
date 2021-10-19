package main

import (
	"fmt"
	area2 "github.com/RomanIschenko/golden-rush-mailru/internal/entities/area"
	"github.com/RomanIschenko/golden-rush-mailru/internal/entities/coin"
	"math/rand"
	"time"
)

type Map struct {
	arr [][]int
}

func (m *Map) setup(w, h int) {
	m.arr = make([][]int, h)

	for i := 0; i < h; i++ {
		m.arr[i] = make([]int, w)
	}
}

func (m *Map) fillSerial() {
	c := 0
	for _, row := range m.arr {
		for i := range row {
			row[i] = c
			c++
		}
	}
}

func (m *Map) getSum(x, y, w, h int) int {
	s := 0
	for i := y; i < y+h; i++ {
		for j := x; j < x + w; j++ {
			s+=m.arr[j][i]
		}
	}
	return s
}

func gen(ch chan area2.Area) {
	areas := area2.GenerateAreas(1, 1, 3500, 3500, 5, 5)

	rand.Shuffle(len(areas), func(i, j int) {
		t := areas[i]
		areas[i] = areas[j]
		areas[j] = t
	})

	for _, a := range areas {
		ch <- a
	}
}

func main() {

	w := coin.NewManager()

	w.Add([]uint32{1,2,3,4,5,6,6,7}...)

	fmt.Println(w.PopNOrAll(3), w.PopNOrAll(3), w.PopNOrAll(7))
	time.Sleep(time.Second * 20)
}