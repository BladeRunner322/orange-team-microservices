package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusRecorder(t *testing.T) {
	t.Run("дефолтный статус 200, если WriteHeader не вызван", func(t *testing.T) {
		rr := httptest.NewRecorder()
		sr := NewStatusRecorder(rr)

		assert.Equal(t, http.StatusOK, sr.Status())
	})

	t.Run("запоминает явный WriteHeader", func(t *testing.T) {
		rr := httptest.NewRecorder()
		sr := NewStatusRecorder(rr)

		sr.WriteHeader(http.StatusNotFound)

		assert.Equal(t, http.StatusNotFound, sr.Status())
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("WriteHeader прокидывает код в оригинальный writer", func(t *testing.T) {
		rr := httptest.NewRecorder()
		sr := NewStatusRecorder(rr)

		sr.WriteHeader(http.StatusCreated)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Write без WriteHeader не меняет статус на 200", func(t *testing.T) {
		rr := httptest.NewRecorder()
		sr := NewStatusRecorder(rr)

		_, _ = sr.Write([]byte("ok"))

		assert.Equal(t, http.StatusOK, sr.Status())
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("WriteHeader после Write игнорируется httptest, но recorder помнит первый", func(t *testing.T) {
		rr := httptest.NewRecorder()
		sr := NewStatusRecorder(rr)

		sr.WriteHeader(http.StatusForbidden)

		assert.Equal(t, http.StatusForbidden, sr.Status())
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}
