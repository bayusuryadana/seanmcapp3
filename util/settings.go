package util

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type AppsSettings struct {
	DBSettings       DatabaseSettings
	WalletSettings   WalletSettings
	TelegramSettings TelegramSettings
	IGSettings       IGSettings
}

type IGSettings struct {
	SessionID string
	CSRFToken string
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
	GroupChatID    int64
}

var (
	once    sync.Once
	config  AppsSettings
	fatalFn = log.Fatal
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
		fatalFn("DATABASE_HOST is not set")
	}

	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		fatalFn("DATABASE_NAME is not set")
	}

	dbPass := os.Getenv("DATABASE_PASS")
	if dbPass == "" {
		fatalFn("DATABASE_PASS is not set")
	}

	dbUser := os.Getenv("DATABASE_USER")
	if dbUser == "" {
		fatalFn("DATABASE_USER is not set")
	}

	walletSecret := os.Getenv("APPS_SECRET_KEY")
	if walletSecret == "" {
		fatalFn("APPS_SECRET_KEY is not set")
	}

	walletPassword := os.Getenv("APPS_PASSWORD")
	if walletPassword == "" {
		fatalFn("APPS_PASSWORD is not set")
	}

	telegramEndpoint := os.Getenv("TELEGRAM_BOT_ENDPOINT")
	if telegramEndpoint == "" {
		fatalFn("TELEGRAM_BOT_ENDPOINT is not set")
	}

	telegramBotname := os.Getenv("TELEGRAM_BOT_NAME")
	if telegramBotname == "" {
		fatalFn("TELEGRAM_BOT_NAME is not set")
	}

	telegramPersonalChatIdStr := os.Getenv("TELEGRAM_PERSONAL_CHAT_ID")
	telegramPersonalChatId, err := strconv.ParseInt(telegramPersonalChatIdStr, 10, 64)
	if err != nil {
		fatalFn("TELEGRAM_PERSONAL_CHAT_ID is not set")
	}

	telegramGroupChatIdStr := os.Getenv("TELEGRAM_GROUP_CHAT_ID")
	telegramGroupChatId, err := strconv.ParseInt(telegramGroupChatIdStr, 10, 64)
	if err != nil {
		fatalFn("TELEGRAM_GROUP_CHAT_ID is not set")
	}

	igSessionID := os.Getenv("IG_SESSION_ID")
	if igSessionID == "" {
		fatalFn("IG_SESSION_ID is not set")
	}

	igCSRFToken := os.Getenv("IG_CSRF_TOKEN")
	if igCSRFToken == "" {
		fatalFn("IG_CSRF_TOKEN is not set")
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
			GroupChatID:    telegramGroupChatId,
		},
		IGSettings: IGSettings{
			SessionID: igSessionID,
			CSRFToken: igCSRFToken,
		},
	}
}

func GetFrontendPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "ui", ".build")
}
