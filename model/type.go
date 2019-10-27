package model

type (
	From struct {
		ID           int    `json:"id"`
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Username     string `json:"username"`
		LanguageCode string `json:"language_code"`
	}
	Chat struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Type      string `json:"type"`
	}

	Entity struct {
		Offset int    `json:"offset"`
		Length int    `json:"length"`
		Type   string `json:"type"`
	}

	// TgMessage represent message recieved from Telegram
	TgMessage struct {
		UpdateID int     `json:"update_id"`
		Message  Message `json:"message"`
	}

	// Message contains detail message information sent from Telegram
	Message struct {
		MessageID int      `json:"message_id"`
		From      From     `json:"from"`
		Chat      Chat     `json:"chat"`
		Entities  []Entity `json:"entities"`
		Date      int      `json:"date"`
		Text      string   `json:"text"`
	}

	// BotMessage represent message send by bot
	BotMessage struct {
		//
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}

	// ReplyMessage represent message sent by bot
	ReplyMessage struct {
		BotMessage
		ReplyToMessageID int `json:"reply_to_message_id"`
	}

	// TimeRange contain a period of time
	TimeRange struct {
		Begin string `json:"begin"`
		End   string `json:"end"`
	}

	// Remind contains reminding configuration information
	Remind struct {
		RemindInterval string    `json:"remind_interval"`
		TimeRange      TimeRange `json:"time_range"`
	}
	// Config represent global configuration
	Config struct {
		TgbotToken      string   `json:"tgbot_token"`
		ListenAddr      string   `json:"listen_addr"`
		CheckUesrs      []string `json:"check_users"`
		WebhookEndpoint string   `json:"webhook_endpoint"`
		Channels        []int64  `json:"channels"`
		Remind          Remind   `json:"remind"`
		//
		CNCalendarServiceEndpoint string `json:"cn_calendar_service_endpoint"`
	}
)

// TextInfo tell user a BotMessage is a common Text interface
func (b BotMessage) TextInfo() string {
	return b.Text
}
