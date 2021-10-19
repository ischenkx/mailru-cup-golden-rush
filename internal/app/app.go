package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RomanIschenko/golden-rush-mailru/internal/config"
	area2 "github.com/RomanIschenko/golden-rush-mailru/internal/entities/area"
	"github.com/RomanIschenko/golden-rush-mailru/internal/entities/coin"
	"github.com/RomanIschenko/golden-rush-mailru/internal/entities/license"
	"github.com/RomanIschenko/golden-rush-mailru/internal/entities/price_controller"
	"github.com/RomanIschenko/golden-rush-mailru/internal/http/api"
	"github.com/RomanIschenko/golden-rush-mailru/internal/http/models"
	"github.com/RomanIschenko/golden-rush-mailru/internal/optimizers"
	"github.com/RomanIschenko/golden-rush-mailru/internal/util/mertics"
	"log"
	"math/rand"
	"time"
)

type App struct {
	wallet        *coin.Manager
	exploredAreas *area2.Queue
	api           *api.API
	licenses      *license.Manager
	metrics       *mertics.Metrics
	config        config.Config
	priceList     *license.PriceList

	priceController *price_controller.PriceController

	depthOptimizer *optimizers.DepthOptimizer
	unexploredAreas chan area2.Area
	treasures chan string
}
func (app *App) runCasher() {
	for t := range app.treasures {
		app.cash(t)
	}
}

func (app *App) cash(t string) (coins int64) {
	data, err := app.api.Cash(t)
	if err != nil {
		app.metrics.IncCounter("cash_errors")
		return
	}
	coins += int64(len(data))

	app.priceController.AddCoins(int64(len(data)))

	app.metrics.IncCounter("cash_ok")
	app.wallet.Add(data...)
	app.metrics.AddAverage("coins_per_treasure", float64(len(data)))
	return
}

func (app *App) initUnexploredAreas() {
	areas := area2.GenerateAreas(
		int64(app.config.App.World.SX),
		int64(app.config.App.World.SY),
		int64(app.config.App.World.Width),
		int64(app.config.App.World.Height),
		int64(app.config.App.Block.Width),
		int64(app.config.App.Block.Height),
	)

	rand.Shuffle(len(areas), func(i, j int) {
		t := areas[i]
		areas[i] = areas[j]
		areas[j] = t
	})

	for _, a := range areas {
		app.unexploredAreas <- a
	}
}

func (app *App) runLogger() {
	//app.metrics.AddMax("licenses_deleted", float64(app.licenses.DeletedLicenses()))
	t := time.NewTicker(app.config.Logger.Interval.Parse())
	for range t.C {
		m := map[string]interface{}{
			"current_score": app.wallet.Amount(),
			"best_depth": app.depthOptimizer.Best(),
			"app":      app.metrics.Snapshot(),
			//"poll_api": app.api.Metrics().Snapshot(),
			"price_list": app.priceList.Map(),
		}
		if data, err := json.Marshal(m); err == nil {
			log.Println(string(data))
		}
	}
}

func (app *App) runExplorer() {
	for {
		select {
		case a := <-app.unexploredAreas:
			report, err := app.api.Explore(models.Area{
				PosX:  a.X,
				PosY:  a.Y,
				SizeX: a.W,
				SizeY: a.H,
			})

			if err != nil {
				app.metrics.IncCounter("explore_errors")
				continue
			}
			app.metrics.IncCounter("explore_ok")
			a.Treasures = report.Amount
			app.metrics.AddAverage("treasures_per_block", float64(a.Treasures))
			if a.Treasures < int64(app.config.App.MinTreasuresPerBlock) {
				continue
			}
			app.metrics.AddCounter("treasures", float64(a.Treasures))
			app.exploredAreas.Push(a)
		}
	}
}

type location struct {
	X, Y, Treasures int64
	NoMore bool
}

