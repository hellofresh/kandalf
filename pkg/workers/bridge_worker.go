package workers

import (
	"encoding/json"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/kandalf/pkg/kafka"
	"github.com/hellofresh/kandalf/pkg/storage"
	"github.com/hellofresh/stats-go"
)

const (
	statsWorkerSection = "worker"
)

type BridgeWorker struct {
	sync.Mutex

	config      config.WorkerConfig
	storage     storage.PersistentStorage
	producer    *kafka.Producer
	statsClient stats.StatsClient

	cache     []*kafka.Message
	lastFlush time.Time
}

func NewBridgeWorker(config config.WorkerConfig, storage storage.PersistentStorage, producer *kafka.Producer, statsClient stats.StatsClient) (*BridgeWorker, error) {
	return &BridgeWorker{config: config, storage: storage, producer: producer, statsClient: statsClient}, nil
}

func (w *BridgeWorker) Execute() {
	var messages []*kafka.Message

	w.Lock()
	defer w.Unlock()
	if len(w.cache) >= w.config.CacheSize || time.Now().Sub(w.lastFlush) >= w.config.CacheFlushTimeout.Duration {
		log.WithFields(log.Fields{"len": len(w.cache), "last_flush": w.lastFlush}).
			Debug("Flushing worker cache to Kafka")

		if len(w.cache) > 0 {
			// copy workers cache to local cache to avoid long locking for worker cache,
			// as all incoming messages will be waiting for network communication with kafka/storage
			copy(messages, w.cache)
			w.cache = []*kafka.Message{}
			go w.publishMessages(messages)
		}
		w.lastFlush = time.Now()
	}
}

func (w *BridgeWorker) Go(interrupt chan bool) {
	readStorageTicker := time.NewTicker(w.config.StorageReadTimeout.Duration)

	go func() {
		for {
			select {
			case <-interrupt:
				return
			case <-readStorageTicker.C:
				w.populatePoolFromStorage()
			default:
				w.Execute()
			}

			// Prevent CPU overload
			log.WithField("timeout", w.config.CycleTimeout).Debug("Bridge worker is going to sleep for a while")
			time.Sleep(time.Second * w.config.CycleTimeout.Duration)
		}
	}()
}

func (w *BridgeWorker) MessageHandler(body []byte, pipe config.Pipe) error {
	return w.cacheMessage(kafka.NewMessage(body, pipe.KafkaTopic))
}

func (w *BridgeWorker) cacheMessage(msg *kafka.Message) error {
	w.Lock()
	defer w.Unlock()

	w.cache = append(w.cache, msg)

	operation := stats.MetricOperation{"cache", "add", msg.Topic}
	w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)

	return nil
}

func (w *BridgeWorker) populatePoolFromStorage() {
	var msg kafka.Message

	for {
		operation := stats.MetricOperation{"storage", "get", stats.MetricEmptyPlaceholder}
		storageMsg, err := w.storage.Get()
		if err != nil {
			if err == storage.ErrStorageIsEmpty {
				break
			}
			log.WithError(err).Error("Failed to read message from persistent storage")
			w.statsClient.TrackOperation(statsWorkerSection, operation, nil, false)
			continue
		}
		w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)

		operation = stats.MetricOperation{"storage", "unmarshal", stats.MetricEmptyPlaceholder}
		err = json.Unmarshal(storageMsg, msg)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal message from persistent storage")
			w.statsClient.TrackOperation(statsWorkerSection, operation, nil, false)
			continue
		}
		w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)

		w.cacheMessage(&msg)
	}
}

func (w *BridgeWorker) publishMessages(messages []*kafka.Message) {
	for _, msg := range messages {
		err := w.producer.Publish(*msg)
		if err != nil {
			log.WithError(err).WithField("msg", msg.String()).
				Warning("Failed to publish messages to Kafka, moving to storage")

			operation := stats.MetricOperation{"storage", "marshal", stats.MetricEmptyPlaceholder}
			data, err := json.Marshal(msg)
			if err != nil {
				log.WithError(err).WithField("msg", msg.String()).
					Error("Failed to marshal message, dropping it")
				w.statsClient.TrackOperation(statsWorkerSection, operation, nil, false)
				continue
			}
			w.statsClient.TrackOperation(statsWorkerSection, operation, nil, true)

			operation = stats.MetricOperation{"storage", "set", stats.MetricEmptyPlaceholder}
			err = w.storage.Put(data)
			if err != nil {
				log.WithError(err).WithField("msg", msg.String()).
					Error("Failed to put message to storage, returning to cache")
				w.cacheMessage(msg)
			}
			w.statsClient.TrackOperation(statsWorkerSection, operation, nil, err == nil)
		}
	}
}
