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

func TestClientMetricsInterceptor(t *testing.T) {
	const (
		target = "test-metrics-target:50051"
		method = "/test.MetricsService/Method"
	)

	// okInvoker — успешный вызов.
	okInvoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		return nil
	}

	// errInvoker — вызов с заданной ошибкой.
	errInvoker := func(err error) grpc.UnaryInvoker {
		return func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			return err
		}
	}

	t.Run("успешный вызов — status=ok", func(t *testing.T) {
		interceptor := ClientMetricsInterceptor(target)

		before := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(target, method, "ok"),
		)

		err := interceptor(context.Background(), method, nil, nil, nil, okInvoker)
		require.NoError(t, err)

		after := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(target, method, "ok"),
		)
		assert.Equal(t, before+1, after)
	})

	t.Run("ошибка NotFound — status=NotFound", func(t *testing.T) {
		interceptor := ClientMetricsInterceptor(target)

		before := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(target, method, "NotFound"),
		)

		err := interceptor(context.Background(), method, nil, nil, nil,
			errInvoker(status.Error(codes.NotFound, "not found")))
		require.Error(t, err)

		after := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(target, method, "NotFound"),
		)
		assert.Equal(t, before+1, after)
	})

	t.Run("транспортная ошибка Unavailable — status=Unavailable", func(t *testing.T) {
		interceptor := ClientMetricsInterceptor(target)

		before := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(target, method, "Unavailable"),
		)

		_ = interceptor(context.Background(), method, nil, nil, nil,
			errInvoker(status.Error(codes.Unavailable, "down")))

		after := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(target, method, "Unavailable"),
		)
		assert.Equal(t, before+1, after)
	})

	t.Run("не-gRPC ошибка — status=unknown", func(t *testing.T) {
		interceptor := ClientMetricsInterceptor(target)

		before := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(target, method, "unknown"),
		)

		_ = interceptor(context.Background(), method, nil, nil, nil,
			errInvoker(assert.AnError))

		after := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(target, method, "unknown"),
		)
		assert.Equal(t, before+1, after)
	})

	t.Run("duration записана в histogram", func(t *testing.T) {
		interceptor := ClientMetricsInterceptor(target)

		before := testutil.CollectAndCount(metrics.GRPCClientRequestDuration)

		slowInvoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}

		err := interceptor(context.Background(), method, nil, nil, nil, slowInvoker)
		require.NoError(t, err)

		after := testutil.CollectAndCount(metrics.GRPCClientRequestDuration)
		assert.GreaterOrEqual(t, after, before, "новая метка должна добавиться или уже быть")
	})

	t.Run("in-flight возвращается к исходному после запроса", func(t *testing.T) {
		interceptor := ClientMetricsInterceptor(target)

		before := testutil.ToFloat64(
			metrics.GRPCClientRequestsInFlight.WithLabelValues(target),
		)

		err := interceptor(context.Background(), method, nil, nil, nil, okInvoker)
		require.NoError(t, err)

		after := testutil.ToFloat64(
			metrics.GRPCClientRequestsInFlight.WithLabelValues(target),
		)
		assert.Equal(t, before, after, "in-flight должен вернуться к исходному")
	})

	t.Run("in-flight растёт во время запроса", func(t *testing.T) {
		interceptor := ClientMetricsInterceptor(target)

		started := make(chan struct{})
		release := make(chan struct{})

		blockingInvoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			close(started)
			<-release
			return nil
		}

		go func() {
			_ = interceptor(context.Background(), method, nil, nil, nil, blockingInvoker)
		}()

		<-started
		during := testutil.ToFloat64(
			metrics.GRPCClientRequestsInFlight.WithLabelValues(target),
		)
		close(release)

		assert.GreaterOrEqual(t, during, float64(1), "во время запроса in-flight >= 1")

		// Дать горутине завершиться.
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("разные target — независимые серии", func(t *testing.T) {
		const (
			targetA = "target-a:50051"
			targetB = "target-b:50051"
		)

		interceptorA := ClientMetricsInterceptor(targetA)
		interceptorB := ClientMetricsInterceptor(targetB)

		beforeA := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(targetA, method, "ok"),
		)
		beforeB := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(targetB, method, "ok"),
		)

		_ = interceptorA(context.Background(), method, nil, nil, nil, okInvoker)
		_ = interceptorB(context.Background(), method, nil, nil, nil, okInvoker)
		_ = interceptorB(context.Background(), method, nil, nil, nil, okInvoker)

		afterA := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(targetA, method, "ok"),
		)
		afterB := testutil.ToFloat64(
			metrics.GRPCClientRequestsTotal.WithLabelValues(targetB, method, "ok"),
		)

		assert.Equal(t, beforeA+1, afterA)
		assert.Equal(t, beforeB+2, afterB)
	})
}