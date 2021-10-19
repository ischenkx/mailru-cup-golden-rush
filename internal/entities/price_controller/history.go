package price_controller

type pricePoint struct {
	CPS float64
	Price int64
	counter float64
}

func (p *pricePoint) add(cps float64) {
	p.CPS = (p.CPS * p.counter+cps)/(p.counter+1)
	p.counter++
}

type priceHistory struct {
	points []pricePoint
	currentMax pricePoint
}

func (p *priceHistory) next(pnt pricePoint) pricePoint {
	p.points = append(p.points, pnt)
	if p.currentMax.CPS > pnt.CPS {
		return p.currentMax
	}
	p.currentMax = pnt
	newPoint := pricePoint{
		CPS:   0,
		Price: pnt.Price+1,
	}
	return newPoint
}
