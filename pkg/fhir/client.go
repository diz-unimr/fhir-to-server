package fhir

import (
	"fhir-to-server/pkg/config"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	models "github.com/samply/golang-fhir-models/fhir-models/fhir"
	"strconv"
	"time"
)

type Client struct {
	rest   *resty.Client
	config config.Fhir
}

func NewClient(fhir config.Fhir) *Client {
	client := resty.New().
		SetLogger(config.DefaultLogger()).
		SetRetryCount(fhir.Retry.Count).
		SetTimeout(time.Duration(fhir.Retry.Timeout) * time.Second).
		SetRetryWaitTime(time.Duration(fhir.Retry.Wait) * time.Second).
		SetRetryMaxWaitTime(time.Duration(fhir.Retry.MaxWait) * time.Second)

	if fhir.Server.Auth != nil {
		client = client.SetBasicAuth(fhir.Server.Auth.User, fhir.Server.Auth.Password)
	}

	return &Client{rest: client, config: fhir}
}

func (c *Client) Send(fhir []byte) bool {
	resp, err := c.rest.R().
		SetBody(fhir).
		SetHeader("Content-Type", "application/fhir+json").
		Post(c.config.Server.BaseUrl)
	check(err)

	// http response status
	success := resp.IsSuccess()

	if resp.RawResponse != nil {
		// http response status and FHIR response status
		success = success && responseSuccess(resp.Body())
	}

	var logEvent *zerolog.Event
	if success {
		logEvent = log.Debug()
	} else {
		logEvent = log.Error()
	}

	logEvent.
		Str("status", resp.Status()).
		Str("body", string(resp.Body())).Msg("FHIR server response")

	return success
}

func responseSuccess(body []byte) bool {
	// check BundleEntryResponse status
	b, err := models.UnmarshalBundle(body)
	if err != nil {
		check(err)
		return false
	}

	for _, e := range b.Entry {
		status, err := strconv.Atoi(e.Response.Status[0:3])

		if err != nil || !statusSuccess(status) {
			return false
		}
	}

	return true
}

func statusSuccess(status int) bool {
	return status > 199 && status < 300
}
