package server

import (
	"fmt"
	"log"
	"time"

	"github.com/zhao-kun/reminder-tgbot/model"
	"github.com/zhao-kun/reminder-tgbot/repo"
)

func newReplyMessage(chatID, replyID int, text string) model.ReplyMessage {
	return model.ReplyMessage{
		BotMessage: model.BotMessage{
			ChatID: chatID,
			Text:   text,
		},
		ReplyToMessageID: replyID,
	}
}

func validateCheckInUser(cfg model.Config, message model.Message) (valid bool, tips string) {
	return isNeedCheckIn(cfg.CheckUesrs, message.From.Username),
		fmt.Sprintf("Hi %s, you are good guy, no need to check every day",
			message.From.Username)
}

func validateSession(cfg model.Config, message model.Message) (valid bool, tips string) {
	return isSessionAllowToCheckIn(cfg, message.Chat.ID),
		"Sorry, current session isn't allowed to check in"
}

func validateCheckInTime(cfg model.Config, message model.Message) (valid bool, tips string) {
	checkInTime := time.Unix(int64(message.Date), 0)
	return isRemindTime(
			checkInTime, cfg.Remind.TimeRange.Begin, cfg.Remind.TimeRange.End),
		fmt.Sprintf("Sorry, please check in at %s - %s", cfg.Remind.TimeRange.Begin,
			cfg.Remind.TimeRange.End)

}

func processNone(repo repo.Repo, msg model.Message) model.ReplyMessage {
	return model.ReplyMessage{}
}

func processCheckIn(r repo.Repo, msg model.Message) model.ReplyMessage {
	resp := newReplyMessage(msg.Chat.ID, msg.MessageID, fmt.Sprintf("OK! you are checkin @%s", msg.From.Username))
	err := r.CheckIn(msg)
	if err != nil {
		log.Printf("%s checkin at %d failed:%s", msg.Chat.Username, msg.Date, err)
		if err == repo.ErrAlreadyCheckedIn {
			resp.Text = fmt.Sprintf("Are you kidding me, you've already checked in.")
		} else {
			resp.Text = fmt.Sprintf("Sorry, checked in failed, please contact tgbot author.")
		}
	}
	return resp
}
