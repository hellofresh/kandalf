package producer

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/hellofresh/stats-go"
	"github.com/stretchr/testify/assert"
)

type sendMessageResult struct {
	partition int32
	offset    int64
	err       error
}

type mockSyncProducer struct {
	sendMessageResult  sendMessageResult
	sendMessagesResult error
	closeResult        error

	lastSendMessageParams  *sarama.ProducerMessage
	lastSendMessagesParams []*sarama.ProducerMessage
}

func (p *mockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	p.lastSendMessageParams = msg
	return p.sendMessageResult.partition, p.sendMessageResult.offset, p.sendMessageResult.err
}

func (p *mockSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	p.lastSendMessagesParams = msgs
	return p.sendMessagesResult
}

func (p *mockSyncProducer) Close() error {
	return p.closeResult
}

func TestKafkaProducer_Close(t *testing.T) {
	closeError := errors.New("Close error result")

	mockProducer := &mockSyncProducer{closeResult: closeError}
	statsClient, _ := stats.NewClient("memory://", "")

	kafkaProducer := &KafkaProducer{mockProducer, statsClient}

	err := kafkaProducer.Close()
	assert.Error(t, err)
	assert.Equal(t, closeError, err)
}

func TestKafkaProducer_Publish(t *testing.T) {
	mockProducer := &mockSyncProducer{}
	statsClient, _ := stats.NewClient("memory://", "")

	body := "hello message body!"
	topic := "some topic"
	msg := NewMessage([]byte(body), topic)

	kafkaProducer := &KafkaProducer{mockProducer, statsClient}

	err := kafkaProducer.Publish(*msg)
	assert.NoError(t, err)

	memoryStats, _ := statsClient.(*stats.MemoryClient)
	assert.Equal(t, 1, memoryStats.CountMetrics[fmt.Sprintf("%s.publish.%s.-", statsKafkaSection, stats.SanitizeMetricName(topic))])
	assert.Equal(t, 1, memoryStats.CountMetrics[fmt.Sprintf("%s-ok.publish.%s.-", statsKafkaSection, stats.SanitizeMetricName(topic))])
	assert.Equal(t, 0, memoryStats.CountMetrics[fmt.Sprintf("%s-fail.publish.%s.-", statsKafkaSection, stats.SanitizeMetricName(topic))])

	messageValue, err := mockProducer.lastSendMessageParams.Value.Encode()
	assert.NoError(t, err)
	assert.Equal(t, body, string(messageValue))
	assert.Equal(t, topic, mockProducer.lastSendMessageParams.Topic)
}

func TestKafkaProducer_Publish_error(t *testing.T) {
	sendMessageError := errors.New("Send message error")
	sendMessageResult := sendMessageResult{0, 0, sendMessageError}

	mockProducer := &mockSyncProducer{sendMessageResult: sendMessageResult}
	statsClient, _ := stats.NewClient("memory://", "")

	body := "hello message body!"
	topic := "some topic"
	msg := NewMessage([]byte(body), topic)

	kafkaProducer := &KafkaProducer{mockProducer, statsClient}

	err := kafkaProducer.Publish(*msg)
	assert.Error(t, err)
	assert.Equal(t, sendMessageError, err)

	memoryStats, _ := statsClient.(*stats.MemoryClient)
	assert.Equal(t, 1, memoryStats.CountMetrics[fmt.Sprintf("%s.publish.%s.-", statsKafkaSection, stats.SanitizeMetricName(topic))])
	assert.Equal(t, 0, memoryStats.CountMetrics[fmt.Sprintf("%s-ok.publish.%s.-", statsKafkaSection, stats.SanitizeMetricName(topic))])
	assert.Equal(t, 1, memoryStats.CountMetrics[fmt.Sprintf("%s-fail.publish.%s.-", statsKafkaSection, stats.SanitizeMetricName(topic))])
}
