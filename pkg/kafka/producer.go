package kafka

import (
	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/stats-go"
)

const (
	statsKafkaSection = "kafka"
)

// Producer contains connection data for Kafka
type Producer struct {
	kafkaClient sarama.SyncProducer
	statsClient stats.StatsClient
}

// NewProducer instantiates and establishes new Kafka connection
func NewProducer(kafkaConfig config.KafkaConfig, statsClient stats.StatsClient) (*Producer, error) {
	cnf := sarama.NewConfig()
	cnf.Producer.RequiredAcks = sarama.WaitForAll
	cnf.Producer.Retry.Max = kafkaConfig.MaxRetry
	// Producer.Return.Successes must be true to be used in a SyncProducer
	cnf.Producer.Return.Successes = true

	client, err := sarama.NewSyncProducer(kafkaConfig.Brokers, cnf)
	if err != nil {
		return nil, err
	}

	return &Producer{kafkaClient: client, statsClient: statsClient}, nil
}

// Close closes Kafka connection
func (p *Producer) Close() error {
	return p.kafkaClient.Close()
}

// Publish publishes message to Kafka
func (p *Producer) Publish(msg Message) error {
	_, _, err := p.kafkaClient.SendMessage(&sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.ByteEncoder(msg.Body),
	})

	if err == nil {
		log.WithField("msg", msg.String()).Debug("Successfully sent message to kafka")
	} else {
		log.WithError(err).WithField("msg", msg.String()).Error("Failed to publish message to kafka")
	}
	operation := stats.MetricOperation{"publish", msg.Topic, stats.MetricEmptyPlaceholder}
	p.statsClient.TrackOperation(statsKafkaSection, operation, nil, err == nil)

	return err
}
