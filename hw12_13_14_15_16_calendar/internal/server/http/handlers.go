package internalhttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/timetz"
)

// eventPayload — тело запроса/ответа REST API (JSON), отделено от storage.Event для явных тегов API.
//
//nolint:tagliatelle // snake_case в JSON сознательно совпадает с полями storage.Event.
type eventPayload struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	DateTime         string    `json:"date_time"`
	Duration         int64     `json:"duration"`
	Description      string    `json:"description"`
	UserID           string    `json:"user_id"`
	TimeNotification *string   `json:"time_notification,omitempty"`
}

type listEventsResponse struct {
	Events []eventPayload `json:"events"`
}

type singleEventResponse struct {
	Event eventPayload `json:"event"`
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("calendar ok\n"))
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var p eventPayload
	if err := json.Unmarshal(body, &p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ev, err := payloadToEventCreate(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.app.CreateEvent(r.Context(), ev); err != nil {
		writeHTTPStorageError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, singleEventResponse{Event: eventToPayload(ev)})
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var p eventPayload
	if err := json.Unmarshal(body, &p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ev, err := payloadToEventUpdate(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.app.UpdateEvent(r.Context(), eventID, ev); err != nil {
		writeHTTPStorageError(w, err)
		return
	}
	updated, err := s.app.GetEvent(r.Context(), eventID)
	if err != nil {
		writeHTTPStorageError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, singleEventResponse{Event: eventToPayload(updated)})
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := s.app.DeleteEvent(r.Context(), eventID); err != nil {
		writeHTTPStorageError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listDay(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	dayStr := q.Get("day")
	userID := resolveUserID(r)
	t, err := timetz.ParseFlexible(dayStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	start := timetz.StartOfDayUTC(t)
	events, err := s.app.ListEventsForDay(r.Context(), start, userID)
	if err != nil {
		writeHTTPStorageError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listEventsResponse{Events: eventsToPayload(events)})
}

func (s *Server) listWeek(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("week_start")
	userID := resolveUserID(r)
	t, err := timetz.ParseFlexible(startStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	start := timetz.StartOfDayUTC(t)
	events, err := s.app.ListEventsForWeek(r.Context(), start, userID)
	if err != nil {
		writeHTTPStorageError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listEventsResponse{Events: eventsToPayload(events)})
}

func (s *Server) listMonth(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("month_start")
	userID := resolveUserID(r)
	t, err := timetz.ParseFlexible(startStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	start := timetz.StartOfMonthUTC(t)
	events, err := s.app.ListEventsForMonth(r.Context(), start, userID)
	if err != nil {
		writeHTTPStorageError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listEventsResponse{Events: eventsToPayload(events)})
}

func resolveUserID(r *http.Request) string {
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return r.Header.Get("X-User-Id")
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeHTTPStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, storage.ErrEventAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, storage.ErrEventDateTimeIsBusy):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func eventToPayload(e *storage.Event) eventPayload {
	p := eventPayload{
		ID:          e.ID,
		Title:       e.Title,
		DateTime:    e.DateTime.Format(timeRFC),
		Duration:    e.Duration,
		Description: e.Description,
		UserID:      e.UserID,
	}
	if !e.TimeNotification.IsZero() {
		s := e.TimeNotification.Format(timeRFC)
		p.TimeNotification = &s
	}
	return p
}

func eventsToPayload(events []*storage.Event) []eventPayload {
	out := make([]eventPayload, 0, len(events))
	for _, e := range events {
		out = append(out, eventToPayload(e))
	}
	return out
}

const timeRFC = "2006-01-02T15:04:05Z07:00"

func payloadToEventCreate(p eventPayload) (*storage.Event, error) {
	dt, err := timetz.ParseFlexible(p.DateTime)
	if err != nil {
		return nil, err
	}
	ev := &storage.Event{
		ID:          p.ID,
		Title:       p.Title,
		DateTime:    dt,
		Duration:    p.Duration,
		Description: p.Description,
		UserID:      p.UserID,
	}
	if p.TimeNotification != nil && *p.TimeNotification != "" {
		tn, err := timetz.ParseFlexible(*p.TimeNotification)
		if err != nil {
			return nil, err
		}
		ev.TimeNotification = tn
	}
	return ev, nil
}

func payloadToEventUpdate(p eventPayload) (*storage.Event, error) {
	// Для update id в теле игнорируем — идентификатор берётся из пути.
	return payloadToEventCreate(p)
}
