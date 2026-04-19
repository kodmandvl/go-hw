package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// Server — HTTP-транспорт; знает только об интерфейсе Application (домен отделён).
type Server struct {
	host   string
	port   int
	logger Logger
	app    Application
	server *http.Server
}

type Logger interface {
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
}

// Application — ровно те операции, которые нужны REST-слою.
type Application interface {
	CreateEvent(ctx context.Context, event *storage.Event) error
	UpdateEvent(ctx context.Context, eventID uuid.UUID, event *storage.Event) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error
	GetEvent(ctx context.Context, eventID uuid.UUID) (*storage.Event, error)
	ListEventsForDay(ctx context.Context, startOfDay time.Time, userID string) ([]*storage.Event, error)
	ListEventsForWeek(ctx context.Context, startOfWeek time.Time, userID string) ([]*storage.Event, error)
	ListEventsForMonth(ctx context.Context, startOfMonth time.Time, userID string) ([]*storage.Event, error)
}

func NewServer(host string, port int, logger Logger, app Application) *Server {
	return &Server{
		host:   host,
		port:   port,
		logger: logger,
		app:    app,
	}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	// Совместимость с ДЗ №12: корневой путь остаётся "живым" проверочным endpoint.
	mux.HandleFunc("GET /{$}", s.health)
	mux.HandleFunc("POST /api/events", s.createEvent)
	mux.HandleFunc("PUT /api/events/{id}", s.updateEvent)
	mux.HandleFunc("DELETE /api/events/{id}", s.deleteEvent)
	mux.HandleFunc("GET /api/events/day", s.listDay)
	mux.HandleFunc("GET /api/events/week", s.listWeek)
	mux.HandleFunc("GET /api/events/month", s.listMonth)
	return mux
}

// Run блокируется до отмены ctx, затем делает graceful shutdown (как ListenAndServe + Shutdown).
func (s *Server) Run(ctx context.Context) error {
	handler := LoggingMiddleware(s.routes(), s.logger)
	s.server = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", s.host, s.port),
		Handler:           handler,
		ReadHeaderTimeout: 20 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		err := s.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
