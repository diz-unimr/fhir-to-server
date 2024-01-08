package fhir

import (
	"fhir-to-server/pkg/config"
	"github.com/jarcoal/httpmock"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// expected values
	retryConfig := config.Retry{Count: 2, Timeout: 3, Wait: 5, MaxWait: 7}
	expectedTimeout := 3 * time.Second
	expectedWait := 5 * time.Second
	expectedMaxWait := 7 * time.Second

	client := NewClient(config.Fhir{Retry: retryConfig})

	// config is set on client
	if retryConfig.Count != client.rest.RetryCount {
		t.Errorf("Expected retry count (%d) is not same as"+
			" actual (%d)", retryConfig.Count, client.rest.RetryCount)
	}
	if expectedTimeout != client.rest.GetClient().Timeout {
		t.Errorf("Expected client timeout (%d) is not same as"+
			" actual (%d)", expectedTimeout, client.rest.GetClient().Timeout)
	}
	if expectedWait != client.rest.RetryWaitTime {
		t.Errorf("Expected retry wait time (%d) is not same as"+
			" actual (%d)", expectedWait, client.rest.RetryWaitTime)
	}
	if expectedMaxWait != client.rest.RetryMaxWaitTime {
		t.Errorf("Expected retry max wait time (%d) is not same as"+
			" actual (%d)", expectedMaxWait, client.rest.RetryMaxWaitTime)
	}
}

func TestSend(t *testing.T) {
	const (
		baseUrl = "https://dummy-url/fhir"
		resp    = `"type": "transaction-response", "resourceType": "Bundle"`
	)
	client := NewClient(config.Fhir{Server: config.Server{BaseUrl: baseUrl}})

	// set up mock
	httpmock.ActivateNonDefault(client.rest.GetClient())
	responder := httpmock.NewStringResponder(200, resp)
	httpmock.RegisterResponder("POST", baseUrl, responder)

	b, _ := fhir.Bundle{Type: fhir.BundleTypeTransaction}.MarshalJSON()

	ok := client.Send(b)

	assert.True(t, ok)
}
