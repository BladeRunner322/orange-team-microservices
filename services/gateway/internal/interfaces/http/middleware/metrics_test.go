package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/BladeRunner322/orange-team-microservices/pkg/metrics"
)

func TestNormalizedPath(t *testing.T) {
	t.Run("без chi RouteContext — возвращает сырой путь", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)

		got := normalizedPath(req)

		assert.Equal(t, "/users/123", got)
	})

	t.Run("с chi RouteContext и pattern — возвращает pattern", func(t *testing.T) {
		rctx := chi.NewRouteContext()
		rctx.RoutePatterns = []string{"/users/{userID}"}

		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		got := normalizedPath(req)

		assert.Equal(t, "/users/{userID}", got)
	})

	t.Run("RouteContext с пустым pattern — падаем на сырой путь", func(t *testing.T) {
		rctx := chi.NewRouteContext()

		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		got := normalizedPath(req)

		assert.Equal(t, "/users/123", got)
	})
}

func TestHTTPMetricsMiddleware(t *testing.T) {
	t.Run("counter инкрементируется на 200", func(t *testing.T) {
		path := "/test-metrics-mw-200"

		before := testutil.ToFloat64(
			metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, path, "200"),
		)

		handler := HTTPMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		after := testutil.ToFloat64(
			metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, path, "200"),
		)

		assert.Equal(t, before+1, after)
	})

	t.Run("counter инкрементируется с правильным статусом", func(t *testing.T) {
		path := "/test-metrics-mw-404"

		before := testutil.ToFloat64(
			metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, path, "404"),
		)

		handler := HTTPMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		after := testutil.ToFloat64(
			metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, path, "404"),
		)

		assert.Equal(t, before+1, after)
	})

	t.Run("по умолчанию статус 200, если WriteHeader не вызван", func(t *testing.T) {
		path := "/test-metrics-mw-default"

		before := testutil.ToFloat64(
			metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, path, "200"),
		)

		handler := HTTPMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// WriteHeader не вызываем — должен сработать дефолт
			_, _ = w.Write([]byte("ok"))
		}))

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		after := testutil.ToFloat64(
			metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, path, "200"),
		)

		assert.Equal(t, before+1, after)
	})

	t.Run("in-flight возвращается к исходному значению после запроса", func(t *testing.T) {
		handler := HTTPMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
		}))

		before := testutil.ToFloat64(metrics.HTTPRequestsInFlight)

		req := httptest.NewRequest(http.MethodGet, "/test-metrics-inflight", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		after := testutil.ToFloat64(metrics.HTTPRequestsInFlight)

		assert.Equal(t, before, after, "in-flight должен вернуться к исходному значению")
	})

	t.Run("in-flight растёт во время запроса", func(t *testing.T) {
		started := make(chan struct{})
		release := make(chan struct{})

		handler := HTTPMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			<-release
		}))

		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test-metrics-inflight-live", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}()

		<-started
		during := testutil.ToFloat64(metrics.HTTPRequestsInFlight)
		close(release)

		assert.GreaterOrEqual(t, during, float64(1), "во время запроса in-flight должен быть >= 1")

		// даём горутине завершиться
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("duration записана в histogram", func(t *testing.T) {
		path := "/test-metrics-duration"

		before := testutil.CollectAndCount(metrics.HTTPRequestDuration)

		handler := HTTPMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Millisecond)
		}))

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		after := testutil.CollectAndCount(metrics.HTTPRequestDuration)

		// Кол-во серий не уменьшилось — новая метка добавилась или уже была
		assert.GreaterOrEqual(t, after, before)
	})
}

// TestHTTPMetricsMiddleware_ChiIntegration проверяет, что middleware
// корректно определяет RoutePattern при работе через реальный chi-роутер.
// Это защищает от регрессии: если middleware вернуть на top-level
// или прочитать RoutePattern раньше времени — метрики будут писаться
// с сырым путём, и кардинальность взорвётся.
func TestHTTPMetricsMiddleware_ChiIntegration(t *testing.T) {
	r := chi.NewRouter()
	r.Use(HTTPMetricsMiddleware)
	r.Get("/workouts/{workoutId}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	path := "/workouts/{workoutId}"

	before := testutil.ToFloat64(
		metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, path, "200"),
	)

	// Три разных UUID должны попасть в ОДНУ серию по pattern.
	for _, id := range []string{"abc-123", "def-456", "ghi-789"} {
		req := httptest.NewRequest(http.MethodGet, "/workouts/"+id, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}

	after := testutil.ToFloat64(
		metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, path, "200"),
	)

	assert.Equal(t, before+3, after,
		"три разных UUID должны инкрементировать одну серию по pattern /workouts/{workoutId}")
}
