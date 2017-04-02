package producer

import (
	"strings"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	body := "message body"
	topic := "message topic"

	msg := NewMessage([]byte(body), topic)
	assert.IsType(t, &Message{}, msg)

	assert.IsType(t, uuid.UUID{}, msg.ID)
	assert.Equal(t, body, string(msg.Body))
	assert.Equal(t, topic, msg.Topic)
}

func TestMessage_String(t *testing.T) {
	body := "message body"
	topic := "message topic"

	msg := NewMessage([]byte(body), topic)
	msgString := msg.String()
	assert.True(t, strings.Contains(msgString, msg.ID.String()))
	assert.True(t, strings.Contains(msgString, msg.Topic))
	assert.True(t, strings.Contains(msgString, "id"))
	assert.True(t, strings.Contains(msgString, "topic"))
}
