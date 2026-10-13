package httputil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGrpcErrorToHTTP(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"nil", nil, 200, "OK"},
		{"invalid argument", status.Error(codes.InvalidArgument, "bad request"), 400, "bad request"},
		{"not found", status.Error(codes.NotFound, "not found"), 404, "not found"},
		{"already exists", status.Error(codes.AlreadyExists, "already exists"), 409, "already exists"},
		{"unauthenticated", status.Error(codes.Unauthenticated, "unauthorized"), 401, "unauthorized"},
		{"internal", status.Error(codes.Internal, "internal error"), 500, "internal error"},
		{"unimplemented", status.Error(codes.Unimplemented, "not implemented"), 501, "not implemented"},
		{"unavailable", status.Error(codes.Unavailable, "service unavailable"), 503, "service unavailable"},
		{"unknown", status.Error(codes.Unknown, "unknown"), 500, "unknown"},
		{"non-grpc error", assert.AnError, 500, "internal server error"},
		{"canceled", status.Error(codes.Canceled, "canceled"), 499, "canceled"},
		{"deadline exceeded", status.Error(codes.DeadlineExceeded, "timeout"), 504, "timeout"},
		{"permission denied", status.Error(codes.PermissionDenied, "forbidden"), 403, "forbidden"},
		{"resource exhausted", status.Error(codes.ResourceExhausted, "too many"), 429, "too many"},
		{"failed precondition", status.Error(codes.FailedPrecondition, "failed"), 400, "failed"},
		{"aborted", status.Error(codes.Aborted, "aborted"), 409, "aborted"},
		{"out of range", status.Error(codes.OutOfRange, "range"), 400, "range"},
		{"data loss", status.Error(codes.DataLoss, "data loss"), 500, "data loss"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, msg := GrpcErrorToHTTP(tt.err)
			assert.Equal(t, tt.wantStatus, statusCode)
			assert.Equal(t, tt.wantMsg, msg)
		})
	}

	t.Run("empty grpc message — fallback", func(t *testing.T) {
		statusCode, msg := GrpcErrorToHTTP(status.Error(codes.InvalidArgument, ""))

		assert.Equal(t, http.StatusBadRequest, statusCode)
		assert.Equal(t, "internal server error", msg)
	})
}
