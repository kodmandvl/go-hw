package internalhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/logger"
	memorystorage "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

// TestRESTFlow — юнит-тест уровня HTTP API без поднятия реального TCP-порта (httptest.Server).
func TestRESTFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := logger.New("error", io.Discard)
	st := memorystorage.New()
	application := app.New(log, st)

	srv := NewServer("", 0, log, application)
	ts := httptest.NewServer(LoggingMiddleware(srv.routes(), log))
	t.Cleanup(ts.Close)

	day := "2025-06-01T10:00:00Z"
	body := map[string]any{
		"title":     "meet",
		"date_time": day,
		"duration":  int64(3600),
		"user_id":   "42",
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/api/events", bytes.NewReader(raw))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)
	var created singleEventResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&created))
	require.NoError(t, res.Body.Close())
	id := created.Event.ID
	require.NotEqual(t, uuid.Nil, id)

	q := ts.URL + "/api/events/day?day=2025-06-01&user_id=42"
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, q, nil)
	require.NoError(t, err)
	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	var list listEventsResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&list))
	require.NoError(t, res.Body.Close())
	require.Len(t, list.Events, 1)

	req, err = http.NewRequestWithContext(ctx, http.MethodDelete, ts.URL+"/api/events/"+id.String(), nil)
	require.NoError(t, err)
	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
	require.NoError(t, res.Body.Close())

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, q, nil)
	require.NoError(t, err)
	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&list))
	require.NoError(t, res.Body.Close())
	require.Len(t, list.Events, 0)
}

// TestRESTHealth — корневой GET сохраняет простой ответ как в ДЗ №12.
func TestRESTHealth(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := logger.New("error", io.Discard)
	application := app.New(log, memorystorage.New())
	srv := NewServer("", 0, log, application)
	ts := httptest.NewServer(LoggingMiddleware(srv.routes(), log))
	t.Cleanup(ts.Close)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.Contains(t, string(b), "calendar")
}

// TestRESTListWeek — проверка диапазона недели на уровне HTTP.
func TestRESTListWeek(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := logger.New("error", io.Discard)
	st := memorystorage.New()
	application := app.New(log, st)
	srv := NewServer("", 0, log, application)
	ts := httptest.NewServer(LoggingMiddleware(srv.routes(), log))
	t.Cleanup(ts.Close)

	ev := bytes.NewReader([]byte(`{
  "title": "w",
  "date_time": "2025-03-05T15:00:00Z",
  "duration": 60,
  "user_id": "1"
}`))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/api/events", ev)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)
	require.NoError(t, res.Body.Close())

	u := ts.URL + "/api/events/week?week_start=2025-03-03T00:00:00Z&user_id=1"
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	require.NoError(t, err)
	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	var list listEventsResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&list))
	require.NoError(t, res.Body.Close())
	require.Len(t, list.Events, 1)
}
