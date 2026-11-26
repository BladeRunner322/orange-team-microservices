// Package leaderboardgrpc реализует gRPC-сервер Leaderboard.
package leaderboardgrpc

import "github.com/BladeRunner322/orange-team-microservices/internal/gen/api/leaderboard"

type Server struct {
	leaderboard.UnimplementedLeaderboardServiceServer
}

func NewServer() *Server {
	return &Server{}
}
