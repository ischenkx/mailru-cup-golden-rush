package area

type Area struct {
	X, Y, W, H, Treasures int64
}

func GenerateAreas(sx, sy, w, h, wstep, hstep int64) []Area {
	as := make([]Area, 0, 1000000)
	for x := sx; x < w; x += wstep {
		for y := sy; y < h; y += hstep {
			as = append(as, Area{
				X:         x,
				Y:         y,
				W:         wstep,
				H:         hstep,
				Treasures: 0,
			})
		}
	}

	return as
}


