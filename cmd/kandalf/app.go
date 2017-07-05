package main

import (
	"net/url"

	"github.com/hellofresh/kandalf/pkg/amqp"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/producer"
	"github.com/hellofresh/kandalf/pkg/storage"
	"github.com/hellofresh/kandalf/pkg/workers"
	"github.com/hellofresh/stats-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RunApp is main application bootstrap and runner
func RunApp(cmd *cobra.Command, args []string) {
	log.WithField("version", version).Info("Kandalf starting...")

	globalConfig, err := config.Load(configPath)
	failOnError(err, "Failed to load application configuration")

	err = globalConfig.Log.Apply()
	failOnError(err, "Failed to configure logger")
	defer globalConfig.Log.Flush()

	pipesList, err := config.LoadPipesFromFile(globalConfig.Kafka.PipesConfig)
	failOnError(err, "Failed to load pipes config")

	statsClient, err := stats.NewClient(globalConfig.Stats.DSN, globalConfig.Stats.Prefix)
	failOnError(err, "Failed to connect to stats service")
	defer func() {
		if err := statsClient.Close(); err != nil {
			log.WithError(err).Error("Got error on closing stats client")
		}
	}()

	storageURL, err := url.Parse(globalConfig.StorageDSN)
	failOnError(err, "Failed to parse Storage DSN")

	persistentStorage, err := storage.NewPersistentStorage(storageURL)
	failOnError(err, "Failed to establish Redis connection")
	// Do not close storage here as it is required in Worker close to store unhandled messages

	kafkaProducer, err := producer.NewKafkaProducer(globalConfig.Kafka, statsClient)
	failOnError(err, "Failed to establish Kafka connection")
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.WithError(err).Error("Got error on closing kafka producer")
		}
	}()

	worker, err := workers.NewBridgeWorker(globalConfig.Worker, persistentStorage, kafkaProducer, statsClient)
	defer func() {
		if err := worker.Close(); err != nil {
			log.WithError(err).Error("Got error on closing persistent storage")
		}
	}()

	queuesHandler := amqp.NewQueuesHandler(pipesList, worker.MessageHandler, statsClient)
	amqpConnection, err := amqp.NewConnection(globalConfig.RabbitDSN, queuesHandler)
	failOnError(err, "Failed to establish initial connection to AMQP")
	defer func() {
		if err := amqpConnection.Close(); err != nil {
			log.WithError(err).Error("Got error on closing AMQP connection")
		}
	}()

	forever := make(chan bool)

	worker.Go(forever)

	log.Infof("[*] Waiting for users. To exit press CTRL+C")
	<-forever
}
