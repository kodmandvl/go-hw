package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultBaseURL     = "http://localhost:8888"
	defaultRabbitURI   = "amqp://rabbit:password@localhost:5672/"
	defaultStatusQueue = "calendar_notifications_status"
)

//nolint:tagliatelle // snake_case соответствует HTTP API календаря.
type eventPayload struct {
	ID               string `json:"id,omitempty"`
	Title            string `json:"title"`
	DateTime         string `json:"date_time"`
	Duration         int64  `json:"duration"`
	Description      string `json:"description"`
	UserID           string `json:"user_id"`
	TimeNotification string `json:"time_notification,omitempty"`
}

type listEventsResponse struct {
	Events []eventPayload `json:"events"`
}

func TestCreateEventAndBusinessError(t *testing.T) {
	baseURL := calendarBaseURL()
	waitForCalendarReady(t, baseURL)

	eventID := uuid.NewString()
	eventDate := uniqueBaseMonthStart().AddDate(0, 0, 5).Add(10 * time.Hour)
	payload := eventPayload{
		ID:          eventID,
		Title:       "integration-create-event",
		DateTime:    eventDate.Format(time.RFC3339),
		Duration:    3600,
		Description: "created by integration test",
		// В текущей SQL-схеме user_id имеет тип INTEGER, поэтому передаём числовой идентификатор.
		UserID: "1001",
	}

	status, body := postEvent(t, baseURL, payload)
	require.Equal(t, http.StatusCreated, status, string(body))

	// Повторная вставка с тем же id должна вернуть 409 (ErrEventAlreadyExists -> Conflict).
	status, body = postEvent(t, baseURL, payload)
	require.Equal(t, http.StatusConflict, status, string(body))
}

func TestListEventsForDayWeekMonth(t *testing.T) {
	baseURL := calendarBaseURL()
	waitForCalendarReady(t, baseURL)

	// В текущей SQL-схеме user_id имеет тип INTEGER, поэтому используем числовой идентификатор.
	userID := "1002"
	monthStart := uniqueBaseMonthStart()
	dayStart := monthStart.AddDate(0, 0, 10)

	events := []eventPayload{
		{
			ID:          uuid.NewString(),
			Title:       "listing-day",
			DateTime:    dayStart.Add(9 * time.Hour).Format(time.RFC3339),
			Duration:    1800,
			Description: "event for day/week/month",
			UserID:      userID,
		},
		{
			ID:          uuid.NewString(),
			Title:       "listing-week",
			DateTime:    dayStart.AddDate(0, 0, 1).Add(10 * time.Hour).Format(time.RFC3339),
			Duration:    1800,
			Description: "event for week/month",
			UserID:      userID,
		},
		{
			ID:          uuid.NewString(),
			Title:       "listing-month",
			DateTime:    dayStart.AddDate(0, 0, 10).Add(12 * time.Hour).Format(time.RFC3339),
			Duration:    1800,
			Description: "event for month",
			UserID:      userID,
		},
	}

	for _, ev := range events {
		status, body := postEvent(t, baseURL, ev)
		require.Equal(t, http.StatusCreated, status, string(body))
	}

	dayResp := getEvents(
		t,
		fmt.Sprintf("%s/api/events/day?day=%s&user_id=%s", baseURL, dayStart.Format("2006-01-02"), userID),
	)
	require.Len(t, dayResp.Events, 1)

	weekResp := getEvents(
		t,
		fmt.Sprintf("%s/api/events/week?week_start=%s&user_id=%s", baseURL, dayStart.Format(time.RFC3339), userID),
	)
	require.Len(t, weekResp.Events, 2)

	monthURL := fmt.Sprintf(
		"%s/api/events/month?month_start=%s&user_id=%s",
		baseURL,
		monthStart.Format(time.RFC3339),
		userID,
	)
	monthResp := getEvents(t, monthURL)
	require.GreaterOrEqual(t, len(monthResp.Events), 3)
}

func TestNotificationDeliveryStatus(t *testing.T) {
	baseURL := calendarBaseURL()
	waitForCalendarReady(t, baseURL)

	eventDate := uniqueBaseMonthStart().AddDate(0, 0, 25).Add(10 * time.Hour)
	payload := eventPayload{
		ID:          uuid.NewString(),
		Title:       "integration-notification-status",
		DateTime:    eventDate.Format(time.RFC3339),
		Duration:    1200,
		Description: "event for scheduler/sender flow",
		// В текущей SQL-схеме user_id имеет тип INTEGER, поэтому передаём числовой идентификатор.
		UserID:           "1003",
		TimeNotification: "2020-01-01T00:00:00Z", // В прошлом, чтобы scheduler взял событие сразу.
	}

	status, body := postEvent(t, baseURL, payload)
	require.Equal(t, http.StatusCreated, status, string(body))

	statusPayload := waitStatusMessageFromRabbit(t, payload.Title)
	require.Equal(t, "sent", statusPayload["status"])
	require.Equal(t, "calendar_sender", statusPayload["sender_service"])
}

func postEvent(t *testing.T, baseURL string, payload eventPayload) (int, []byte) {
	t.Helper()

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/events", bytes.NewReader(raw))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, body
}

func getEvents(t *testing.T, url string) listEventsResponse {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out listEventsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))

	return out
}

func waitForCalendarReady(t *testing.T, baseURL string) {
	t.Helper()

	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/", http.NoBody)
		if reqErr != nil {
			cancel()
			time.Sleep(2 * time.Second)
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		cancel()
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf("calendar service is not ready on %s", baseURL)
}

func waitStatusMessageFromRabbit(t *testing.T, expectedTitle string) map[string]string {
	t.Helper()

	conn, err := amqp.Dial(rabbitURI())
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	_, err = ch.QueueDeclare(statusQueue(), true, false, false, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(statusQueue(), "", false, false, false, false, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("notification status message was not received for title %q", expectedTitle)
		case msg := <-msgs:
			var payload map[string]string
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				_ = msg.Nack(false, false)
				continue
			}
			if strings.Contains(payload["notification"], expectedTitle) {
				require.NoError(t, msg.Ack(false))
				return payload
			}
			// Чужие сообщения интеграционных запусков отбрасываем.
			require.NoError(t, msg.Ack(false))
		}
	}
}

func calendarBaseURL() string {
	return envOrDefault("CALENDAR_BASE_URL", defaultBaseURL)
}

func rabbitURI() string {
	return envOrDefault("RABBIT_URI", defaultRabbitURI)
}

func statusQueue() string {
	return envOrDefault("RABBIT_STATUS_QUEUE", defaultStatusQueue)
}

func envOrDefault(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}

func uniqueBaseMonthStart() time.Time {
	now := time.Now().UTC()
	// Берём первое число следующего месяца: повторные прогоны не конфликтуют с фиксированными датами из прошлого.
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
}
