package service

import (
	"seanmcapp/external"
	"strings"
)

type TelegramUpdateHandler interface {
	HandleUpdate(update external.TelegramUpdate) error
}

type TelegramCommand interface {
	Name() string
	Handle(message external.TelegramResult, args []string) error
}

type TelegramCommandDispatcher struct {
	Botname  string
	Commands []TelegramCommand
}

func (d *TelegramCommandDispatcher) HandleUpdate(update external.TelegramUpdate) error {
	if update.Message == nil || update.Message.Text == nil {
		return nil
	}

	name, args, ok := d.parseCommand(*update.Message.Text)
	if !ok {
		return nil
	}

	for _, command := range d.Commands {
		if command != nil && strings.EqualFold(command.Name(), name) {
			return command.Handle(*update.Message, args)
		}
	}
	return nil
}

func (d *TelegramCommandDispatcher) parseCommand(text string) (string, []string, bool) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "", nil, false
	}

	commandToken := strings.TrimPrefix(strings.ToLower(fields[0]), "/")
	parts := strings.SplitN(commandToken, "@", 2)
	botname := strings.TrimPrefix(strings.ToLower(d.Botname), "@")
	if len(parts) == 2 && parts[1] != botname {
		return "", nil, false
	}
	return parts[0], fields[1:], true
}
