package interceptors

import (
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor возвращает интерсептор для восстановления после паники.
//
// При панике в handler логирует её и возвращает клиенту codes.Internal.
// Это позволяет MetricsInterceptor и LoggingInterceptor, стоящим выше
// в цепочке, увидеть корректный статус (а не "ok"), и не отдавать
// клиенту пустой успешный ответ.
func RecoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(
		recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Error("panic recovered", "panic", p)
			return status.Error(codes.Internal, "internal server error")
		}),
	)
}
