package main

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/hellofresh/kandalf/pkg/amqp"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/producer"
	"github.com/hellofresh/kandalf/pkg/storage"
	"github.com/hellofresh/kandalf/pkg/workers"
	"github.com/hellofresh/stats-go"
	"github.com/hellofresh/stats-go/bucket"
	"github.com/hellofresh/stats-go/client"
	"github.com/hellofresh/stats-go/hooks"
	statsLogger "github.com/hellofresh/stats-go/log"
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

	statsClient := initStatsClient(globalConfig.Stats)
	defer func() {
		if err := statsClient.Close(); err != nil {
			log.WithError(err).Error("Got error on closing stats client")
		}
	}()

	pipesList, err := config.LoadPipesFromFile(globalConfig.Kafka.PipesConfig)
	failOnError(err, "Failed to load pipes config")

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

func initStatsClient(config config.StatsConfig) client.Client {
	statsLogger.SetHandler(func(msg string, fields map[string]interface{}, err error) {
		entry := log.WithFields(log.Fields(fields))
		if err == nil {
			entry.Debug(msg)
		} else {
			entry.WithError(err).Error(msg)
		}
	})

	statsClient, err := stats.NewClient(config.DSN)
	failOnError(err, "Failed to init stats client!")

	log.AddHook(hooks.NewLogrusHook(statsClient, config.ErrorsSection))

	host, err := os.Hostname()
	if nil != err {
		host = "-unknown-"
	}

	_, appFile := filepath.Split(os.Args[0])
	statsClient.TrackMetric("app", bucket.MetricOperation{"init", host, appFile})

	return statsClient
}
