package producer

// Producer is an interface for publishing messages service
type Producer interface {
	Publish(msg Message) error
	Close() error
}
