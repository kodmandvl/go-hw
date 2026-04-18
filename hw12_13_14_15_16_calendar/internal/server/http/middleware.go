package internalhttp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kodmandvl/go-hw/hw12_13_14_15_16_calendar/internal/server/http/response"
)

// LoggingMiddleware — обёртка для тестов и для Server.Run (формат строки как в ДЗ №12).
func LoggingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := response.NewResponseWriter(w)
		initTime := time.Now()
		next.ServeHTTP(rw, r)
		latency := time.Since(initTime)
		serverLog(logger, rw, r, initTime, latency)
	})
}

func serverLog(logger Logger, rw *response.XResponseWriter, r *http.Request, time time.Time, latency time.Duration) {
	logger.Info(fmt.Sprintf(
		"%s [%s] %s %s %s %d %s \"%s\"",
		r.RemoteAddr,
		time.Format("2006-01-02 15:04:05"),
		r.Method,
		r.URL,
		r.Proto,
		rw.Code(),
		latency,
		r.UserAgent(),
	))
}
