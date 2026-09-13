package client

import (
	"crypto/tls"
	"fmt"
	"time"
)

// TLSMode определяет, как клиент устанавливает соединение с сервером.
type TLSMode string

const (
	// TLSModeDisabled — без шифрования (только для локальной разработки).
	TLSModeDisabled TLSMode = "disabled"

	// TLSModeInsecure — TLS без проверки сертификата сервера
	// (для self-signed сертификатов в dev/staging).
	TLSModeInsecure TLSMode = "insecure"

	// TLSModeVerify — TLS с проверкой сертификата (для прода).
	TLSModeVerify TLSMode = "verify"
)

// Config — параметры подключения к gRPC-серверу.
type Config struct {
	// Target — адрес сервера, например "auth-service:50051".
	Target string

	// TLSMode — режим TLS.
	TLSMode TLSMode

	// Timeout — таймаут на dial. Если 0 — 5 секунд по умолчанию.
	Timeout time.Duration
}

func (c Config) validate() error {
	if c.Target == "" {
		return fmt.Errorf("grpc client target is empty")
	}
	switch c.TLSMode {
	case TLSModeDisabled, TLSModeInsecure, TLSModeVerify:
		return nil
	default:
		return fmt.Errorf("unknown tls mode %q", c.TLSMode)
	}
}

// tlsConfig возвращает *tls.Config для данного режима.
// Для TLSModeDisabled возвращает nil.
func (c Config) tlsConfig() *tls.Config {
	switch c.TLSMode {
	case TLSModeVerify:
		return &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	case TLSModeInsecure:
		return &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true, //nolint:gosec // осознанно для self-signed
		}
	default:
		return nil
	}
}
