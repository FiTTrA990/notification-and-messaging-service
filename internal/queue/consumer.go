package queue

import (
	"context"

	"github.com/FiTTrA990/notification-and-messaging-service/blob/master/internal/domain"
)

//go:generate mockgen -source=consumer.go -destination=../../mocks/mock_consumer.go -package=mocks

// MessageConsumer abstracts the queue consumer (RabbitMQ or Kafka)
type MessageConsumer interface {
	// Consume starts listening to the queue and calls handler for each message
	Consume(ctx context.Context, handler func(ctx context.Context, msg *domain.EventMessage) error) error
	// Close gracefully shuts down the consumer
	Close() error
}

// Publisher abstracts publishing back to a queue (e.g., for dead-letter or retry)
type Publisher interface {
	Publish(ctx context.Context, routingKey string, msg *domain.EventMessage) error
	Close() error
}

// RabbitMQConsumer is the RabbitMQ implementation (stub)
type RabbitMQConsumer struct {
	conn     interface{} // amqp091.Connection placeholder
	queueName string
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer (stub)
func NewRabbitMQConsumer(dsn, queueName string) (*RabbitMQConsumer, error) {
	// TODO: establish AMQP connection
	panic("NewRabbitMQConsumer: not implemented")
}

// Consume – STUB
func (r *RabbitMQConsumer) Consume(ctx context.Context, handler func(ctx context.Context, msg *domain.EventMessage) error) error {
	// TODO: implement AMQP consumer loop
	panic("RabbitMQConsumer.Consume: not implemented")
}

// Close – STUB
func (r *RabbitMQConsumer) Close() error {
	// TODO: close AMQP connection
	panic("RabbitMQConsumer.Close: not implemented")
}

// KafkaConsumer is the Kafka implementation (stub)
type KafkaConsumer struct {
	reader    interface{} // kafka.Reader placeholder
	groupID   string
	topic     string
}

// NewKafkaConsumer creates a new Kafka consumer (stub)
func NewKafkaConsumer(brokers []string, topic, groupID string) (*KafkaConsumer, error) {
	// TODO: set up kafka.Reader
	panic("NewKafkaConsumer: not implemented")
}

// Consume – STUB
func (k *KafkaConsumer) Consume(ctx context.Context, handler func(ctx context.Context, msg *domain.EventMessage) error) error {
	// TODO: implement Kafka fetch loop
	panic("KafkaConsumer.Consume: not implemented")
}

// Close – STUB
func (k *KafkaConsumer) Close() error {
	// TODO: close kafka.Reader
	panic("KafkaConsumer.Close: not implemented")
}
