// Package leaderboardgrpc реализует gRPC-сервер Leaderboard.
package leaderboardgrpc

import "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/leaderboard"

type Server struct {
	leaderboard.UnimplementedLeaderboardServiceServer
}

// TODO: Подключить сценарий чтения к GetDaily, GetWeekly и GetMonthly и преобразовать топ в protobuf.
func NewServer() *Server {
	return &Server{}
}
