package clients

import (
	"context"
	"crypto/tls"

	"github.com/BladeRunner322/orange-team-microservices/internal/gen/api/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client auth.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	// TLS с самоподписанными сертификатами (пропускаем проверку)
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}
	return &AuthClient{
		conn:   conn,
		client: auth.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	resp, err := c.client.ValidateToken(ctx, &auth.ValidateTokenRequest{Token: token})
	if err != nil {
		return "", err
	}
	if !resp.Valid {
		return "", nil
	}
	return resp.UserId, nil
}

func (c *AuthClient) Register(ctx context.Context, email, password, fullName string) (*auth.RegisterResponse, error) {
	req := &auth.RegisterRequest{
		Email:    email,
		Password: password,
		FullName: fullName,
	}
	return c.client.Register(ctx, req)
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (*auth.LoginResponse, error) {
	req := &auth.LoginRequest{
		Email:    email,
		Password: password,
	}
	return c.client.Login(ctx, req)
}

func (c *AuthClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
