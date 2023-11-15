package fhir

import (
	"fhir-to-server/pkg/config"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	log "github.com/sirupsen/logrus"
)

type Processor struct {
	client *Client
	filter *DateFilter
}

func NewProcessor(config config.Fhir) *Processor {
	var filter *DateFilter
	if config.Filter.Date.Value == nil {
		filter = nil
	} else {
		filter = NewDateFilter(config.Filter.Date)
	}
	return &Processor{client: NewClient(config), filter: filter}
}

func (p *Processor) ProcessMessage(msg *kafka.Message) bool {

	// filter
	if p.filter != nil && !p.filter.apply(msg.Value) {
		// filtered, don't send but mark processed
		return true
	}

	success := p.client.Send(msg.Value)
	if success {
		log.WithFields(log.Fields{"topic": *msg.TopicPartition.Topic, "key": string(msg.Key), "offset": msg.TopicPartition.Offset.String()}).Debug("Successfully processed message")
		return true
	}

	log.WithFields(log.Fields{"topic": *msg.TopicPartition.Topic, "key": string(msg.Key), "offset": msg.TopicPartition.Offset.String()}).Error("Failed to process message")
	return false
}
