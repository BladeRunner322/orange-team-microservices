package utils

import (
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, msg := GrpcErrorToHTTP(tt.err)
			assert.Equal(t, tt.wantStatus, statusCode)
			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}
