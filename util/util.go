package util

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const walletSubject = "wallet-user"

func JwtCreateToken(walletSettings WalletSettings, userPassword string) string {

	if userPassword != walletSettings.Password {
		return ""
	}

	claims := jwt.RegisteredClaims{
		Subject:   walletSubject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(walletSettings.SecretKey))
	if err != nil {
		return ""
	}
	return signedToken
}

func JwtValidateToken(walletSettings WalletSettings, token string) bool {
	trimmed := strings.TrimPrefix(token, "Bearer ")

	parsedToken, err := jwt.ParseWithClaims(
		trimmed,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(walletSettings.SecretKey), nil
		},
		jwt.WithValidMethods([]string{"HS256"}), // pin the algorithm; reject anything else
	)

	if err != nil {
		return false
	}

	if claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims); ok && parsedToken.Valid {
		return claims.Subject == walletSubject
	}

	return false
}

