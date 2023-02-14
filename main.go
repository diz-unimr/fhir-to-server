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
	// signal handler to break the loop
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// create FHIR REST client
	client := fhir.NewClient(appConfig.Fhir)

	offsets := make(map[*kafka.Consumer]*kafka.Message)

	for _, topic := range appConfig.Kafka.InputTopics {
		t := topic
		go func() {
			// create consumer and subscribe to input topics
			consumer := subscribe(appConfig, t)
			defer closeConsumer(consumer)

			for {
				select {
				case <-sigchan:
					return

				default:
					msg, err := consumer.ReadMessage(2000)
					if err == nil {
						success := processMessages(client, msg)
						if success {
							offsets[consumer] = msg
						} else {
							sigchan <- syscall.SIGINT
						}
					} else {
						if err.(kafka.Error).Code() != kafka.ErrTimedOut {
							// The producer will automatically try to recover from all errors.
							log.WithError(err).Error("Consumer error")
						}
					}
				}
			}
		}()
	}
	<-sigchan

	log.Info("Caught signal. Shutting down gracefully")
	syncCommits(offsets)
}

func syncCommits(consumers map[*kafka.Consumer]*kafka.Message) {
	for c, msg := range consumers {
		parts, err := c.CommitMessage(msg)
		if err != nil {
			return
		}

		for _, tp := range parts {
			log.WithFields(log.Fields{"topic": *tp.Topic, "offset": tp.Offset.String()}).Trace("Offsets committed")
		}
	}
}

func syncCommits2(consumer *kafka.Consumer, msg *kafka.Message) {
	parts, err := consumer.CommitMessage(msg)
	if err != nil {
		return
	}

	for _, tp := range parts {
		log.WithFields(log.Fields{"topic": *tp.Topic, "offset": tp.Offset.String()}).Trace("Offsets committed")
	}

}

func closeConsumer(consumer *kafka.Consumer) {
	err := consumer.Close()
	if err != nil {
		log.Error("Failed to close consumer properly.")
	}
}

func processMessages(client *fhir.Client, msg *kafka.Message) bool {
	success := client.Send(msg.Value)
	if success {
		log.WithFields(log.Fields{"topic": *msg.TopicPartition.Topic, "key": string(msg.Key), "offset": msg.TopicPartition.Offset.String()}).Debug("Successfully processed message")
		return true
	} else {
		log.WithFields(log.Fields{"topic": *msg.TopicPartition.Topic, "key": string(msg.Key), "offset": msg.TopicPartition.Offset.String()}).Error("Failed to process message")
		return false
	}
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
		"auto.offset.reset":        "earliest",
	})

	if err != nil {
		panic(err)
	}

	err = consumer.Subscribe(topic, nil)
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
