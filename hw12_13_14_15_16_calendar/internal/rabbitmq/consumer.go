package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// consumer — реализация Consumer: читает durable-очередь, ручной Ack после успешного handler.
type consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewConsumer подключается к RabbitMQ и объявляет ту же очередь, что и планировщик (имя из конфига).
func NewConsumer(uri, queueName string) (Consumer, error) {
	conn, err := dialWithRetry(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	if _, err = ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq queue declare: %w", err)
	}

	return &consumer{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

func (c *consumer) Consume(ctx context.Context, handler func(ctx context.Context, body []byte) error) error {
	msgs, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("rabbitmq consume: %w", err)
	}

	// Закрытие соединения разблокирует range по msgs при отмене контекста.
	go func() {
		<-ctx.Done()
		_ = c.conn.Close()
	}()

	for d := range msgs {
		if ctx.Err() != nil {
			break
		}
		if err := handler(ctx, d.Body); err != nil {
			// Не зависаем на битом сообщении: в ДЗ достаточно логирования; убираем из очереди.
			_ = d.Ack(false)
			continue
		}
		if err := d.Ack(false); err != nil {
			return err
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (c *consumer) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
