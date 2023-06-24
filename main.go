package main

import (
	"fhir-to-server/pkg/config"
	"fhir-to-server/pkg/fhir"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	log "github.com/sirupsen/logrus"
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

		go func(clientId string, topic string) {

			defer wg.Done()
			// create consumer and subscribe to input topics
			consumer := subscribe(appConfig, topic)
			log.WithFields(log.Fields{
				"topic":     topic,
				"group-id":  appConfig.App.Name,
				"client-id": clientId,
			}).Info("Consumer created")

			for {
				select {
				case <-sigchan:
					log.WithFields(log.Fields{
						"client-id": clientId,
						"topic":     topic,
					}).Info("Consumer shutting down gracefully")

					syncConsumerCommits(consumer)
					return
				default:
					msg, err := consumer.ReadMessage(1 * time.Second)
					if err == nil {

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
								sigchan <- syscall.SIGKILL
							}
						}
					}
				}
			}
		}(strconv.Itoa(i+1), topic)
	}
	<-sigchan
	close(sigchan)
	wg.Wait()

	log.Info("All consumers stopped")
}

func storeMessage(c *kafka.Consumer, msg *kafka.Message, clientId string) {
	_, err := c.StoreMessage(msg)
	if err != nil {
		log.WithFields(log.Fields{
			"client-id": clientId,
			"key":       string(msg.Key),
			"topic":     *msg.TopicPartition.Topic,
			"offset":    msg.TopicPartition.Offset.String()}).
			Warn("Failed to commit offset for message")
	} else {
		log.WithFields(log.Fields{
			"client-id": clientId,
			"key":       string(msg.Key),
			"topic":     *msg.TopicPartition.Topic,
			"offset":    msg.TopicPartition.Offset.String()}).
			Debug("Offset for message stored")
	}
}

func syncConsumerCommits(c *kafka.Consumer) {
	err := c.Unsubscribe()
	if err != nil {
		log.Error("Failed to unsubscribe consumer from the current subscription")
	}
	parts, err := c.Commit()
	if err != nil {
		if err.(kafka.Error).Code() == kafka.ErrNoOffset {
			return
		}
		log.WithError(err).Error("Failed to commit offsets")
	} else {

		for _, tp := range parts {
			log.WithFields(log.Fields{
				"topic":     *tp.Topic,
				"partition": tp.Partition,
				"offset":    tp.Offset.String()}).
				Info("Stored offsets committed")
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
