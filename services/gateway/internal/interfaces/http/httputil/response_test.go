package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendJSON(t *testing.T) {
	t.Run("с данными — сериализует и пишет статус", func(t *testing.T) {
		rr := httptest.NewRecorder()

		SendJSON(rr, http.StatusOK, map[string]string{"status": "ok"})

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var resp map[string]string
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "ok", resp["status"])
	})

	t.Run("nil — только статус, тело пустое", func(t *testing.T) {
		rr := httptest.NewRecorder()

		SendJSON(rr, http.StatusNoContent, nil)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Empty(t, rr.Body.String())
	})

	t.Run("кастомный статус 201", func(t *testing.T) {
		rr := httptest.NewRecorder()

		SendJSON(rr, http.StatusCreated, map[string]any{"id": "123"})

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "123", resp["id"])
	})
}

func TestSendError(t *testing.T) {
	t.Run("пишет статус и JSON с полем error", func(t *testing.T) {
		rr := httptest.NewRecorder()

		SendError(rr, http.StatusBadRequest, "invalid input")

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var resp map[string]string
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Equal(t, "invalid input", resp["error"])
	})

	t.Run("пустое сообщение", func(t *testing.T) {
		rr := httptest.NewRecorder()

		SendError(rr, http.StatusInternalServerError, "")

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var resp map[string]string
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.Empty(t, resp["error"])
	})
}