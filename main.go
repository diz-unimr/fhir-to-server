package main

import (
	"fhir-to-server/pkg/config"
	"fhir-to-server/pkg/fhir"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	appConfig := loadConfig()
	configureLogger(appConfig.App)

	// create consumer and subscribe to input topics
	consumer := subscribeTo(appConfig)

	// create FHIR REST client
	client := fhir.NewClient(appConfig.Fhir)

	// signal handler to break the loop
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case sig := <-sigchan:
			log.WithField("signal", sig).Info("Caught signal. Terminating")
			return

		default:
			msg, err := consumer.ReadMessage(-1)

			if err == nil {
				log.WithFields(log.Fields{"key": string(msg.Key), "topic": *msg.TopicPartition.Topic}).Debug("Message received")
				processed := client.Send(msg.Value)

				if processed {
					offsets, err := consumer.CommitMessage(msg)
					if err != nil {
						log.WithError(err).Error("Failed to commit offsets. Terminating")
						os.Exit(1)
					}

					log.WithField("offsets", offsets).Trace("Offsets committed")
				} else {
					log.WithFields(log.Fields{"key": string(msg.Key), "topic": *msg.TopicPartition.Topic}).Error("Failed to process message. Terminating")
					os.Exit(1)
				}
			} else {
				// The client will automatically try to recover from all errors.
				log.WithError(err).Error("Consumer error")
			}
		}
	}
}

func subscribeTo(config config.AppConfig) *kafka.Consumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        config.Kafka.BootstrapServers,
		"security.protocol":        config.Kafka.SecurityProtocol,
		"ssl.ca.location":          config.Kafka.Ssl.CaLocation,
		"ssl.key.location":         config.Kafka.Ssl.KeyLocation,
		"ssl.certificate.location": config.Kafka.Ssl.CertificateLocation,
		"ssl.key.password":         config.Kafka.Ssl.KeyPassword,
		"broker.address.family":    "v4",
		"group.id":                 config.App.Name,
		"enable.auto.commit":       false,
		"auto.offset.reset":        "earliest",
	})

	if err != nil {
		panic(err)
	}

	err = consumer.SubscribeTopics(config.Kafka.InputTopics, nil)
	check(err)

	return consumer
}

func check(err error) {
	if err == nil {
		return
	}

	log.WithError(err).Error("Terminating")
	os.Exit(1)
}

func loadConfig() config.AppConfig {
	c, err := config.LoadConfig(".")
	if err != nil {
		log.WithError(err).Fatal("Unable to load config file")
	}
	return c
}

func configureLogger(config config.App) {
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)
}
