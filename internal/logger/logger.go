package logger

import (
	"go.uber.org/zap"
	"net/http"
)

func NewLogger() *zap.SugaredLogger {
	core, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	return core.Sugar()
}

type (
	responseData struct {
		Status int
		Size   int
	}

	LoggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *responseData
	}
)

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{ResponseWriter: w, ResponseData: &responseData{}}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.ResponseData.Status = code       // запоминаем статус
	lrw.ResponseWriter.WriteHeader(code) // вызываем оригинальный метод
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b) // записываем в реальный поток
	lrw.ResponseData.Size += size            // считаем байты
	return size, err
}
