// Package scheduler содержит цикл планировщика: выбор событий, публикация уведомлений в RabbitMQ,
// очистка старых событий. Зависит от интерфейсов хранилища и rabbitmq.Publisher, не от конкретного драйвера БД/AMQP.
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/rabbitmq"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// Log — минимальный набор методов логгера (реализуется internal/logger.Logger).
type Log interface {
	Info(msg string, params ...any)
	Error(msg string, params ...any)
	Warning(msg string, params ...any)
}

// Store — операции БД, нужные планировщику (реализует *sqlstorage.Storage).
type Store interface {
	ListEventsDueForNotification(ctx context.Context, now time.Time) ([]*storage.Event, error)
	ClearNotificationTime(ctx context.Context, eventID uuid.UUID) error
	DeleteEventsEndedBefore(ctx context.Context, before time.Time) (int64, error)
}

// Run выполняет периодический тик: уведомления + удаление «давно завершившихся» событий.
func Run(ctx context.Context, log Log, store Store, pub rabbitmq.Publisher, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Первый прогон сразу после старта, далее по таймеру.
	if err := tick(ctx, log, store, pub); err != nil {
		log.Error("scheduler tick: %s", err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := tick(ctx, log, store, pub); err != nil {
				log.Error("scheduler tick: %s", err.Error())
			}
		}
	}
}

func tick(ctx context.Context, log Log, store Store, pub rabbitmq.Publisher) error {
	now := time.Now()

	if err := dispatchNotifications(ctx, log, store, pub, now); err != nil {
		return err
	}

	cutoff := now.AddDate(-1, 0, 0)
	deleted, err := store.DeleteEventsEndedBefore(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("delete old events: %w", err)
	}
	if deleted > 0 {
		log.Info("scheduler: deleted %d event(s) ended before %s", deleted, cutoff.Format(time.RFC3339))
	}

	return nil
}

func dispatchNotifications(ctx context.Context, log Log, store Store, pub rabbitmq.Publisher, now time.Time) error {
	events, err := store.ListEventsDueForNotification(ctx, now)
	if err != nil {
		return fmt.Errorf("list due notifications: %w", err)
	}

	for _, ev := range events {
		n := storage.Notification{
			EventID:  ev.ID.String(),
			Title:    ev.Title,
			DateTime: ev.DateTime,
			UserID:   ev.UserID,
		}

		body, err := json.Marshal(n)
		if err != nil {
			log.Error("marshal notification for event %s: %s", ev.ID.String(), err.Error())
			continue
		}

		if err := pub.Publish(ctx, body); err != nil {
			log.Error("publish notification for event %s: %s", ev.ID.String(), err.Error())
			continue
		}

		if err := store.ClearNotificationTime(ctx, ev.ID); err != nil {
			log.Warning("clear notification_time for event %s: %s", ev.ID.String(), err.Error())
			continue
		}

		log.Info("scheduler: queued notification for event %s", ev.ID.String())
	}

	return nil
}
