package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
	"golang.org/x/exp/maps"
)

type Storage struct {
	mu     sync.RWMutex
	events map[uuid.UUID]*storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[uuid.UUID]*storage.Event),
	}
}

func (s *Storage) Connect(_ context.Context) error {
	return nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) CreateEvent(_ context.Context, event *storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.events[event.ID]; found {
		return storage.ErrEventAlreadyExists
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, eventID uuid.UUID, event *storage.Event) error {
	// Блокируем один раз: вызов GetEventByDate под локом привёл бы к взаимной блокировке.
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, found := s.events[eventID]; !found {
		return storage.ErrEventNotFound
	}

	for id, ev := range s.events {
		if id != eventID && ev.DateTime.Equal(event.DateTime) {
			return storage.ErrEventDateTimeIsBusy
		}
	}

	updated := *event
	updated.ID = eventID
	s.events[eventID] = &updated

	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, eventID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.events[eventID]; !found {
		return storage.ErrEventNotFound
	}

	delete(s.events, eventID)
	return nil
}

func (s *Storage) GetEvent(_ context.Context, eventID uuid.UUID) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, found := s.events[eventID]
	if !found {
		return nil, storage.ErrEventNotFound
	}

	return event, nil
}

func (s *Storage) GetEventByDate(_ context.Context, eventDatetime time.Time) (*storage.Event, error) {
	// because can be already locked by parent function.
	if s.mu.TryLock() {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	for _, event := range s.events {
		if event.DateTime == eventDatetime {
			return event, nil
		}
	}

	return nil, storage.ErrEventNotFound
}

func (s *Storage) GetEvents(_ context.Context) ([]*storage.Event, error) {
	return maps.Values(s.events), nil
}

// getEventsForRange — полуинтервал [startRange, endRange): начало включаем, конец исключаем.
func (s *Storage) getEventsForRange(startRange time.Time, endRange time.Time) []*storage.Event {
	var events []*storage.Event
	for _, event := range s.events {
		if !event.DateTime.Before(startRange) && event.DateTime.Before(endRange) {
			events = append(events, event)
		}
	}

	return events
}

func (s *Storage) GetEventsForDay(_ context.Context, startOfDay time.Time) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getEventsForRange(startOfDay, startOfDay.Add(24*time.Hour)), nil
}

func (s *Storage) GetEventsForWeek(_ context.Context, startOfWeek time.Time) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getEventsForRange(startOfWeek, startOfWeek.AddDate(0, 0, 7)), nil
}

func (s *Storage) GetEventsForMonth(_ context.Context, startOfMonth time.Time) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getEventsForRange(startOfMonth, startOfMonth.AddDate(0, 1, 0)), nil
}
