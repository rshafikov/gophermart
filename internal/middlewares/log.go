package middlewares

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type (
	respData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		respData *respData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.respData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.respData.status = statusCode
}

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rData := &respData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			respData:       rData,
		}
		h.ServeHTTP(&lw, r)
		statusString := fmt.Sprintf("%d %s", rData.status, http.StatusText(rData.status))

		var outputColor string
		switch {
		case rData.status < 200:
			outputColor = "\033[34m"
		case rData.status < 300:
			outputColor = "\033[32m"
		case rData.status < 400:
			outputColor = "\033[36m"
		case rData.status < 500:
			outputColor = "\033[31m"
		default:
			outputColor = "\033[37m"
		}
		duration := time.Since(start)

		logger.L.Info(outputColor,
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.String("status", strings.ToUpper(statusString)),
			zap.String("duration", duration.String()),
			zap.String("from", r.RemoteAddr),
			zap.Int("size", rData.size),
		)
	})
}

type MinmalLogFormatter struct{}

func (f *MinmalLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &MinimalLogEntry{
		Request: r,
		Start:   time.Now(),
	}
	logger.L.Info(
		"[START]",
		zap.String("method", r.Method),
		zap.String("URI", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
	)
	return entry
}

type MinimalLogEntry struct {
	Request *http.Request
	Start   time.Time
}

func (e *MinimalLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	logger.L.Info(
		"[DONE]",
		zap.Int("status", status),
		zap.Int("bytes", bytes),
		zap.Duration("duration", elapsed),
	)
}

func (e *MinimalLogEntry) Panic(v interface{}, stack []byte) {
	logger.L.Panic("[PANIC]", zap.Reflect("v", v), zap.ByteString("stack", stack))
}
