package sqlstorage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// ListEventsDueForNotification возвращает события, для которых наступило время напоминания:
// notification_time задан и не позже текущего момента now.
func (s *Storage) ListEventsDueForNotification(ctx context.Context, now time.Time) ([]*storage.Event, error) {
	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
		WHERE notification_time IS NOT NULL AND notification_time <= $1
		ORDER BY notification_time ASC
	`

	rows, err := s.DB.QueryContext(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*storage.Event

	for rows.Next() {
		ev := &storage.Event{}
		var nt sql.NullTime

		if err := rows.Scan(
			&ev.ID,
			&ev.Title,
			&ev.DateTime,
			&ev.Duration,
			&ev.Description,
			&ev.UserID,
			&nt,
		); err != nil {
			return nil, err
		}

		ev.TimeNotification = notificationTimeFromNull(nt)
		out = append(out, ev)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// ClearNotificationTime сбрасывает notification_time после успешной постановки уведомления в очередь,
// чтобы не отправлять дубликаты на следующих тиках.
func (s *Storage) ClearNotificationTime(ctx context.Context, eventID uuid.UUID) error {
	const query = `UPDATE event SET notification_time = NULL WHERE id = $1`

	_, err := s.DB.ExecContext(ctx, query, eventID)
	return err
}

// DeleteEventsEndedBefore удаляет события, у которых момент окончания (date_time + duration)
// строго раньше порога before (для планировщика: «сейчас минус один год»).
func (s *Storage) DeleteEventsEndedBefore(ctx context.Context, before time.Time) (int64, error) {
	const query = `
		DELETE FROM event
		WHERE (date_time + (duration * interval '1 second')) < $1
	`

	res, err := s.DB.ExecContext(ctx, query, before)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
