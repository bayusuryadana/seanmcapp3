package service

import (
	"seanmcapp/external"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func telegramUpdate(chatID int64, text string) external.TelegramUpdate {
	return external.TelegramUpdate{
		UpdateID: 1,
		Message: &external.TelegramResult{
			Chat: external.TelegramChat{ID: chatID, Type: "private"},
			Text: &text,
		},
	}
}

type fakeTelegramCommand struct {
	name     string
	message  external.TelegramResult
	args     []string
	runCount int
	err      error
}

func (c *fakeTelegramCommand) Name() string {
	return c.name
}

func (c *fakeTelegramCommand) Handle(message external.TelegramResult, args []string) error {
	c.message = message
	c.args = args
	c.runCount++
	return c.err
}

func TestTelegramCommandDispatcherRoutesRegisteredCommands(t *testing.T) {
	change := &fakeTelegramCommand{name: "change"}
	status := &fakeTelegramCommand{name: "status"}
	dispatcher := &TelegramCommandDispatcher{
		Botname:  "sean_bot",
		Commands: []TelegramCommand{change, status},
	}

	require.NoError(t, dispatcher.HandleUpdate(telegramUpdate(42, "change session csrf")))
	require.NoError(t, dispatcher.HandleUpdate(telegramUpdate(42, "/status now")))

	assert.Equal(t, 1, change.runCount)
	assert.Equal(t, []string{"session", "csrf"}, change.args)
	assert.Equal(t, int64(42), change.message.Chat.ID)
	assert.Equal(t, 1, status.runCount)
	assert.Equal(t, []string{"now"}, status.args)
}

func TestTelegramCommandDispatcherHandlesBotSuffix(t *testing.T) {
	command := &fakeTelegramCommand{name: "change"}
	dispatcher := &TelegramCommandDispatcher{
		Botname:  "@sean_bot",
		Commands: []TelegramCommand{command},
	}

	require.NoError(t, dispatcher.HandleUpdate(telegramUpdate(42, "/change@sean_bot session csrf")))
	require.NoError(t, dispatcher.HandleUpdate(telegramUpdate(42, "/change@another_bot session csrf")))

	assert.Equal(t, 1, command.runCount)
}

func TestTelegramCommandDispatcherIgnoresUnsupportedUpdates(t *testing.T) {
	command := &fakeTelegramCommand{name: "change"}
	dispatcher := &TelegramCommandDispatcher{
		Botname:  "sean_bot",
		Commands: []TelegramCommand{nil, command},
	}

	require.NoError(t, dispatcher.HandleUpdate(external.TelegramUpdate{}))
	require.NoError(t, dispatcher.HandleUpdate(telegramUpdate(42, "")))
	require.NoError(t, dispatcher.HandleUpdate(telegramUpdate(42, "unknown value")))

	assert.Zero(t, command.runCount)
}

func TestTelegramCommandDispatcherReturnsHandlerError(t *testing.T) {
	command := &fakeTelegramCommand{name: "change", err: assert.AnError}
	dispatcher := &TelegramCommandDispatcher{
		Botname:  "sean_bot",
		Commands: []TelegramCommand{command},
	}

	err := dispatcher.HandleUpdate(telegramUpdate(42, "change session csrf"))

	assert.ErrorIs(t, err, assert.AnError)
}
