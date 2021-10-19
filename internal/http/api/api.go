package api

import (
	"github.com/RomanIschenko/golden-rush-mailru/internal/config"
	"github.com/RomanIschenko/golden-rush-mailru/internal/http/models"
	"github.com/RomanIschenko/golden-rush-mailru/internal/http/poller"
	"github.com/valyala/fasthttp"
	"time"
)

func validateResponse(data interface{}, err error) (interface{}, error, bool) {
	if err == nil {
		return data, nil, true
	}

	switch err.(type) {
	case NotStatedErr:
		return nil, err, false
	}

	return data, err, true
}

type API struct {
	pollers struct {
		dig, licenses, cash, explore, healthCheck *poller.Poller
	}
	client *client
}

// returns a treasure list and an error
func (api *API) Dig(data models.Dig) ([]string, error) {
	res, err := api.pollers.dig.Do(func(dl time.Time) (interface{}, error, bool) {
		return validateResponse(api.client.Dig(dl, data))
	})
	if result, ok := res.([]string); ok {
		return result, err
	}
	return nil, err
}

func (api *API) IssueLicenses(data []uint32) (models.License, error) {
	res, err := api.pollers.licenses.Do(func(dl time.Time) (interface{}, error, bool) {
		return validateResponse(api.client.IssueLicenses(dl, data))
	})
	if result, ok := res.(models.License); ok {
		return result, err
	}
	return models.License{}, err
}

func (api *API) ListLicenses() ([]models.License, error) {
	res, err := api.pollers.licenses.Do(func(dl time.Time) (interface{}, error, bool) {
		return validateResponse(api.client.ListLicenses(dl))
	})
	if result, ok := res.([]models.License); ok {
		return result, err
	}
	return nil, err
}

func (api *API) Cash(data string) ([]uint32, error) {
	res, err := api.pollers.cash.Do(func(dl time.Time) (interface{}, error, bool) {
		return validateResponse(api.client.Cash(dl, data))
	})
	if result, ok := res.([]uint32); ok {
		return result, err
	}
	return nil, err
}

func (api *API) Explore(data models.Area) (models.Report, error) {
	res, err := api.pollers.explore.Do(func(dl time.Time) (interface{}, error, bool) {
		return validateResponse(api.client.Explore(dl, data))
	})
	if result, ok := res.(models.Report); ok {
		return result, err
	}
	return models.Report{}, err
}

func (api *API) ExploreDeadline(deadline time.Time, data models.Area) (models.Report, error) {
	res, err := api.pollers.explore.DoDeadline(deadline, func(dl time.Time) (interface{}, error, bool) {
		return validateResponse(api.client.Explore(dl, data))
	})
	if result, ok := res.(models.Report); ok {
		return result, err
	}
	return models.Report{}, err
}

func (api *API) HealthCheck() error {
	_, err := api.pollers.healthCheck.Do(func(dl time.Time) (interface{}, error, bool) {
		if e := api.client.HealthCheck(dl); e != nil {
			return nil, e, false
		}
		return nil, nil, true
	})
	return err
}

func (api *API) initPollers(cfg config.Config) {
	api.pollers.licenses = poller.FromConfig(cfg.Api.IssueLicensePoller)
	api.pollers.dig = poller.FromConfig(cfg.Api.DigPoller)
	api.pollers.healthCheck = poller.FromConfig(cfg.Api.HealthCheckPoller)
	api.pollers.cash = poller.FromConfig(cfg.Api.CashPoller)
	api.pollers.explore = poller.FromConfig(cfg.Api.ExplorePoller)
}

func New(config config.Config) *API {
	api := &API{}

	api.client = newClient(config.BaseURL, &fasthttp.Client{
		MaxConnsPerHost:               50000,
		MaxIdleConnDuration:           time.Minute * 10,
		MaxConnDuration:               time.Minute * 10,
		ReadBufferSize:                2048,
		WriteBufferSize:               2048,
		ReadTimeout:                   time.Second * 5,
		MaxConnWaitTimeout:            time.Second * 30,
	})

	api.initPollers(config)

	return api
}