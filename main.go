package main

import (
	"errors"
	"fhir-to-server/pkg/config"
	"fhir-to-server/pkg/fhir"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	appConfig := loadConfig()
	configureLogger(appConfig.App)
	// signal handler to break the loop
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// create processor
	processor := fhir.NewProcessor(appConfig.Fhir)

	var wg sync.WaitGroup

	for i, topic := range appConfig.Kafka.InputTopics {
		wg.Add(1)
		clientId := strconv.Itoa(i + 1)

		go func() {
			defer wg.Done()

			// create consumer and subscribe to input topics
			consumer := subscribe(appConfig, topic)
			log.Info().
				Str("topic", topic).
				Str("group-id", appConfig.App.Name).
				Str("client-id", clientId).Msg("Consumer created")

			for {
				select {
				case <-sigchan:

					syncConsumerCommits(consumer)
					log.Info().
						Str("client-id", clientId).
						Str("topic", topic).
						Msg("Consumer shut down gracefully")
					return
				default:
					msg, err := consumer.ReadMessage(1 * time.Second)
					if err == nil {

						log.Debug().
							Str("client-id", clientId).
							Str("topic", topic).
							Str("key", string(msg.Key)).
							Msg("Message received")
						success := processor.ProcessMessage(msg)
						select {
						case <-sigchan:
							if success {
								storeMessage(consumer, msg, clientId)
							}
						default:
							if success {
								storeMessage(consumer, msg, clientId)
							} else {
								sigchan <- syscall.SIGTERM
							}
						}
					} else {
						var kafkaErr kafka.Error
						if errors.As(err, &kafkaErr) {
							// The client will automatically try to recover from all errors.
							// Timeout is not considered an error because it is raised by
							// ReadMessage in absence of messages.
							if kafkaErr.IsTimeout() {
								continue
							}

							log.Error().Err(kafkaErr).
								Str("client-id", clientId).
								Str("topic", topic).
								Msg("Consumer error")

							// Exceeding 'max.poll.interval.ms' makes the client leave the consumer group
							if kafkaErr.Code() == kafka.ErrMaxPollExceeded {
								sigchan <- syscall.SIGTERM
							}

						} else {
							log.Fatal().
								Err(kafkaErr).
								Str("client-id", clientId).
								Str("topic", topic).
								Msg("Unexpected error type")
							sigchan <- syscall.SIGTERM
						}
					}
				}
			}
		}()
	}
	<-sigchan
	close(sigchan)
	wg.Wait()

	log.Info().Msg("All consumers stopped")
}

func storeMessage(c *kafka.Consumer, msg *kafka.Message, clientId string) {
	_, err := c.StoreMessage(msg)

	var logEvent *zerolog.Event
	var logMsg string

	if err != nil {
		logEvent = log.Warn()
		logMsg = "Failed to commit offset for message"
	} else {
		logEvent = log.Debug()
		logMsg = "Offset for message stored"
	}

	logEvent.
		Str("client-id", clientId).
		Str("key", string(msg.Key)).
		Str("topic", *msg.TopicPartition.Topic).
		Str("offset", msg.TopicPartition.Offset.String()).
		Msg(logMsg)
}

func syncConsumerCommits(c *kafka.Consumer) {
	err := c.Unsubscribe()
	if err != nil {
		log.Error().Msg("Failed to unsubscribe consumer from the current subscription")
	}
	parts, err := c.Commit()
	if err != nil {
		if err.(kafka.Error).Code() == kafka.ErrNoOffset {
			return
		}
		log.Error().Err(err).Msg("Failed to commit offsets")
	} else {

		for _, tp := range parts {
			log.Info().
				Str("topic", *tp.Topic).
				Int32("partition", tp.Partition).
				Str("offset", tp.Offset.String()).
				Msg("Stored offsets committed")
		}
	}
	err = c.Close()
	check(err)
}

func subscribe(config config.AppConfig, topic string) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        config.Kafka.BootstrapServers,
		"security.protocol":        config.Kafka.SecurityProtocol,
		"ssl.ca.location":          config.Kafka.Ssl.CaLocation,
		"ssl.key.location":         config.Kafka.Ssl.KeyLocation,
		"ssl.certificate.location": config.Kafka.Ssl.CertificateLocation,
		"ssl.key.password":         config.Kafka.Ssl.KeyPassword,
		"broker.address.family":    "v4",
		"group.id":                 config.App.Name,
		"enable.auto.commit":       true,
		"enable.auto.offset.store": false,
		"auto.offset.reset":        "earliest",
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to Kafka")
	}

	err = consumer.Subscribe(topic, nil)
	check(err)

	return consumer
}

func check(err error) {
	if err == nil {
		return
	}

	log.Error().Err(err).Msg("Terminating")
	os.Exit(1)
}

func loadConfig() config.AppConfig {
	c, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to load config file")
	}
	return c
}

func configureLogger(appConfig config.App) {
	// log level
	logLevel, err := zerolog.ParseLevel(appConfig.LogLevel)
	if err == nil {
		zerolog.SetGlobalLevel(logLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// pretty logging (APP_ENV: dev)
	if appConfig.Env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
