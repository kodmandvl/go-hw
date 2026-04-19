package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage storage.EventStorage
}

type Logger interface {
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
}

func New(logger Logger, storage storage.EventStorage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event *storage.Event) error {
	// Генерируем ID на уровне приложения, чтобы HTTP/gRPC не дублировали логику.
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	return a.storage.CreateEvent(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, eventID uuid.UUID, event *storage.Event) error {
	return a.storage.UpdateEvent(ctx, eventID, event)
}

func (a *App) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	return a.storage.DeleteEvent(ctx, eventID)
}

func (a *App) GetEvents(ctx context.Context) ([]*storage.Event, error) {
	return a.storage.GetEvents(ctx)
}

func (a *App) GetEvent(ctx context.Context, eventID uuid.UUID) (*storage.Event, error) {
	return a.storage.GetEvent(ctx, eventID)
}

func (a *App) GetEventByDate(ctx context.Context, eventDatetime time.Time) (*storage.Event, error) {
	return a.storage.GetEventByDate(ctx, eventDatetime)
}

// ListEventsForDay — события за календарный день; userID пустой = без фильтра по пользователю.
func (a *App) ListEventsForDay(ctx context.Context, startOfDay time.Time, userID string) ([]*storage.Event, error) {
	events, err := a.storage.GetEventsForDay(ctx, startOfDay)
	if err != nil {
		return nil, err
	}
	return filterByUser(events, userID), nil
}

func (a *App) ListEventsForWeek(ctx context.Context, startOfWeek time.Time, userID string) ([]*storage.Event, error) {
	events, err := a.storage.GetEventsForWeek(ctx, startOfWeek)
	if err != nil {
		return nil, err
	}
	return filterByUser(events, userID), nil
}

func (a *App) ListEventsForMonth(ctx context.Context, startOfMonth time.Time, userID string) ([]*storage.Event, error) {
	events, err := a.storage.GetEventsForMonth(ctx, startOfMonth)
	if err != nil {
		return nil, err
	}
	return filterByUser(events, userID), nil
}

func filterByUser(events []*storage.Event, userID string) []*storage.Event {
	if userID == "" {
		return events
	}
	out := make([]*storage.Event, 0, len(events))
	for _, e := range events {
		if e.UserID == userID {
			out = append(out, e)
		}
	}
	return out
}
