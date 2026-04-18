package grpcserver

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	calendarv1 "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/pb/calendar/v1"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/timetz"
)

// eventToProto — маппинг доменной модели в protobuf (транспортный контракт).
func eventToProto(e *storage.Event) *calendarv1.Event {
	if e == nil {
		return nil
	}
	ev := &calendarv1.Event{
		Id:          e.ID.String(),
		Title:       e.Title,
		DateTime:    e.DateTime.UTC().Format(time.RFC3339Nano),
		Duration:    e.Duration,
		Description: e.Description,
		UserId:      e.UserID,
	}
	if !e.TimeNotification.IsZero() {
		ev.NotificationTime = e.TimeNotification.UTC().Format(time.RFC3339Nano)
	}
	return ev
}

func eventsToProto(events []*storage.Event) []*calendarv1.Event {
	out := make([]*calendarv1.Event, 0, len(events))
	for _, e := range events {
		out = append(out, eventToProto(e))
	}
	return out
}

// protoToEvent — разбор protobuf-сообщения в storage.Event (для create/update).
func protoToEvent(pb *calendarv1.Event) (*storage.Event, error) {
	if pb == nil {
		return nil, fmt.Errorf("event is required")
	}
	var id uuid.UUID
	if pb.GetId() != "" {
		parsed, err := uuid.Parse(pb.GetId())
		if err != nil {
			return nil, fmt.Errorf("id: %w", err)
		}
		id = parsed
	}

	dt, err := timetz.ParseFlexible(pb.GetDateTime())
	if err != nil {
		return nil, fmt.Errorf("date_time: %w", err)
	}

	var notify time.Time
	if pb.GetNotificationTime() != "" {
		notify, err = timetz.ParseFlexible(pb.GetNotificationTime())
		if err != nil {
			return nil, fmt.Errorf("notification_time: %w", err)
		}
	}

	return &storage.Event{
		ID:               id,
		Title:            pb.GetTitle(),
		DateTime:         dt,
		Duration:         pb.GetDuration(),
		Description:      pb.GetDescription(),
		UserID:           pb.GetUserId(),
		TimeNotification: notify,
	}, nil
}
