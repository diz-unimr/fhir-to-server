package fhir

import (
	"fhir-to-server/pkg/config"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	if resp.RawResponse != nil {

		var logEvent *zerolog.Event
		if resp.IsSuccess() {
			logEvent = log.Debug()
		} else {
			logEvent = log.Error()
		}

		logEvent.
			Str("status", resp.Status()).
			Str("body", string(resp.Body())).Msg("FHIR server response")
	}

	return resp.IsSuccess()
}
