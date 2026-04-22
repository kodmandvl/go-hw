package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/lib/pq" // PG
	"github.com/pressly/goose"
)

type Storage struct {
	DB               *sql.DB
	connectionString string
}

func New(connectionString string) *Storage {
	return &Storage{
		connectionString: connectionString,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", s.connectionString)
	if err != nil {
		return err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	s.DB = db

	return s.migrate()
}

func (s *Storage) migrate() error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(s.DB, "migrations")
}

func (s *Storage) Close() error {
	return s.DB.Close()
}

// notificationTimeFromNull мапит NULL в БД в нулевое time.Time в модели (опциональное напоминание).
func notificationTimeFromNull(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

func (s *Storage) CreateEvent(ctx context.Context, event *storage.Event) error {
	// В SQL-хранилище сохраняем ту же бизнес-логику, что и в memory:
	// нельзя создать второе событие на тот же date_time.
	if _, err := s.GetEventByDate(ctx, event.DateTime); err == nil {
		return storage.ErrEventDateTimeIsBusy
	} else if !errors.Is(err, storage.ErrEventNotFound) {
		return err
	}

	// Явно передаём id: в миграции есть DEFAULT, но приложение всегда генерирует UUID заранее.
	const query = `
		INSERT INTO event (id, title, date_time, duration, description, user_id, notification_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.DB.ExecContext(
		ctx,
		query,
		event.ID,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.TimeNotification,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return storage.ErrEventAlreadyExists
		}
		return err
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, eventID uuid.UUID, event *storage.Event) error {
	// Проверяем конфликт по времени для других событий (кроме обновляемого).
	const checkBusyQuery = `
		SELECT COUNT(1)
		FROM event
		WHERE date_time = $1 AND id <> $2
	`
	var busyCount int
	if err := s.DB.QueryRowContext(ctx, checkBusyQuery, event.DateTime, eventID).Scan(&busyCount); err != nil {
		return err
	}
	if busyCount > 0 {
		return storage.ErrEventDateTimeIsBusy
	}

	const query = `
		UPDATE event
		SET title = $1, date_time = $2, duration = $3, description = $4, user_id = $5, notification_time = $6
		WHERE id = $7
	`

	res, err := s.DB.ExecContext(
		ctx,
		query,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.TimeNotification,
		eventID,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	const query = `DELETE FROM event WHERE id = $1`

	_, err := s.DB.ExecContext(ctx, query, eventID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetEvent(ctx context.Context, eventID uuid.UUID) (*storage.Event, error) {
	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
		WHERE id = $1
	`

	row := s.DB.QueryRowContext(ctx, query, eventID)

	var event storage.Event
	var nt sql.NullTime

	err := row.Scan(
		&event.ID,
		&event.Title,
		&event.DateTime,
		&event.Duration,
		&event.Description,
		&event.UserID,
		&nt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrEventNotFound
		}
		return nil, err
	}

	event.TimeNotification = notificationTimeFromNull(nt)

	return &event, nil
}

func (s *Storage) GetEvents(ctx context.Context) ([]*storage.Event, error) {
	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
	`
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var events []*storage.Event

	for rows.Next() {
		event := &storage.Event{}
		var nt sql.NullTime

		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.DateTime,
			&event.Duration,
			&event.Description,
			&event.UserID,
			&nt,
		)
		if err != nil {
			return nil, err
		}

		event.TimeNotification = notificationTimeFromNull(nt)
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) GetEventByDate(ctx context.Context, eventDatetime time.Time) (*storage.Event, error) {
	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
		WHERE date_time = $1
	`

	row := s.DB.QueryRowContext(ctx, query, eventDatetime)

	var event storage.Event
	var nt sql.NullTime

	err := row.Scan(
		&event.ID,
		&event.Title,
		&event.DateTime,
		&event.Duration,
		&event.Description,
		&event.UserID,
		&nt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrEventNotFound
		}
		return nil, err
	}

	event.TimeNotification = notificationTimeFromNull(nt)

	return &event, nil
}

// general mehtod for getting events by date range.
func (s *Storage) getEventsForRange(
	ctx context.Context,
	startRange time.Time,
	endRange time.Time,
) ([]*storage.Event, error) {
	var events []*storage.Event

	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
		WHERE date_time >= $1 AND date_time < $2
	`

	rows, err := s.DB.QueryContext(ctx, query, startRange, endRange)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate on the results of the query and create event objects
	for rows.Next() {
		event := &storage.Event{}
		var nt sql.NullTime

		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.DateTime,
			&event.Duration,
			&event.Description,
			&event.UserID,
			&nt,
		)
		if err != nil {
			return nil, err
		}

		event.TimeNotification = notificationTimeFromNull(nt)
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) GetEventsForDay(ctx context.Context, startOfDay time.Time) ([]*storage.Event, error) {
	return s.getEventsForRange(ctx, startOfDay, startOfDay.Add(24*time.Hour))
}

func (s *Storage) GetEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]*storage.Event, error) {
	return s.getEventsForRange(ctx, startOfWeek, startOfWeek.AddDate(0, 0, 7))
}

func (s *Storage) GetEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]*storage.Event, error) {
	return s.getEventsForRange(ctx, startOfMonth, startOfMonth.AddDate(0, 1, 0))
}
