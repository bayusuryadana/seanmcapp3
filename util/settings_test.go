package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAppSettings(t *testing.T) {
	origEnv := map[string]string{
		"DATABASE_HOST":             "db-host",
		"DATABASE_NAME":             "db-name",
		"DATABASE_PASS":             "db-pass",
		"DATABASE_USER":             "db-user",
		"APPS_SECRET_KEY":           "secret-key",
		"APPS_PASSWORD":             "password",
		"TELEGRAM_BOT_ENDPOINT":     "https://api.telegram.org/bot",
		"TELEGRAM_BOT_NAME":         "botname",
		"TELEGRAM_PERSONAL_CHAT_ID": "123",
		"TELEGRAM_GROUP_CHAT_ID":    "456",
		"IG_SESSION_ID":             "sess",
		"IG_CSRF_TOKEN":             "csrf",
	}
	for key, value := range origEnv {
		require.NoError(t, os.Setenv(key, value))
	}
	defer func() {
		for key := range origEnv {
			_ = os.Unsetenv(key)
		}
		once = sync.Once{}
		fatalFn = log.Fatal
	}()

	once = sync.Once{}
	fatalFn = func(v ...any) { panic(fmt.Sprint(v...)) }
	settings := GetAppSettings()

	assert.Equal(t, "db-host", settings.DBSettings.Host)
	assert.Equal(t, "db-name", settings.DBSettings.Name)
	assert.Equal(t, "db-pass", settings.DBSettings.Pass)
	assert.Equal(t, "db-user", settings.DBSettings.User)
	assert.Equal(t, "secret-key", settings.WalletSettings.SecretKey)
	assert.Equal(t, "password", settings.WalletSettings.Password)
	assert.Equal(t, "https://api.telegram.org/bot", settings.TelegramSettings.Endpoint)
	assert.Equal(t, "botname", settings.TelegramSettings.Botname)
	assert.Equal(t, int64(123), settings.TelegramSettings.PersonalChatID)
	assert.Equal(t, int64(456), settings.TelegramSettings.GroupChatID)
	assert.Equal(t, "sess", settings.IGSettings.SessionID)
	assert.Equal(t, "csrf", settings.IGSettings.CSRFToken)
}

func TestGetAppSettingsMissingEnvPanics(t *testing.T) {
	for _, key := range []string{"DATABASE_HOST", "DATABASE_NAME", "DATABASE_PASS", "DATABASE_USER", "APPS_SECRET_KEY", "APPS_PASSWORD", "TELEGRAM_BOT_ENDPOINT", "TELEGRAM_BOT_NAME", "TELEGRAM_PERSONAL_CHAT_ID", "TELEGRAM_GROUP_CHAT_ID", "IG_SESSION_ID", "IG_CSRF_TOKEN"} {
		_ = os.Unsetenv(key)
	}
	defer func() {
		fatalFn = log.Fatal
		once = sync.Once{}
	}()

	once = sync.Once{}
	fatalFn = func(v ...any) { panic(fmt.Sprint(v...)) }

	assert.Panics(t, func() { GetAppSettings() })
}

func TestGetFrontendPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	got := GetFrontendPath()
	assert.Equal(t, filepath.Join(wd, "ui", ".build"), got)
}
