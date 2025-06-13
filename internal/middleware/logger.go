package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

type (
	responseData struct {
		statusCode int
		size       int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return size, err
	}
	r.responseData.size += size
	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.statusCode = statusCode
}

func LoggingWiddleware(h http.Handler) http.Handler {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugar = *logger.Sugar()
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			statusCode: 0,
			size:       0,
		}

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(lrw, r)

		duration := time.Since(start)

		sugar.Infow("Request received",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
			"status_code", responseData.statusCode,
			"size", responseData.size,
		)
	}

	return http.HandlerFunc(logFn)
}