func (app *App) runDigger() {
	areaChannel := make(chan area2.Area)
	locationChannel := make(chan location, 3)
	go func(app *App, areaChannel <-chan area2.Area, locationChannel chan <-location) {
		for a := range areaChannel {
			if a.Treasures <= 0 {
				locationChannel <- location{NoMore: true}
				continue
			}
			for x := a.X; x < a.X + a.W; x++ {
				if a.Treasures <= 0 {
					break
				}
				for y := a.Y; y < a.Y + a.H; y++ {
					if a.Treasures <= 0 {
						break
					}
					res, err := app.api.Explore(models.Area{
						PosX:  x,
						PosY:  y,
						SizeX: 1,
						SizeY: 1,
					})

					if err != nil {
						continue
					}

					if res.Amount <= 0 {
						continue
					}

					a.Treasures -= res.Amount

					locationChannel <- location{
						X:         x,
						Y:         y,
						Treasures: res.Amount,
						NoMore:    false,
					}
				}
			}
			locationChannel <- location{NoMore: true}
		}
	}(app, areaChannel, locationChannel)

	for {
		area := app.exploredAreas.Pop()
		areaChannel <- area

		var licenseHandle license.Handle

		for loc := range locationChannel {
			if loc.NoMore {
				break
			}
			depth := int64(1)
			maxDepth := app.depthOptimizer.Next()
			for depth <= maxDepth {
				if !licenseHandle.Active() {
					getLidStartTime := time.Now()
					licenseHandle = app.licenses.Get()
					app.metrics.AddAverage("get_lid_time", float64(time.Since(getLidStartTime)))
				}
				s := time.Now()
				// x, y, depth, licenseHandle.ID()
				result, err := app.api.Dig(models.Dig{
					Depth:     depth,
					LicenseID: licenseHandle.ID(),
					PosX:      loc.X,
					PosY:      loc.Y,
				})
				timePerDig := time.Since(s)
				app.metrics.AddAverage(fmt.Sprintf("dig_time_at_%d", depth), float64(timePerDig))
				app.metrics.AddCounter(fmt.Sprintf("treasures_%d_feet_deep", depth), float64(len(result)))
				app.metrics.AddAverage(fmt.Sprintf("treasures_%d_feet_deep", depth), float64(len(result)))
				if err != nil {
					app.metrics.IncCounter("dig_errors")
					switch err.(type) {
					case api.TreasureNotFoundErr:
						app.metrics.AddAverage("coins_per_dig", 0)
						app.metrics.IncCounter("empty_digs")
						depth++
						licenseHandle.Close()
					case api.NoSuchLicenseErr:
						app.metrics.IncCounter("no_such_licenses_errors")
						licenseHandle.Close()
					case api.WrongDepthErr:
						app.metrics.IncCounter("wrong_coordinates_errors(wrong_depth!)")
						app.metrics.AddMax("max_depth", float64(depth))
						licenseHandle.Close()
						depth++
					default:
						app.metrics.IncCounter("default_dig_errors")
					}
					continue
				}
				app.metrics.IncCounter("dig_ok")
				licenseHandle.Close()
				loc.Treasures -= int64(len(result))
				for _, t := range result {
					app.treasures <- t
				}
				app.metrics.IncCounter("treasures_put")
				depth++
				if loc.Treasures <= 0 {
					break
				}
			}
			if loc.Treasures > 0 {
				app.metrics.AddCounter("lost_treasures", float64(loc.Treasures))
			}
			app.metrics.IncCounter("finished_cells")
		}
	}
}

func (app *App) reRunPreExplorers(interval time.Duration) {
	t := time.NewTicker(interval)
	for range t.C {
		deadline := time.Now().Add(app.config.App.PreExplorationTimeout.Parse())
		for i := 0; i < 30; i++ {
			go app.preExplore(deadline, 1000)
		}
	}
}

