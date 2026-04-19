// Package rabbitmq задаёт абстракции над брокером: процессы scheduler/sender зависят только от этих
// интерфейсов, конкретный клиент amqp091-go спрятан в отдельных файлах реализации.
package rabbitmq

import "context"

// Publisher публикует уже сериализованное тело сообщения (например JSON уведомления) в очередь.
type Publisher interface {
	Publish(ctx context.Context, body []byte) error
	Close() error
}

// Consumer читает сообщения из очереди и отдаёт тело в handler (рассыльщик логирует payload).
type Consumer interface {
	Consume(ctx context.Context, handler func(ctx context.Context, body []byte) error) error
	Close() error
}
