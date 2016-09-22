package workers

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/redis.v3"

	"kandalf/config"
	"kandalf/logger"
	"kandalf/statsd"
)

type internalMessage struct {
	Body  json.RawMessage `json:"body"`
	Topic string          `json:"topic"`
}

type internalQueue struct {
	flushTimeout   time.Duration
	isWorking      bool
	maxSize        int
	messages       []internalMessage
	mutex          *sync.Mutex
	producer       *internalProducer
	rd             *redis.Client
	rdFlushTimeout time.Duration
	rdKeyName      string
}

// Returns new instance of queue
func newInternalQueue() (*internalQueue, error) {
	var err error

	rdAddr, err := config.Instance().String("queue.redis.address")
	if err != nil {
		return nil, err
	}

	rdClient := redis.NewClient(&redis.Options{
		Addr: rdAddr,
	})

	_, err = rdClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	producer, err := newInternalProducer()
	if err != nil {
		return nil, fmt.Errorf("An error occured while instantiating producer: %v", err)
	}

	qFlushTimeout, err := time.ParseDuration(config.Instance().UString("queue.flush_timeout", "5s"))
	if err != nil {
		qFlushTimeout = 5 * time.Second
	}

	qRedisFlushTimeout, err := time.ParseDuration(config.Instance().UString("queue.redis.flush_timeout", "10s"))
	if err != nil {
		qRedisFlushTimeout = 10 * time.Second
	}

	return &internalQueue{
		flushTimeout:   qFlushTimeout,
		maxSize:        config.Instance().UInt("queue.max_size", 10),
		messages:       make([]internalMessage, 0),
		mutex:          &sync.Mutex{},
		producer:       producer,
		rd:             rdClient,
		rdFlushTimeout: qRedisFlushTimeout,
		rdKeyName:      config.Instance().UString("queue.redis.key_name", "failed_messages"),
	}, nil
}

// Main working cycle
func (q *internalQueue) run(wg *sync.WaitGroup, die chan bool) {
	defer wg.Done()

	q.isWorking = true

	// Periodically flush the queue to the kafka
	go func() {
		var (
			lastFlush  time.Time = time.Now()
			nbMessages int
			oneSec     time.Duration = time.Second
		)

		for {
			select {
			case <-die:
				q.isWorking = false

				// Before exiting the working cycle, try to flush the messages to kafka
				q.handleMessages()
				return
			default:
			}

			nbMessages = len(q.messages)

			if (nbMessages >= q.maxSize) || (nbMessages > 0 && time.Now().Sub(lastFlush) > q.flushTimeout) {
				q.handleMessages()

				lastFlush = time.Now()
			} else {
				time.Sleep(oneSec)
			}
		}
	}()

	// And move messages from redis back to normal queue
	go q.flushRedis()
}

// Adds message to the internal queue which will be flushed to kafka later
func (q *internalQueue) add(msg internalMessage) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.messages = append(q.messages, msg)

	statsd.Instance().Increment("internal-queue.new-messages")

	logger.Instance().
		WithFields(log.Fields{
			"topic": msg.Topic,
		}).
		Debug("Added message to internal queue")
}

// Tries to send messages to the kafka
// If message failed, it will be stored in redis
func (q *internalQueue) handleMessages() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	var (
		err            error
		failedMessages []internalMessage = make([]internalMessage, 0)
	)

	logger.Instance().Debug("Start sending messages to kafka")

	for _, msg := range q.messages {
		err = q.producer.handleMessage(msg)
		if err == nil {
			continue
		} else {
			if strings.Contains(err.Error(), "topic or partition that does not exist") {
				continue
			}

			logger.Instance().
				WithError(err).
				Warning("Unable to send message to kafka")
		}

		err = q.storeInRedis(msg)

		// If it is not possible to store message even in redis,
		// we'll put the message into memory and process it later
		if err != nil {
			statsd.Instance().Increment("redis.failed-messages")

			failedMessages = append(failedMessages, msg)

			logger.Instance().
				WithError(err).
				Warning("Unable to store message in Redis")
		} else {
			statsd.Instance().Increment("redis.new-messages")

			logger.Instance().Debug("Successfully stored message in Redis")
		}
	}

	q.messages = failedMessages
}

// Stores the failed message in redis
func (q *internalQueue) storeInRedis(msg internalMessage) error {
	strMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return q.rd.LPush(q.rdKeyName, string(strMsg)).Err()
}

// Periodically move messages from Redis back to the main queue
func (q *internalQueue) flushRedis() {
	var (
		err      error
		bytesMsg []byte
		msg      internalMessage
	)

	for q.isWorking {
		for {
			bytesMsg, err = q.rd.LPop(q.rdKeyName).Bytes()
			if err != nil || len(bytesMsg) == 0 {
				break
			}

			err = json.Unmarshal(bytesMsg, &msg)
			if err != nil {
				logger.Instance().
					WithError(err).
					Warning("Unable to unmarshal JSON from redis")
				continue
			}

			q.add(msg)
		}

		time.Sleep(q.rdFlushTimeout)
	}

	// Close connection to redis
	err = q.rd.Close()
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("An error occurred while closing connection to redis")
	}
}
