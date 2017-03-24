package workers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/stats-go"
	"gopkg.in/Shopify/sarama.v1"
)

const (
	statsKafkaSection = "kafka"
)

type internalProducer struct {
	kafkaClient sarama.SyncProducer
	statsClient stats.StatsClient
}

// Returns new instance of kafka producer
func newInternalProducer(kafkaConfig config.KafkaConfig, statsClient stats.StatsClient) (*internalProducer, error) {
	cnf := sarama.NewConfig()
	cnf.Producer.RequiredAcks = sarama.WaitForAll
	cnf.Producer.Retry.Max = kafkaConfig.MaxRetry

	client, err := sarama.NewSyncProducer(kafkaConfig.Brokers, cnf)
	if err != nil {
		return nil, err
	}

	return &internalProducer{client, statsClient}, nil
}

// Sends message to the kafka
func (p *internalProducer) handleMessage(msg internalMessage) (err error) {
	_, _, err = p.kafkaClient.SendMessage(&sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.ByteEncoder(msg.Body),
	})

	if err == nil {
		log.WithField("topic", msg.Topic).Debug("Successfully sent message to kafka")
	} else {
		log.WithError(err).WithField("topic", msg.Topic).
			Error("An error occurred while sending message to kafka")
	}
	operation := stats.MetricOperation{"messages", stats.MetricEmptyPlaceholder, stats.MetricEmptyPlaceholder}
	p.statsClient.TrackOperation(statsKafkaSection, operation, nil, err == nil)

	return err
}
