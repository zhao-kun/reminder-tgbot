package telegram

import (
	"encoding/json"
	"fmt"

	httpclient "github.com/zhao-kun/reminder-tgbot/client"
	"github.com/zhao-kun/reminder-tgbot/model"
)

type (
	// Client represent a telegram client which send requst to specific
	// group or channel
	Client interface {
		Reply(message model.ReplyMessage) error
		Message(message model.BotMessage) error
	}

	client struct {
		cfg model.Config
	}
)

var _ Client = client{}

func (c client) Reply(message model.ReplyMessage) error {
	if message.ReplyToMessageID <= 0 {
		return fmt.Errorf("Reply message should refer to a origin message")
	}
	return sendMessage(c.cfg, message)
}

func (c client) Message(message model.BotMessage) error {
	return sendMessage(c.cfg, message)
}

func sendMessage(cfg model.Config, message interface{}) error {
	if text, ok := message.(model.Text); ok {
		if text.TextInfo() == "" {
			return fmt.Errorf("Message should contains text")
		}
	}
	request, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("Marsh json of resp %+v error %s", message, err)
	}

	_, err = httpclient.HandleRequest("POST",
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.TgbotToken), request)
	if err != nil {
		return err
	}
	return nil
}

// NewClient return a telegram Client object
func NewClient(cfg model.Config) Client {
	return client{cfg}
}
