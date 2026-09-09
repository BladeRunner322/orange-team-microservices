package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/any", nil)
	rr := httptest.NewRecorder()
	ProxyHandler(rr, req)

	assert.Equal(t, http.StatusNotImplemented, rr.Code)
	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "endpoint not implemented yet", resp["error"])
}
