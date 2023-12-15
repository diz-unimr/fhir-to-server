package fhir

import (
	"fhir-to-server/pkg/config"
	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"time"
)

type Client struct {
	rest   *resty.Client
	config config.Fhir
}

func NewClient(config config.Fhir) *Client {
	client := resty.New().
		SetLogger(log.New()).
		SetRetryCount(config.Retry.Count).
		SetTimeout(time.Duration(config.Retry.Timeout) * time.Second).
		SetRetryWaitTime(time.Duration(config.Retry.Wait) * time.Second).
		SetRetryMaxWaitTime(time.Duration(config.Retry.MaxWait) * time.Second)

	if config.Server.Auth != nil {
		client = client.SetBasicAuth(config.Server.Auth.User, config.Server.Auth.Password)
	}

	return &Client{rest: client, config: config}
}

func (c *Client) Send(fhir []byte) bool {
	resp, err := c.rest.R().
		SetBody(fhir).
		SetHeader("Content-Type", "application/fhir+json").
		Post(c.config.Server.BaseUrl)
	check(err)

	if resp.RawResponse != nil {
		respLog := log.WithFields(log.Fields{"status": resp.Status(), "body": string(resp.Body())})
		responseMsg := "FHIR server response"
		if resp.IsSuccess() {
			respLog.Debug(responseMsg)
		} else {
			respLog.Error(responseMsg)
		}
	}

	return resp.IsSuccess()
}
