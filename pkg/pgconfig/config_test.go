package pgconfig

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setAllRequired(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_HOST", "postgres-test")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_TIMEOUT", "30s")
}

func TestLoad_Success(t *testing.T) {
	setAllRequired(t)
	t.Setenv("POSTGRES_PORT", "5433")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "postgres-test", cfg.Host)
	assert.Equal(t, "5433", cfg.Port)
	assert.Equal(t, "testuser", cfg.User)
	assert.Equal(t, "testpass", cfg.Password)
	assert.Equal(t, "testdb", cfg.Database)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}

func TestLoad_DefaultPort(t *testing.T) {
	setAllRequired(t)

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "5432", cfg.Port)
}

func TestLoad_MissingHost(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_TIMEOUT", "30s")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "POSTGRES_HOST")
}

func TestLoad_MissingUser(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "postgres-test")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_TIMEOUT", "30s")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "POSTGRES_USER")
}

func TestLoad_MissingPassword(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "postgres-test")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_TIMEOUT", "30s")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "POSTGRES_PASSWORD")
}

func TestLoad_MissingDatabase(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "postgres-test")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_DB", "")
	t.Setenv("POSTGRES_TIMEOUT", "30s")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "POSTGRES_DB")
}

func TestLoad_MissingTimeout(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "postgres-test")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_TIMEOUT", "0s")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "POSTGRES_TIMEOUT")
}

func TestMustLoad_Success(t *testing.T) {
	setAllRequired(t)

	assert.NotPanics(t, func() {
		cfg := MustLoad()
		assert.Equal(t, "postgres-test", cfg.Host)
	})
}

func TestMustLoad_Panics(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "")

	assert.Panics(t, func() {
		MustLoad()
	})
}

func TestLoad_InvalidTimeout(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "postgres-test")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_TIMEOUT", "not-a-duration")

	_, err := Load()
	assert.Error(t, err)
}
