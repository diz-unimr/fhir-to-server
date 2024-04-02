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
	cases := []struct {
		name     string
		baseUrl  string
		code     int
		resp     string
		expected bool
	}{
		{
			name:     "success",
			baseUrl:  "https://dummy-url/fhir",
			code:     200,
			resp:     `{"type": "transaction-response", "resourceType": "Bundle"}`,
			expected: true,
		},
		{
			name:    "error",
			baseUrl: "https://dummy-url/fhir",
			code:    200,
			resp: `{"type": "batch-response",
					"entry": [
					{
					  "response": {
						"status": "422",
						"outcome": {
						  "issue": [
							{
							  "severity": "error",
							  "code": "not-supported",
							  "diagnostics": "Unsupported method 'PATCH'.",
							  "expression": [
								"Bundle.entry[0].request.method"
							  ]
							}
						  ],
						  "resourceType": "OperationOutcome"
						}
					  }
					}
					], "resourceType": "Bundle"}`,
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			client := NewClient(config.Fhir{Server: config.Server{BaseUrl: c.baseUrl}})

			// set up mock
			httpmock.ActivateNonDefault(client.rest.GetClient())
			responder := httpmock.NewStringResponder(200, c.resp)
			httpmock.RegisterResponder("POST", c.baseUrl, responder)

			b, _ := fhir.Bundle{Type: fhir.BundleTypeTransaction}.MarshalJSON()

			actual := client.Send(b)

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestResponseSuccess(t *testing.T) {

	cases := []struct {
		name     string
		response string
		expected bool
	}{
		{
			name:     "unprocessable",
			response: `{"type": "batch-response", "entry": [{"response": {"status": "422"}}], "resourceType": "Bundle"}`,
			expected: false,
		},
		{
			name:     "empty",
			response: `{"type": "batch-response", "entry": [], "resourceType": "Bundle"}`,
			expected: true,
		},
		{
			name:     "ok",
			response: `{"type": "batch-response", "entry": [{"response": {"status": "200"}}], "resourceType": "Bundle"}`,
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := responseSuccess([]byte(c.response))

			assert.Equal(t, actual, c.expected)
		})
	}
}
