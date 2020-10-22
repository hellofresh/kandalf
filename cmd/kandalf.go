package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hellofresh/stats-go"
	"github.com/hellofresh/stats-go/bucket"
	"github.com/hellofresh/stats-go/client"
	"github.com/hellofresh/stats-go/hooks"
	statsLogger "github.com/hellofresh/stats-go/log"
	log "github.com/sirupsen/logrus"

	"github.com/hellofresh/kandalf/pkg/amqp"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/producer"
	"github.com/hellofresh/kandalf/pkg/storage"
	"github.com/hellofresh/kandalf/pkg/workers"
)

// RunApp is main application bootstrap and runner
func RunApp(version, configPath string) error {
	log.WithField("version", version).Info("Kandalf starting...")

	globalConfig, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load application configuration: %w", err)
	}

	err = globalConfig.Log.Apply()
	if err != nil {
		return fmt.Errorf("failed to configure logger: %w", err)
	}
	defer globalConfig.Log.Flush()

	statsClient, err := initStatsClient(globalConfig.Stats)
	if err != nil {
		return err
	}
	defer func() {
		if err := statsClient.Close(); err != nil {
			log.WithError(err).Error("Got error on closing stats client")
		}
	}()

	pipesList, err := config.LoadPipesFromFile(globalConfig.Kafka.PipesConfig)
	if err != nil {
		return fmt.Errorf("failed to load pipes config: %w", err)
	}

	storageURL, err := url.Parse(globalConfig.StorageDSN)
	if err != nil {
		return fmt.Errorf("failed to load pipes config: %w", err)
	}

	persistentStorage, err := storage.NewPersistentStorage(storageURL)
	if err != nil {
		return fmt.Errorf("failed to establish Redis connection: %w", err)
	}
	// Do not close storage here as it is required in Worker close to store unhandled messages

	kafkaProducer, err := producer.NewKafkaProducer(globalConfig.Kafka, statsClient)
	if err != nil {
		return fmt.Errorf("failed to establish Kafka connection: %w", err)
	}
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
	if err != nil {
		return fmt.Errorf("failed to establish initial connection to AMQP: %w", err)
	}
	defer func() {
		if err := amqpConnection.Close(); err != nil {
			log.WithError(err).Error("Got error on closing AMQP connection")
		}
	}()

	forever := make(chan bool)

	worker.Go(forever)

	log.Infof("[*] Waiting for users. To exit press CTRL+C")
	<-forever

	return nil
}

func initStatsClient(config config.StatsConfig) (client.Client, error) {
	statsLogger.SetHandler(func(msg string, fields map[string]interface{}, err error) {
		entry := log.WithFields(fields)
		if err == nil {
			entry.Debug(msg)
		} else {
			entry.WithError(err).Error(msg)
		}
	})

	statsClient, err := stats.NewClient(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to init stats client: %w", err)
	}

	log.AddHook(hooks.NewLogrusHook(statsClient, config.ErrorsSection))

	host, err := os.Hostname()
	if nil != err {
		host = "-unknown-"
	}

	_, appFile := filepath.Split(os.Args[0])
	statsClient.TrackMetric("app", bucket.NewMetricOperation("init", host, appFile))

	return statsClient, nil
}
