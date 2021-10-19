package price_controller

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)



type PriceController struct {
	currentPrice int64
	priceChan chan float64
	startCoef float64
	maxDelta float64
	totalCoins int64
}

func (p *PriceController) DeleteCoins(amount int64) {
	atomic.AddInt64(&p.totalCoins, -amount)
}

func (p *PriceController) AddCoins(amount int64) {
	atomic.AddInt64(&p.totalCoins, amount)
}

func (p *PriceController) runBenchmark(interval time.Duration) {
	ticker := time.NewTicker(interval)

	prevCoins := 0.

	avg := 0.

	counter := 0.

	prevTime := time.Now().UnixNano()

	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixNano()
			timeDelta := now - prevTime
			prevTime = now
			currentCoins := float64(atomic.LoadInt64(&p.totalCoins))
			coinsDelta := (currentCoins - prevCoins) / float64(timeDelta)
			prevCoins = currentCoins
			if coinsDelta == 0 {
				continue
			}
			avg = (avg * counter + coinsDelta)/(counter+1)
			counter++
			p.push(avg)
		}
	}
}

func (p *PriceController) push(coinsPerSecond float64) {
	p.priceChan <- coinsPerSecond
}

func (p *PriceController) GetPrice() int64 {
	return atomic.LoadInt64(&p.currentPrice)
}

func (p *PriceController) savePrice(price int64) {
	atomic.StoreInt64(&p.currentPrice, price)
}

func (p *PriceController) Run(interval time.Duration) {
	go p.runBenchmark(interval)

	currentPrice := p.GetPrice()

	counter := 0.
	counterBreaker := 30.

	currentAVGCPS := 0.

	maxCps := 0.
	maxCpsPrice := int64(0)

	currentCoef := 1.

	for cps := range p.priceChan {
		toBeIncreased, toBeDecreased := false, false

		if counter == 19 {
			fmt.Println(maxCps/currentAVGCPS, "- ratio", maxCps, currentAVGCPS)
		}

		if maxCps/currentAVGCPS >= 1.05 && counter > counterBreaker/5 || counter > counterBreaker/6 && currentAVGCPS <= 0 {
			toBeDecreased = true
		} else if counter >= counterBreaker {
			if currentAVGCPS > maxCps {
				maxCps = currentAVGCPS
				maxCpsPrice = currentPrice
			}
			toBeIncreased = true
		}

		if toBeDecreased {
			newPrice := maxCpsPrice
			if maxCpsPrice == currentPrice {
				newPrice = maxCpsPrice/2
				maxCpsPrice /= 2
				maxCps = 0.
			}
			log.Printf("%d => %d (%d)", currentPrice, newPrice, currentPrice - newPrice)
			currentPrice = newPrice
			currentAVGCPS = 0
			currentCoef = 1.
			counter = 0
		} else if toBeIncreased {
			log.Printf("%d => %d (%d)", currentPrice, currentPrice+1, 1)
			currentPrice += int64(currentCoef)
			currentCoef *= p.startCoef
			currentAVGCPS = 0
			counter = 0
		} else {
			currentAVGCPS = (currentAVGCPS * counter + cps)/(counter+1)
			counter++
		}
		p.savePrice(currentPrice)
	}
}

func New() *PriceController {
	return &PriceController{
		currentPrice: 0,
		priceChan:    make(chan float64),
		startCoef:    1.0003,
		totalCoins:   0,
	}
}