package main

import (
	"flag"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/amqp"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/kafka"
	"github.com/hellofresh/kandalf/pkg/storage"
	"github.com/hellofresh/kandalf/pkg/workers"
	"github.com/hellofresh/stats-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.WithError(err).Panic(msg)
	}
}

func main() {
	var (
		globalConfig config.GlobalConfig
		err          error
	)

	configPath := flag.String("c", "", "Path to config file, set if you want to load settings from YAML file, otherwise settings are loaded from environment variables")
	flag.Parse()

	if *configPath != "" {
		globalConfig, err = config.LoadConfigFromFile(*configPath)
		failOnError(err, "Failed to load config from file")
	} else {
		globalConfig, err = config.LoadConfigFromEnv()
		failOnError(err, "Failed to load config from environment")
	}

	level, err := log.ParseLevel(strings.ToLower(globalConfig.LogLevel))
	failOnError(err, "Failed to get log level")
	log.SetLevel(level)

	pipesList, err := config.LoadPipesFromFile(globalConfig.Kafka.PipesConfig)
	failOnError(err, "Failed to load pipes config")

	statsClient := stats.NewStatsdStatsClient(globalConfig.Stats.DSN, globalConfig.Stats.Prefix)
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

	producer, err := kafka.NewProducer(globalConfig.Kafka, statsClient)
	failOnError(err, "Failed to establish Kafka connection")
	defer func() {
		if err := producer.Close(); err != nil {
			log.WithError(err).Error("Got error on closing kafka producer")
		}
	}()

	worker, err := workers.NewBridgeWorker(globalConfig.Worker, persistentStorage, producer, statsClient)
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
