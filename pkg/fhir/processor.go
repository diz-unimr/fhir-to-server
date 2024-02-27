package fhir

import (
	"fhir-to-server/pkg/config"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog/log"
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
		log.Debug().
			Str("topic", *msg.TopicPartition.Topic).
			Str("key", string(msg.Key)).
			Str("offset", msg.TopicPartition.Offset.String()).
			Msg("Successfully processed message")
		return true
	}

	log.Error().
		Str("topic", *msg.TopicPartition.Topic).
		Str("key", string(msg.Key)).
		Str("offset", msg.TopicPartition.Offset.String()).
		Msg("Failed to process message")
	return false
}
