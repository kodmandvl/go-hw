package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// publisher — реализация Publisher на amqp091-go (не экспортируем, чтобы снаружи оставался только интерфейс).
type publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewPublisher подключается к RabbitMQ, объявляет durable-очередь и возвращает объект для публикации.
func NewPublisher(uri, queueName string) (Publisher, error) {
	conn, err := dialWithRetry(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	// Очередь должна существовать у consumer-а; объявление идемпотентно (durable + те же параметры).
	if _, err = ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq queue declare: %w", err)
	}

	return &publisher{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

// dialWithRetry даёт RabbitMQ время на старт после docker-compose up.
func dialWithRetry(uri string) (*amqp.Connection, error) {
	const (
		maxAttempts = 20
		retryDelay  = time.Second
	)

	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var conn *amqp.Connection
		conn, err = amqp.Dial(uri)
		if err == nil {
			return conn, nil
		}
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("rabbitmq dial: %w", err)
}

func (p *publisher) Publish(ctx context.Context, body []byte) error {
	return p.channel.PublishWithContext(ctx, "", p.queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (p *publisher) Close() error {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
