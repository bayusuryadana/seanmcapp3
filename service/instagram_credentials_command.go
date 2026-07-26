package service

import (
	"fmt"
	"seanmcapp/external"
	"seanmcapp/repository"
)

const instagramChangeUsage = "Usage: `change <session_id> <csrf_token>`"

type InstagramCredentialsCommand struct {
	ConfigRepo     repository.ConfigRepo
	TelegramClient external.TelegramClient
	PersonalChatID int64
}

func (c *InstagramCredentialsCommand) Name() string {
	return "change"
}

func (c *InstagramCredentialsCommand) Handle(message external.TelegramResult, args []string) error {
	if message.Chat.ID != c.PersonalChatID {
		return nil
	}
	if len(args) != 2 {
		_, err := c.TelegramClient.SendMessage(message.Chat.ID, instagramChangeUsage)
		return err
	}

	err := c.ConfigRepo.SetValues(map[string]string{
		repository.ConfigKeyIGSessionID: args[0],
		repository.ConfigKeyIGCSRFToken: args[1],
	})
	if err != nil {
		_, _ = c.TelegramClient.SendMessage(message.Chat.ID, "❌ Could not update Instagram credentials. Please try again.")
		return fmt.Errorf("update Instagram credentials: %w", err)
	}

	_, err = c.TelegramClient.SendMessage(message.Chat.ID, "✅ Instagram credentials updated.")
	return err
}
