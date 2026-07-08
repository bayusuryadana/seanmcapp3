package util

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSettings = WalletSettings{SecretKey: "test-secret", Password: "correct-password"}

func TestJwtCreateToken(t *testing.T) {
	t.Run("correct password yields a token", func(t *testing.T) {
		token := JwtCreateToken(testSettings, "correct-password")
		assert.NotEmpty(t, token)
	})

	t.Run("wrong password yields empty", func(t *testing.T) {
		assert.Empty(t, JwtCreateToken(testSettings, "nope"))
	})
}

func TestJwtValidateToken(t *testing.T) {
	valid := JwtCreateToken(testSettings, "correct-password")
	require.NotEmpty(t, valid)

	t.Run("valid token with Bearer prefix", func(t *testing.T) {
		assert.True(t, JwtValidateToken(testSettings, "Bearer "+valid))
	})

	t.Run("valid token without prefix", func(t *testing.T) {
		assert.True(t, JwtValidateToken(testSettings, valid))
	})

	t.Run("malformed token", func(t *testing.T) {
		assert.False(t, JwtValidateToken(testSettings, "Bearer not.a.jwt"))
	})

	t.Run("wrong secret", func(t *testing.T) {
		assert.False(t, JwtValidateToken(WalletSettings{SecretKey: "other"}, "Bearer "+valid))
	})

	t.Run("wrong subject is rejected", func(t *testing.T) {
		claims := jwt.RegisteredClaims{
			Subject:   "someone-else",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}
		signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSettings.SecretKey))
		require.NoError(t, err)
		assert.False(t, JwtValidateToken(testSettings, "Bearer "+signed))
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		claims := jwt.RegisteredClaims{
			Subject:   walletSubject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		}
		signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSettings.SecretKey))
		require.NoError(t, err)
		assert.False(t, JwtValidateToken(testSettings, "Bearer "+signed))
	})
}
