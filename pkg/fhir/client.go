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

	return &Client{rest: client, config: config}
}

func (c *Client) Send(fhir []byte) bool {
	resp, err := c.rest.R().
		SetBody(fhir).
		SetHeader("Content-Type", "application/json+fhir").
		Post(c.config.Server.BaseUrl)
	check(err)

	if resp.IsSuccess() {
		log.WithField("status", resp.Status()).Debug("Successfully sent bundle to FHIR server")
		log.WithField("body", string(resp.Body())).Trace("Response")
		return true
	}
	log.WithFields(log.Fields{"status": resp.Status(), "response-body": string(resp.Body())}).Error("Error sending bundle to FHIR server")
	return false
}

func check(err error) {
	if err == nil {
		return
	}

	log.Error(err)
}
