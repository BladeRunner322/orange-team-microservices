package httputil

import "net/http"

// StatusRecorder оборачивает http.ResponseWriter и запоминает статус-код
// ответа. Если WriteHeader не вызывался — статус по умолчанию 200.
//
// Используется в middleware Gateway (logger, metrics), чтобы получить
// статус ответа после вызова next.ServeHTTP. Заменяет две прежние
// обёртки — responseWriterWrapper (logger) и metricsResponseWriter
// (metrics), которые были вложены друг в друга и дублировали логику.
type StatusRecorder struct {
	http.ResponseWriter
	status int
}

// NewStatusRecorder создаёт обёртку с дефолтным статусом 200.
func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// WriteHeader записывает статус-код в обёртку и прокидывает его дальше.
func (r *StatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Status возвращает записанный статус-код.
// Если WriteHeader не вызывался — 200.
func (r *StatusRecorder) Status() int {
	return r.status
}
