package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	ctx := context.Background()

	baseCfg := func() Config {
		return Config{
			Target:  "localhost:50051",
			TLSMode: TLSModeDisabled,
			Timeout: 5 * time.Second,
		}
	}

	t.Run("minimal config TLSModeDisabled", func(t *testing.T) {
		conn, err := New(ctx, baseCfg())
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()
	})

	t.Run("TLSModeInsecure", func(t *testing.T) {
		cfg := baseCfg()
		cfg.TLSMode = TLSModeInsecure

		conn, err := New(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()
	})

	t.Run("TLSModeVerify", func(t *testing.T) {
		cfg := baseCfg()
		cfg.TLSMode = TLSModeVerify

		conn, err := New(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()
	})

	t.Run("CB enabled", func(t *testing.T) {
		cfg := baseCfg()
		cfg.CBEnabled = true

		conn, err := New(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()
	})

	t.Run("Retry enabled", func(t *testing.T) {
		cfg := baseCfg()
		cfg.RetryEnabled = true
		cfg.RetryMethods = []string{"/test.Service/Method"}

		conn, err := New(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()
	})

	t.Run("CB and Retry enabled", func(t *testing.T) {
		cfg := baseCfg()
		cfg.CBEnabled = true
		cfg.RetryEnabled = true
		cfg.RetryMethods = []string{"/test.Service/Method"}

		conn, err := New(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, conn)
		defer conn.Close()
	})

	t.Run("invalid target — validation error", func(t *testing.T) {
		cfg := baseCfg()
		cfg.Target = ""

		_, err := New(ctx, cfg)
		assert.Error(t, err)
	})
}
