package fhir

import (
	"fhir-to-server/pkg/config"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessMessage(t *testing.T) {
	cases := []struct {
		name     string
		payload  []byte
		resultOk bool
		wasSent  bool
	}{
		{
			name:     "success",
			payload:  []byte(`{"resourceType": "Bundle","type": "batch","entry": [{"resource": {"resourceType": "Patient"}}]}`),
			resultOk: true,
			wasSent:  true,
		},
		{
			name:     "tombstone",
			payload:  nil,
			resultOk: true,
			wasSent:  false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			baseUrl := "https://dummy-url/fhir"
			conf := config.Fhir{
				Server: config.Server{
					BaseUrl: baseUrl,
				},
			}
			p := NewProcessor(conf)

			// set up mock
			httpmock.Reset()
			httpmock.ActivateNonDefault(p.client.rest.GetClient())
			responder := httpmock.NewStringResponder(200, `{"type": "batch-response", "entry": [{"response": {"status": "200"}}], "resourceType": "Bundle"}`)
			httpmock.RegisterResponder("POST", baseUrl, responder)

			testTopic := "test"
			ok := p.ProcessMessage(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &testTopic, Offset: 42},
				Value:          c.payload,
				Key:            []byte("test"),
			})
			assert.Equal(t, c.resultOk, ok, "Expected Kafka message to be processed ok: %s", c.resultOk)
			assert.Equal(t, c.wasSent, httpmock.GetTotalCallCount() == 1)
		})

	}
}
