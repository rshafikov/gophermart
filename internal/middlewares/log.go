package middlewares

import (
	"fmt"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
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

		defer func() {
			status := fmt.Sprintf("%d %s", rData.status, http.StatusText(rData.status))
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
				outputColor = "\033[31;1m"
			}
			duration := time.Since(start)
			resetColor := []byte{'\033', '[', '0', 'm'}
			logger.L.Info(outputColor,
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.String("status", status),
				zap.String("duration", duration.String()),
				zap.String("from", r.RemoteAddr),
				zap.Int("size", rData.size),
			)

			_, err := os.Stdout.Write(resetColor)
			if err != nil {
				logger.L.Error("unable to reset color", zap.Error(err))
			}
		}()
	})
}
