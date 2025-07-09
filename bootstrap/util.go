package bootstrap

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AppsSettings struct {
	DBSettings       DatabaseSettings
	WalletSettings   WalletSettings
	TelegramSettings TelegramSettings
}

type DatabaseSettings struct {
	Host string
	Name string
	Pass string
	User string
}

type WalletSettings struct {
	SecretKey string
	Password  string
}

type TelegramSettings struct {
	Endpoint       string
	Botname        string
	PersonalChatID int64
}

var (
	once   sync.Once
	config AppsSettings
)

func GetAppSettings() AppsSettings {
	once.Do(func() {
		config = getAppSettings()
	})
	return config
}

func getAppSettings() AppsSettings {
	dbHost := os.Getenv("DATABASE_HOST")
	if dbHost == "" {
		log.Fatal("DATABASE_HOST is not set")
	}

	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		log.Fatal("DATABASE_NAME is not set")
	}

	dbPass := os.Getenv("DATABASE_PASS")
	if dbPass == "" {
		log.Fatal("DATABASE_PASS is not set")
	}

	dbUser := os.Getenv("DATABASE_USER")
	if dbUser == "" {
		log.Fatal("DATABASE_USER is not set")
	}

	walletSecret := os.Getenv("APPS_SECRET_KEY")
	if walletSecret == "" {
		log.Fatal("APPS_SECRET_KEY is not set")
	}

	walletPassword := os.Getenv("APPS_PASSWORD")
	if walletPassword == "" {
		log.Fatal("APPS_PASSWORD is not set")
	}

	telegramEndpoint := os.Getenv("TELEGRAM_BOT_ENDPOINT")
	if telegramEndpoint == "" {
		log.Fatal("TELEGRAM_BOT_ENDPOINT is not set")
	}

	telegramBotname := os.Getenv("TELEGRAM_BOT_NAME")
	if telegramBotname == "" {
		log.Fatal("TELEGRAM_BOT_NAME is not set")
	}

	telegramPersonalChatIdStr := os.Getenv("TELEGRAM_PERSONAL_CHAT_ID")
	telegramPersonalChatId, err := strconv.ParseInt(telegramPersonalChatIdStr, 10, 64)
	if err != nil {
		log.Fatal("TELEGRAM_PERSONAL_CHAT_ID is not set")
	}

	return AppsSettings{
		DBSettings: DatabaseSettings{
			Host: dbHost,
			Name: dbName,
			Pass: dbPass,
			User: dbUser,
		},
		WalletSettings: WalletSettings{
			SecretKey: walletSecret,
			Password:  walletPassword,
		},
		TelegramSettings: TelegramSettings{
			Endpoint:       telegramEndpoint,
			Botname:        telegramBotname,
			PersonalChatID: telegramPersonalChatId,
		},
	}
}

func JwtCreateToken(walletSettings WalletSettings, userPassword string) string {

	if userPassword != walletSettings.Password {
		return ""
	}

	claims := jwt.RegisteredClaims{
		Subject:   userPassword,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
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

	parsedToken, err := jwt.ParseWithClaims(trimmed, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(walletSettings.SecretKey), nil
	})

	if err != nil {
		return false
	}

	if claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims); ok && parsedToken.Valid {
		return claims.Subject == walletSettings.Password
	}

	return false
}

func GetFrontendPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "ui", ".build")
}
