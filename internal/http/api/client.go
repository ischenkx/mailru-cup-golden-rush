package api

import (
	"fmt"
	"github.com/RomanIschenko/golden-rush-mailru/internal/http/models"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"time"
)

type client struct {
	urls struct {
		dig,
		licenses,
		cash,
		explore,
		healthCheck string
	}

	httpClient *fasthttp.Client
}


// returns a treasure list and an error
func (c *client) Dig(deadline time.Time, data models.Dig) ([]string, error) {
	bts, err := data.MarshalJSON()

	if err != nil {
		return nil, err
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(c.urls.dig)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody(bts)

	err = c.httpClient.DoDeadline(req, res, deadline)

	if err != nil {
		return nil, err
	}

	switch res.StatusCode() {
	case 200:
		var treasures []string
		if err := jsoniter.Unmarshal(res.Body(), &treasures); err != nil {
			return nil, err
		}
		return treasures, nil
	case 403:
		return nil, NoSuchLicenseErr{}
	case 404:
		return nil, TreasureNotFoundErr{}
	case 422:
		var response models.Error
		if err := jsoniter.Unmarshal(res.Body(), &response); err != nil {
			return nil, err
		}
		switch response.Code {
		case 1000:
			return nil, WrongCoordinatesErr{}
		case 1001:
			return nil, WrongDepthErr{}
		default:
			return nil, NotStatedErr{}
		}
	default:
		return nil, NotStatedErr{}
	}
}

func (c *client) IssueLicenses(deadline time.Time, data []uint32) (models.License, error) {

	var license models.License

	bts, err := jsoniter.Marshal(data)

	if err != nil {
		return license, err
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(c.urls.licenses)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody(bts)

	err = c.httpClient.DoDeadline(req, res, deadline)

	if err != nil {
		return license, err
	}

	switch res.StatusCode() {
	case 200:
		if err := jsoniter.Unmarshal(res.Body(), &license); err != nil {
			return license, err
		}
		return license, nil
	case 409:
		return license, NoMoreLicensesAllowedErr{}
	default:
		return license, NotStatedErr{}
	}
}

func (c *client) ListLicenses(deadline time.Time) ([]models.License, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(c.urls.licenses)
	req.Header.SetMethod("GET")
	req.Header.SetContentType("application/json")

	err := c.httpClient.DoDeadline(req, res, deadline)

	if err != nil {
		return nil, err
	}

	switch res.StatusCode() {
	case 200:
		var licenseList []models.License
		if err := jsoniter.Unmarshal(res.Body(), &licenseList); err != nil {
			return nil, err
		}
		return licenseList, nil
	default:
		return nil, NotStatedErr{}
	}
}

func (c *client) Cash(deadline time.Time, data string) ([]uint32, error) {
	bts, err := jsoniter.Marshal(data)
	if err != nil {
		return nil, err
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(c.urls.cash)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody(bts)

	err = c.httpClient.DoDeadline(req, res, deadline)

	if err != nil {
		return nil, err
	}

	switch res.StatusCode() {
	case 200:
		var wallet []uint32
		if err := jsoniter.Unmarshal(res.Body(), &wallet); err != nil {
			return nil, err
		}
		return wallet, nil
	case 409:
		return nil, TreasureIsNotDugErr{}
	default:
		return nil, NotStatedErr{}
	}
}

func (c *client) Explore(deadline time.Time, data models.Area) (models.Report, error) {

	var report models.Report

	bts, err := jsoniter.Marshal(data)
	if err != nil {
		return report, err
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(c.urls.explore)
	req.Header.SetMethod("POST")

	req.Header.SetContentType("application/json")
	req.SetBody(bts)
	err = c.httpClient.DoDeadline(req, res, deadline)

	if err != nil {
		return models.Report{}, err
	}

	switch res.StatusCode() {
	case 200:
		if err := report.UnmarshalJSON(res.Body()); err != nil {
			return report, err
		}
		return report, nil
	case 422:
		return report, WrongCoordinatesErr{}
	default:
		return report, NotStatedErr{}
	}
}

func (c *client) HealthCheck(deadline time.Time) error {

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(c.urls.healthCheck)
	req.Header.SetMethod("GET")

	req.Header.SetContentType("application/json")

	err := c.httpClient.DoDeadline(req, res, deadline)
	if err != nil {
		return err
	}
	switch res.StatusCode() {
	case 200:
		return nil
	default:
		return NotStatedErr{}
	}
}

func (c *client) setupUrls(baseUrl string) {
	c.urls.licenses = fmt.Sprintf("%s/licenses", baseUrl)
	c.urls.dig = fmt.Sprintf("%s/dig", baseUrl)
	c.urls.cash = fmt.Sprintf("%s/cash", baseUrl)
	c.urls.explore = fmt.Sprintf("%s/explore", baseUrl)
	c.urls.healthCheck = fmt.Sprintf("%s/health-check", baseUrl)
}

func newClient(baseUrl string, httpClient *fasthttp.Client) *client {
	if httpClient == nil {
		httpClient = &fasthttp.Client{}
	}

	c := &client{
		httpClient: httpClient,
	}
	c.setupUrls(baseUrl)

	return c
}