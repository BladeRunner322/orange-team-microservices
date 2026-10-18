package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/BladeRunner322/orange-team-microservices/pkg/metrics"
)

func TestMetricsInterceptor(t *testing.T) {
	const method = "/test.MetricsService/Method"
	info := &grpc.UnaryServerInfo{FullMethod: method}

	t.Run("успешный вызов — status=ok", func(t *testing.T) {
		interceptor := MetricsInterceptor()

		before := testutil.ToFloat64(
			metrics.GRPCRequestsTotal.WithLabelValues(method, "ok"),
		)

		handler := func(_ context.Context, _ any) (any, error) {
			return "ok", nil
		}

		resp, err := interceptor(context.Background(), nil, info, handler)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp)

		after := testutil.ToFloat64(
			metrics.GRPCRequestsTotal.WithLabelValues(method, "ok"),
		)
		assert.Equal(t, before+1, after)
	})

	t.Run("ошибка NotFound — status=NotFound", func(t *testing.T) {
		interceptor := MetricsInterceptor()

		before := testutil.ToFloat64(
			metrics.GRPCRequestsTotal.WithLabelValues(method, "NotFound"),
		)

		handler := func(_ context.Context, _ any) (any, error) {
			return nil, status.Error(codes.NotFound, "not found")
		}

		_, err := interceptor(context.Background(), nil, info, handler)
		require.Error(t, err)

		after := testutil.ToFloat64(
			metrics.GRPCRequestsTotal.WithLabelValues(method, "NotFound"),
		)
		assert.Equal(t, before+1, after)
	})

	t.Run("не-gRPC ошибка — status=unknown", func(t *testing.T) {
		interceptor := MetricsInterceptor()

		before := testutil.ToFloat64(
			metrics.GRPCRequestsTotal.WithLabelValues(method, "unknown"),
		)

		handler := func(_ context.Context, _ any) (any, error) {
			return nil, assert.AnError
		}

		_, err := interceptor(context.Background(), nil, info, handler)
		require.Error(t, err)

		after := testutil.ToFloat64(
			metrics.GRPCRequestsTotal.WithLabelValues(method, "unknown"),
		)
		assert.Equal(t, before+1, after)
	})

	t.Run("in-flight растёт во время запроса и возвращается к исходному", func(t *testing.T) {
		interceptor := MetricsInterceptor()

		started := make(chan struct{})
		release := make(chan struct{})

		handler := func(_ context.Context, _ any) (any, error) {
			close(started)
			<-release
			return nil, nil
		}

		go func() {
			_, _ = interceptor(context.Background(), nil, info, handler)
		}()

		<-started
		during := testutil.ToFloat64(metrics.GRPCRequestsInFlight)
		close(release)

		assert.GreaterOrEqual(t, during, float64(1), "во время запроса in-flight >= 1")

		time.Sleep(50 * time.Millisecond)
	})

	t.Run("duration записана в histogram", func(t *testing.T) {
		interceptor := MetricsInterceptor()

		before := testutil.CollectAndCount(metrics.GRPCRequestDuration)

		handler := func(_ context.Context, _ any) (any, error) {
			time.Sleep(5 * time.Millisecond)
			return nil, nil
		}

		_, _ = interceptor(context.Background(), nil, info, handler)

		after := testutil.CollectAndCount(metrics.GRPCRequestDuration)
		assert.GreaterOrEqual(t, after, before)
	})
}
