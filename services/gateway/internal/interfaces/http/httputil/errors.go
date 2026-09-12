package httputil

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcErrorToHTTP преобразует gRPC-ошибку в HTTP-статус и сообщение.
func GrpcErrorToHTTP(err error) (int, string) {
	if err == nil {
		return http.StatusOK, "OK"
	}
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, "internal server error"
	}
	msg := st.Message()
	if msg == "" {
		msg = "internal server error"
	}
	switch st.Code() {
	case codes.OK:
		return http.StatusOK, msg
	case codes.Canceled:
		return 499, msg
	case codes.Unknown:
		return http.StatusInternalServerError, msg
	case codes.InvalidArgument:
		return http.StatusBadRequest, msg
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, msg
	case codes.NotFound:
		return http.StatusNotFound, msg
	case codes.AlreadyExists:
		return http.StatusConflict, msg
	case codes.PermissionDenied:
		return http.StatusForbidden, msg
	case codes.Unauthenticated:
		return http.StatusUnauthorized, msg
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, msg
	case codes.FailedPrecondition:
		return http.StatusBadRequest, msg
	case codes.Aborted:
		return http.StatusConflict, msg
	case codes.OutOfRange:
		return http.StatusBadRequest, msg
	case codes.Unimplemented:
		return http.StatusNotImplemented, msg
	case codes.Internal:
		return http.StatusInternalServerError, msg
	case codes.Unavailable:
		return http.StatusServiceUnavailable, msg
	case codes.DataLoss:
		return http.StatusInternalServerError, msg
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