func (app *App) preExplore(deadline time.Time, max int) {
	c := 0
	for {
		select {
		case ua := <-app.unexploredAreas:
			if time.Now().After(deadline) {
				app.metrics.AddAverage("preExplorations", float64(c))
				return
			}
			if c >= max {
				return
			}

			s := time.Now()
			rep, err := app.api.ExploreDeadline(deadline, models.Area{
				PosX:  ua.X,
				PosY:  ua.Y,
				SizeX: ua.W,
				SizeY: ua.H,
			})
			app.metrics.IncCounter("total_pre_explore")
			if err != nil {
				app.metrics.IncCounter("pre_explore_errors")
				continue
			}

			app.metrics.AddAverage(fmt.Sprintf("%dby%d_explore", ua.W, ua.H), float64(time.Since(s)))
			app.metrics.IncCounter("pre_explore_ok")
			ua.Treasures = rep.Amount
			if ua.Treasures < int64(app.config.App.MinTreasuresPerBlock) {
				continue
			}
			c++
			app.exploredAreas.PushWithoutBlocking(ua)
		}
	}
}
func (app *App) runLicenseIssuer() {
	for {
		handle := app.licenses.RequestAdd()

		maxCoins := app.priceController.GetPrice()

		price := app.priceList.Next(maxCoins)

		//fmt.Println("ready to spend:", price.CoinsAmount, price.Experimental(), maxCoins)

		coins := app.wallet.Pop(int(price.CoinsAmount))
		price.RealAmount = int64(len(coins))
		res, err := app.api.IssueLicenses(coins)
		if err != nil {
			handle.Fail()
			price.Failed = true
			app.priceList.Commit(price)
			app.wallet.Add(coins...)
			app.metrics.IncCounter(err.Error())
			continue
		}

		app.priceController.DeleteCoins(int64(len(coins)))

		app.metrics.AddCounter("spent_on_license", float64(price.CoinsAmount))
		app.metrics.AddAverage("license_price", float64(price.CoinsAmount))
		id := res.ID
		digs := res.DigAllowed - res.DigUsed
		price.Failed = false
		price.Digs = digs
		app.priceList.Commit(price)
		if digs <= 0 {
			handle.Fail()
			continue
		}
		if err := handle.Ok(id, digs); err != nil {
			app.metrics.IncCounter("license_add_errors")
		}
	}
}

func (app *App) Start(ctx context.Context) {
	go app.runLogger()
	go app.initUnexploredAreas()
	if err := app.api.HealthCheck(); err != nil {
		log.Println("failed to get response from health check")
		return

	}

	time.Sleep(3 * time.Second)

	if app.config.App.Block.Auto {
		s := time.Now()
		log.Println("auto_block_size_time:", time.Since(s))
		log.Println("(auto_size)", app.config.App.Block.Width, app.config.App.Block.Height)
	}

	preexplorationDeadline := time.Now().Add(app.config.App.PreExplorationTimeout.Parse())

	log.Println("preexploration deadline", preexplorationDeadline)

	for i := 0; i < app.config.App.PreExploreWorkers; i++ {
		go app.preExplore(preexplorationDeadline, app.config.App.MaxBlocksPerPreExploreWorker)
	}

	time.Sleep(app.config.App.PreExplorationTimeout.Parse())

	log.Println("finished preexploration")

	go app.reRunPreExplorers(time.Minute * 2)

	for i := 0; i < app.config.App.LicenseIssuers; i++ {
		go app.runLicenseIssuer()
	}

	for i := 0; i < app.config.App.Cashers; i++ {
		go app.runCasher()
	}

	for i := 0; i < app.config.App.Explorers; i++ {
		go app.runExplorer()
	}
	for i := 0; i < app.config.App.Diggers; i++ {
		go app.runDigger()
	}
	go app.priceController.Run(time.Millisecond * 250)
	<-ctx.Done()
}

func New(config config.Config) *App {
	return &App{
		wallet:          coin.NewManager(),
		exploredAreas:   area2.NewQueue(60),
		api:             api.New(config),
		treasures:       make(chan string, 100000),
		licenses:        license.NewManager(config.App.License.MaxAmount),
		metrics:         mertics.New(config.Logger.Enabled),
		config:          config,
		priceList:       license.NewPriceList(config),
		priceController: price_controller.New(),
		depthOptimizer:  optimizers.NewDepthOptimizer(config),
		unexploredAreas: make(chan area2.Area, 1000000),
	}
}
