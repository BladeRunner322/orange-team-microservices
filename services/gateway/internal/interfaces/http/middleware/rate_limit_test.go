package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== ParseTrustedProxies =====

func TestParseTrustedProxies(t *testing.T) {
	t.Run("пустой список", func(t *testing.T) {
		nets, err := ParseTrustedProxies(nil)
		require.NoError(t, err)
		assert.Empty(t, nets)
	})

	t.Run("один валидный CIDR", func(t *testing.T) {
		nets, err := ParseTrustedProxies([]string{"10.0.0.0/8"})
		require.NoError(t, err)
		require.Len(t, nets, 1)
		assert.True(t, nets[0].Contains(net.ParseIP("10.0.0.5")))
	})

	t.Run("несколько CIDR", func(t *testing.T) {
		nets, err := ParseTrustedProxies([]string{"10.0.0.0/8", "192.168.0.0/16"})
		require.NoError(t, err)
		assert.Len(t, nets, 2)
	})

	t.Run("пустые строки и пробелы игнорируются", func(t *testing.T) {
		nets, err := ParseTrustedProxies([]string{"", "  ", "10.0.0.0/8"})
		require.NoError(t, err)
		assert.Len(t, nets, 1)
	})

	t.Run("невалидный CIDR — ошибка", func(t *testing.T) {
		_, err := ParseTrustedProxies([]string{"not-a-cidr"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not-a-cidr")
	})
}

// ===== isTrusted =====

func TestIsTrusted(t *testing.T) {
	trusted := mustParseCIDRs(t, "10.0.0.0/8", "192.168.0.0/16")

	t.Run("пустой список — false", func(t *testing.T) {
		assert.False(t, isTrusted("10.0.0.5", nil))
	})

	t.Run("IP входит в одну из сетей — true", func(t *testing.T) {
		assert.True(t, isTrusted("10.0.0.5", trusted))
		assert.True(t, isTrusted("192.168.1.100", trusted))
	})

	t.Run("IP не входит ни в одну сеть — false", func(t *testing.T) {
		assert.False(t, isTrusted("8.8.8.8", trusted))
	})

	t.Run("невалидный IP — false", func(t *testing.T) {
		assert.False(t, isTrusted("not-an-ip", trusted))
	})
}

// ===== clientIP =====

func TestClientIP(t *testing.T) {
	trusted := mustParseCIDRs(t, "10.0.0.0/8")

	t.Run("trusted пуст, XFF есть — игнорируем XFF", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "8.8.8.8:12345"
		r.Header.Set("X-Forwarded-For", "1.2.3.4")

		ip := clientIP(r, nil)

		assert.Equal(t, "8.8.8.8", ip, "XFF должен игнорироваться без trusted proxies")
	})

	t.Run("trusted пуст, XFF нет — RemoteAddr", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "8.8.8.8:12345"

		ip := clientIP(r, nil)

		assert.Equal(t, "8.8.8.8", ip)
	})

	t.Run("RemoteAddr trusted, XFF есть — берём XFF", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "10.0.0.5:12345"
		r.Header.Set("X-Forwarded-For", "1.2.3.4")

		ip := clientIP(r, trusted)

		assert.Equal(t, "1.2.3.4", ip)
	})

	t.Run("RemoteAddr trusted, XFF нет — RemoteAddr", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "10.0.0.5:12345"

		ip := clientIP(r, trusted)

		assert.Equal(t, "10.0.0.5", ip)
	})

	t.Run("RemoteAddr не trusted, XFF есть — игнорируем XFF", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "8.8.8.8:12345"
		r.Header.Set("X-Forwarded-For", "1.2.3.4")

		ip := clientIP(r, trusted)

		assert.Equal(t, "8.8.8.8", ip, "XFF от не-trusted источника должен игнорироваться")
	})

	t.Run("XFF с несколькими IP — берём первый не-trusted справа", func(t *testing.T) {
		// XFF = "клиент, предыдущий_прокси, текущий_прокси"
		// Обходим справа: 10.0.0.5 trusted → пропускаем, 5.6.7.8 не trusted → это клиент.
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "10.0.0.5:12345"
		r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8, 10.0.0.5")

		ip := clientIP(r, trusted)

		assert.Equal(t, "5.6.7.8", ip)
	})

	t.Run("XFF с пробелами — trim", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "10.0.0.5:12345"
		r.Header.Set("X-Forwarded-For", "  1.2.3.4  , 5.6.7.8")

		ip := clientIP(r, trusted)

		assert.Equal(t, "5.6.7.8", ip)
	})

	t.Run("RemoteAddr без порта — возвращается как есть", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "8.8.8.8"

		ip := clientIP(r, nil)

		assert.Equal(t, "8.8.8.8", ip)
	})

	t.Run("Caddy добавляет свой IP в XFF — берём реального клиента", func(t *testing.T) {
		// Клиент → Caddy (172.18.0.11). Caddy дописал свой IP справа.
		// Реальный клиент — последний не-trusted справа.
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", "203.0.113.42, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "203.0.113.42", ip)
	})

	t.Run("клиент подделал XFF — берём реальный IP от Caddy", func(t *testing.T) {
		// Злоумышленник послал XFF: 1.2.3.4 (подделка).
		// Caddy добавил реальный IP клиента справа: 203.0.113.42.
		// XFF на входе Gateway: "1.2.3.4, 203.0.113.42, 172.18.0.11".
		// Обходим справа: 172.18.0.11 trusted, 203.0.113.42 не trusted → реальный клиент.
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", "1.2.3.4, 203.0.113.42, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "203.0.113.42", ip, "подделка слева должна игнорироваться")
	})

	t.Run("все элементы XFF trusted — fall back на RemoteAddr", func(t *testing.T) {
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", "172.18.0.5, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "172.18.0.11", ip)
	})

	t.Run("мусор в XFF пропускается", func(t *testing.T) {
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", "not-an-ip, 203.0.113.42, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "203.0.113.42", ip)
	})

	t.Run("пустые элементы XFF пропускаются", func(t *testing.T) {
		trustedDocker := mustParseCIDRs(t, "172.18.0.0/16")

		r := httptest.NewRequest(http.MethodPost, "/login", nil)
		r.RemoteAddr = "172.18.0.11:54321"
		r.Header.Set("X-Forwarded-For", ", , 203.0.113.42, 172.18.0.11")

		ip := clientIP(r, trustedDocker)

		assert.Equal(t, "203.0.113.42", ip)
	})
}

// mustParseCIDRs — хелпер для тестов: парсит CIDR или падает.
func mustParseCIDRs(t *testing.T, cidrs ...string) []*net.IPNet {
	t.Helper()
	nets, err := ParseTrustedProxies(cidrs)
	require.NoError(t, err)
	return nets
}
