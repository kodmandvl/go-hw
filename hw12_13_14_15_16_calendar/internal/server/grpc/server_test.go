package grpcserver

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/logger"
	calendarv1 "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/pb/calendar/v1"
	memorystorage "github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

// TestGRPCCreateListDelete — юнит-тест gRPC без сети (bufconn).
func TestGRPCCreateListDelete(t *testing.T) {
	t.Parallel()

	log := logger.New("error", io.Discard)
	st := memorystorage.New()
	application := app.New(log, st)

	lis := bufconn.Listen(bufSize)
	srv := NewGRPCServer(log, application)
	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(func() {
		srv.GracefulStop()
	})

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := calendarv1.NewCalendarServiceClient(conn)
	ctx := context.Background()

	created, err := client.CreateEvent(ctx, &calendarv1.CreateEventRequest{
		Event: &calendarv1.Event{
			Title:    "g",
			DateTime: "2025-01-02T11:00:00Z",
			Duration: 120,
			UserId:   "7",
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.GetEvent().GetId())

	list, err := client.ListEventsForDay(ctx, &calendarv1.ListEventsForDayRequest{
		Day:    "2025-01-02",
		UserId: "7",
	})
	require.NoError(t, err)
	require.Len(t, list.GetEvents(), 1)

	_, err = client.DeleteEvent(ctx, &calendarv1.DeleteEventRequest{Id: created.GetEvent().GetId()})
	require.NoError(t, err)

	_, err = client.DeleteEvent(ctx, &calendarv1.DeleteEventRequest{Id: created.GetEvent().GetId()})
	require.Error(t, err)
	stt, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, stt.Code())
}
