package grpcserver

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/app"
	calendarv1 "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/pb/calendar/v1"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/timetz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Logger совпадает по форме с логгером приложения (см. internal/server/http).
type Logger interface {
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
}

// CalendarService — связка сгенерированного gRPC-интерфейса с *app.App.
type CalendarService struct {
	calendarv1.UnimplementedCalendarServiceServer
	app *app.App
}

func NewCalendarService(application *app.App) *CalendarService {
	return &CalendarService{app: application}
}

// NewGRPCServer регистрирует сервис и подключает перехватчик логирования.
func NewGRPCServer(log Logger, application *app.App) *grpc.Server {
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(UnaryLoggingInterceptor(log)))
	calendarv1.RegisterCalendarServiceServer(srv, NewCalendarService(application))
	return srv
}

func (s *CalendarService) CreateEvent(
	ctx context.Context,
	req *calendarv1.CreateEventRequest,
) (*calendarv1.CreateEventResponse, error) {
	ev, err := protoToEvent(req.GetEvent())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.app.CreateEvent(ctx, ev); err != nil {
		return nil, mapStorageErr(err)
	}
	return &calendarv1.CreateEventResponse{Event: eventToProto(ev)}, nil
}

func (s *CalendarService) UpdateEvent(
	ctx context.Context,
	req *calendarv1.UpdateEventRequest,
) (*calendarv1.UpdateEventResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "id: %v", err)
	}
	ev, err := protoToEvent(req.GetEvent())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := s.app.UpdateEvent(ctx, id, ev); err != nil {
		return nil, mapStorageErr(err)
	}
	updated, err := s.app.GetEvent(ctx, id)
	if err != nil {
		return nil, mapStorageErr(err)
	}
	return &calendarv1.UpdateEventResponse{Event: eventToProto(updated)}, nil
}

func (s *CalendarService) DeleteEvent(
	ctx context.Context,
	req *calendarv1.DeleteEventRequest,
) (*calendarv1.DeleteEventResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "id: %v", err)
	}
	if err := s.app.DeleteEvent(ctx, id); err != nil {
		return nil, mapStorageErr(err)
	}
	return &calendarv1.DeleteEventResponse{}, nil
}

func (s *CalendarService) ListEventsForDay(
	ctx context.Context,
	req *calendarv1.ListEventsForDayRequest,
) (*calendarv1.ListEventsResponse, error) {
	t, err := timetz.ParseFlexible(req.GetDay())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "day: %v", err)
	}
	start := timetz.StartOfDayUTC(t)
	events, err := s.app.ListEventsForDay(ctx, start, req.GetUserId())
	if err != nil {
		return nil, mapStorageErr(err)
	}
	return &calendarv1.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func (s *CalendarService) ListEventsForWeek(
	ctx context.Context,
	req *calendarv1.ListEventsForWeekRequest,
) (*calendarv1.ListEventsResponse, error) {
	t, err := timetz.ParseFlexible(req.GetWeekStart())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "week_start: %v", err)
	}
	start := timetz.StartOfDayUTC(t)
	events, err := s.app.ListEventsForWeek(ctx, start, req.GetUserId())
	if err != nil {
		return nil, mapStorageErr(err)
	}
	return &calendarv1.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func (s *CalendarService) ListEventsForMonth(
	ctx context.Context,
	req *calendarv1.ListEventsForMonthRequest,
) (*calendarv1.ListEventsResponse, error) {
	t, err := timetz.ParseFlexible(req.GetMonthStart())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "month_start: %v", err)
	}
	start := timetz.StartOfMonthUTC(t)
	events, err := s.app.ListEventsForMonth(ctx, start, req.GetUserId())
	if err != nil {
		return nil, mapStorageErr(err)
	}
	return &calendarv1.ListEventsResponse{Events: eventsToProto(events)}, nil
}

func mapStorageErr(err error) error {
	switch {
	case errors.Is(err, storage.ErrEventNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, storage.ErrEventAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, storage.ErrEventDateTimeIsBusy):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
