package interceptors

import (
	"github.com/BladeRunner322/orange-team-microservices/pkg/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
)

// RecoveryInterceptor возвращает интерсептор для восстановления после паники.
func RecoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(
		recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Error("panic recovered", "panic", p)
			return nil
		}),
	)
}
