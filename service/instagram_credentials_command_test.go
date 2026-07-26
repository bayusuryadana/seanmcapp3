package service

import (
	"errors"
	"seanmcapp/external"
	"seanmcapp/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func telegramResultForChat(chatID int64) external.TelegramResult {
	return external.TelegramResult{
		Chat: external.TelegramChat{ID: chatID, Type: "private"},
	}
}

func TestInstagramCredentialsCommand(t *testing.T) {
	configRepo := &fakeConfigRepo{}
	telegramClient := &fakeTelegramClient{}
	command := &InstagramCredentialsCommand{
		ConfigRepo:     configRepo,
		TelegramClient: telegramClient,
		PersonalChatID: 42,
	}

	assert.Equal(t, "change", command.Name())
	require.NoError(t, command.Handle(telegramResultForChat(42), []string{"session-value", "csrf-value"}))

	require.Len(t, configRepo.setCalls, 1)
	assert.Equal(t, map[string]string{
		repository.ConfigKeyIGSessionID: "session-value",
		repository.ConfigKeyIGCSRFToken: "csrf-value",
	}, configRepo.setCalls[0])
	require.Len(t, telegramClient.messages, 1)
	assert.Equal(t, int64(42), telegramClient.messages[0].chatID)
	assert.Equal(t, "✅ Instagram credentials updated.", telegramClient.messages[0].text)
	assert.NotContains(t, telegramClient.messages[0].text, "session-value")
	assert.NotContains(t, telegramClient.messages[0].text, "csrf-value")
}

func TestInstagramCredentialsCommandUsage(t *testing.T) {
	configRepo := &fakeConfigRepo{}
	telegramClient := &fakeTelegramClient{}
	command := &InstagramCredentialsCommand{
		ConfigRepo:     configRepo,
		TelegramClient: telegramClient,
		PersonalChatID: 42,
	}

	require.NoError(t, command.Handle(telegramResultForChat(42), []string{"only-session"}))

	assert.Empty(t, configRepo.setCalls)
	require.Len(t, telegramClient.messages, 1)
	assert.Equal(t, instagramChangeUsage, telegramClient.messages[0].text)
}

func TestInstagramCredentialsCommandIgnoresUnauthorizedChat(t *testing.T) {
	configRepo := &fakeConfigRepo{}
	telegramClient := &fakeTelegramClient{}
	command := &InstagramCredentialsCommand{
		ConfigRepo:     configRepo,
		TelegramClient: telegramClient,
		PersonalChatID: 42,
	}

	require.NoError(t, command.Handle(telegramResultForChat(99), []string{"session", "csrf"}))

	assert.Empty(t, configRepo.setCalls)
	assert.Empty(t, telegramClient.messages)
}

func TestInstagramCredentialsCommandReportsStorageFailure(t *testing.T) {
	configRepo := &fakeConfigRepo{setErr: errors.New("database unavailable")}
	telegramClient := &fakeTelegramClient{}
	command := &InstagramCredentialsCommand{
		ConfigRepo:     configRepo,
		TelegramClient: telegramClient,
		PersonalChatID: 42,
	}

	err := command.Handle(telegramResultForChat(42), []string{"session", "csrf"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database unavailable")
	require.Len(t, telegramClient.messages, 1)
	assert.Contains(t, telegramClient.messages[0].text, "Could not update")
	assert.NotContains(t, telegramClient.messages[0].text, "session")
	assert.NotContains(t, telegramClient.messages[0].text, "csrf")
}
