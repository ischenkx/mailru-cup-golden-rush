package license

import (
	"container/heap"
	"github.com/RomanIschenko/golden-rush-mailru/internal/config"
	"sync"
)

type intRevSlice []int64

func (p *intRevSlice) Len() int           { return len(*p) }
func (p *intRevSlice) Less(i, j int) bool { return (*p)[i] < (*p)[j] }
func (p *intRevSlice) Swap(i, j int)      { (*p)[i], (*p)[j] = (*p)[j], (*p)[i] }
func (p *intRevSlice) Push(i interface{})      { *p = append(*p, i.(int64))}
func (p *intRevSlice) Pop() interface{} {
	l := len(*p)
	el := (*p)[l-1]
	*p = (*p)[:l-1]
	return el
}

type Price struct {
	CoinsAmount int64
	RealAmount int64
	Digs int64
	Failed bool
	experimental bool
}

func (p *Price) Experimental() bool {
	return p.experimental
}

type sortedPrices []Price

func (s *sortedPrices) Swap(i, j int) {
	slc := *s
	t := slc[i]
	slc[i] = slc[j]
	slc[j] = t
}

func (s *sortedPrices) Less(i, j int) bool {
	slc := *s
	return slc[j].Digs > slc[i].Digs
}

func (s *sortedPrices) Len() int {
	slc := *s
	return len(slc)
}

func (s *sortedPrices) Push(data interface{}) {
	*s = append(*s, data.(Price))
}

func (s *sortedPrices) Pop() interface{} {
	slc := *s
	n := len(slc)
	el := slc[n-1]
	*s = slc[:n-1]
	return el
}

type PriceList struct {
	mu sync.RWMutex
	results map[int64]int64
	experimentalAmounts []int64
	sortedPrices *sortedPrices
	g float64
	k int64

	calculatorCounter int64

	counter int64
}

func (p *PriceList) sortable() *intRevSlice {
	return (*intRevSlice)(&p.experimentalAmounts)
}

func NewPriceList(config config.Config) *PriceList {
	arr := make([]int64, 0, config.App.License.PriceList.Experiments)
	for i := 1; i < config.App.License.PriceList.Experiments; i+=3 {
		heap.Push((*intRevSlice)(&arr), int64(i))
	}
	p := &PriceList{
		mu:                  sync.RWMutex{},
		results:             map[int64]int64{},
		experimentalAmounts: arr,
		sortedPrices:        (*sortedPrices)(&[]Price{}),
		g:                   config.App.License.PriceList.G,
		k:                   int64(config.App.License.PriceList.K),
		counter:             0,
	}
	return p
}

func (p *PriceList) Commit(price Price) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !price.experimental {
		return
	}
	if price.Failed || price.RealAmount != price.CoinsAmount {
		heap.Push(p.sortable(), price.CoinsAmount)
		return
	}
	heap.Push(p.sortedPrices, price)
	p.results[price.CoinsAmount] = price.Digs
}

func (p *PriceList) Map() map[int64]int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	m := map[int64]int64{}

	for key, val := range p.results {
		m[key] = val
	}

	return m
}

func (p *PriceList) calculateOptimalPrice(max int64) int64 {
	slc := *p.sortedPrices
	amount := max
	maxDigs := int64(0)
	for _, prc := range slc {
		if prc.CoinsAmount > max {
			break
		}
		if prc.Digs > maxDigs {
			maxDigs = prc.Digs
			amount = prc.CoinsAmount
		} else if prc.Digs == maxDigs && amount > prc.CoinsAmount {
			amount = prc.CoinsAmount
		}
	}
	return amount
}

func (p *PriceList) Next(coinsAmount int64) Price {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.counter++
	if p.counter % p.k == 0 {
		if len(p.experimentalAmounts) == 0 {
			return Price{
				CoinsAmount: p.calculateOptimalPrice(coinsAmount),
				RealAmount:  0,
				Digs:        0,
				experimental: false,
			}
		}

		c := heap.Pop(p.sortable()).(int64)

		if coinsAmount < c {
			heap.Push(p.sortable(), c)
			return Price{
				CoinsAmount: p.calculateOptimalPrice(coinsAmount),
				RealAmount:  0,
				Digs:        0,
				experimental: false,
			}
		}

		return Price{
			CoinsAmount:  c,
			RealAmount:   0,
			Digs:         0,
			experimental: true,
		}
	} else {
		return Price{
			CoinsAmount: p.calculateOptimalPrice(coinsAmount),
			RealAmount:  0,
			Digs:        0,
			experimental: false,
		}
	}
}