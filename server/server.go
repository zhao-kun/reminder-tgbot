package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/zhao-kun/reminder-tgbot/client"
	"github.com/zhao-kun/reminder-tgbot/model"
	"github.com/zhao-kun/reminder-tgbot/repo"
	"github.com/zhao-kun/reminder-tgbot/task"
	"github.com/zhao-kun/reminder-tgbot/telegram"
	"github.com/zhao-kun/reminder-tgbot/util"
)

const (
	noneOpsCommand string = ""
	checkInCommand string = "/checkin"
	//
	contextTodayIsFestivalKey = "today_is_festival_key"
)

type (
	calendarResp struct {
		Data int `json:"data"`
		Code int `json:"code"`
	}

	response struct {
		Ok bool `json:"ok"`
	}

	//ProcessChatFunc is func which process dedicated request sent by tg
	ProcessChatFunc func(repo.Repo, model.Message) model.ReplyMessage

	// ResponseFunc wrap ProcesschatFunc
	ResponseFunc func(telegram.Client, repo.Repo, rest.ResponseWriter) error

	validateFunc func(model.Config, model.Message) (bool, string)
)

var (
	chatFuncs = map[string]ProcessChatFunc{
		checkInCommand: processCheckIn,
		noneOpsCommand: processNone,
	}
)

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
			resp.Text = fmt.Sprintf("Are you kidding me, you've already checkin at")
		} else {
			resp.Text = fmt.Sprintf("Sorry, checkin failed, Please check in manullay again")
		}
	}
	return resp
}

// TelegramServerHandle service checkin command sent from tgchannel
func TelegramServerHandle(c telegram.Client, r repo.Repo, w rest.ResponseWriter, req *rest.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rest.Error(w, "read request body error", http.StatusBadGateway)
		return
	}

	fmt.Printf(" request %s is comming body is [%s]\n", req.URL.Path[1:], body)
	var messages model.TgMessage
	err = json.Unmarshal(body, &messages)
	if err != nil {
		log.Printf("Can't unmarshal body [%s] to message", body)
		ok(w)
		return
	}

	respFunc, err := dispatch(r.Cfg(),
		[]model.TgMessage{messages},
		chatFuncs,
		validateSession,
		validateCheckInUser,
		validateCheckInTime)
	if err != nil {
		log.Printf("processChat error %s", err)
		ok(w)
		return
	}

	if respFunc == nil {
		ok(w)
		return
	}

	err = respFunc(c, r, w)
	if err != nil {
		log.Printf("respFunc run error %s", err)
		ok(w)
		return
	}

	return
}

func ok(w rest.ResponseWriter) {
	w.WriteJson(response{true})
	w.WriteHeader(http.StatusOK)
}

func dispatch(cfg model.Config, messages []model.TgMessage,
	chatFuncs map[string]ProcessChatFunc,
	funcs ...validateFunc) (ResponseFunc, error) {
	var pcf ProcessChatFunc = nil
	var currentMsg model.Message

	for _, message := range messages {
		if message.Message.MessageID > 0 &&
			isCommand(message.Message.Entities) &&
			message.Message.From.IsBot == false {
			switch message.Message.Text {
			case checkInCommand:
				pcf = getChatFuncs(checkInCommand, chatFuncs)
				currentMsg = message.Message
				break
			}
		}
	}

	if pcf != nil {
		return func(c telegram.Client, r repo.Repo, w rest.ResponseWriter) error {
			if pcf == nil {
				return nil
			}
			for _, validFunc := range funcs {
				valid, tips := validFunc(cfg, currentMsg)
				if !valid {
					reply := newReplyMessage(currentMsg.Chat.ID,
						currentMsg.MessageID, tips)
					return c.Reply(reply)
				}
			}
			reply := pcf(r, currentMsg)
			return c.Reply(reply)
		}, nil
	}
	return nil, nil
}

func getChineseFestivalCalendar(c telegram.Client, r repo.Repo, context task.Context) bool {
	if !isChinaTimeZoneNewDay() {
		return true
	}

	context[contextTodayIsFestivalKey] = 0
	config := r.Cfg()
	today := ""
	url := fmt.Sprintf("%s?date=%s", config.CNCalendarServiceEndpoint, today)
	resp, err := client.HandleRequest("GET", url, nil)
	if err != nil {
		log.Printf("request %s failed: %s", url, err)
		return true
	}

	var cal calendarResp
	err = json.Unmarshal(resp, &cal)
	if err != nil {
		log.Printf("unmarsh resp %+v failed: %s", resp, err)
	}

	context["today_is_fesetival"] = cal.Data
	log.Printf("today is [%d] day", cal.Data)
	return true
}

func reminder(c telegram.Client, r repo.Repo, context task.Context) bool {
	for _, u := range r.Cfg().CheckUesrs {
		if !r.IsUserNeedCheckIn(u) {
			continue
		}

		if !isWorkDay(context) ||
			!isRemindTime(time.Now(),
				r.Cfg().Remind.TimeRange.Begin,
				r.Cfg().Remind.TimeRange.End) {
			// no work day and no reminder time range
			continue
		}

		for _, chatID := range r.Cfg().Channels {
			message := model.BotMessage{
				ChatID: chatID,
				Text:   fmt.Sprintf("Hi @%s, you need to check in now", u),
			}
			if err := c.Message(message); err != nil {
				log.Printf("send message %+v to channel %d failed: %s", message, chatID, err)
			}
		}
	}
	return true
}

// StartAllBotTask start task which need be run by the bot
func StartAllBotTask(c telegram.Client, r repo.Repo) error {
	context := task.NewContext()
	context["Calendar"] = map[string]int{}

	calendarTask, err := task.New("get_chinese_festival_task", "2m",
		task.WrapWithRepoAndTelegramClient(c, r, context, getChineseFestivalCalendar))
	if err != nil {
		return fmt.Errorf("create calendarTask error: %s", err)
	}

	remindTask, err := task.New("remind_task", r.Cfg().Remind.RemindInterval,
		task.WrapWithRepoAndTelegramClient(c, r, context, reminder))
	if err != nil {
		return fmt.Errorf("create remindTask error: %s", err)
	}

	registry := task.NewTaskRegistry()

	err = registry.AddTask(calendarTask)
	if err != nil {
		return fmt.Errorf("Add %s task error: %s", calendarTask.GetName(), err)
	}
	err = registry.AddTask(remindTask)
	if err != nil {
		return fmt.Errorf("Add %s task error: %s", remindTask.GetName(), err)
	}
	registry.StartAllTask()
	return nil
}

func isWorkDay(context task.Context) bool {
	value := context[contextTodayIsFestivalKey]
	workday, ok := value.(int)
	if !ok {
		log.Printf("WARN: the context value[%+v] of [%s] is not int type",
			value, contextTodayIsFestivalKey)
		workday = 0
	}
	return workday <= 0
}

func isRemindTime(t time.Time, begin, end string) bool {
	return timeInRange(t, begin, end)
}

func timeInRange(t time.Time, begin, end string) bool {
	y, m, d := t.Date()
	beginTimeStr := fmt.Sprintf("%04d-%02d-%02dT%s", y, m, d, begin)
	endTimeStr := fmt.Sprintf("%04d-%02d-%02dT%s", y, m, d, end)

	beginTime, err := time.Parse(time.RFC3339, beginTimeStr)
	if err != nil {
		log.Printf("convert %s to time error %s", beginTimeStr, err)
		return false
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		log.Printf("convert %s to time error %s", endTime, err)
		return false
	}

	if t.After(beginTime) && t.Before(endTime) {
		return true
	}
	return false
}

func isNeedCheckIn(checkUsrs []string, usrname string) bool {
	return util.StrInSlice(usrname, checkUsrs)
}

func isCommand(ents []model.Entity) (yes bool) {
	for _, ent := range ents {
		if ent.Type == "bot_command" {
			return true
		}
	}
	return
}

func getChatFuncs(funcName string, funcs map[string]ProcessChatFunc) ProcessChatFunc {
	return funcs[funcName]
}

func isSessionAllowToCheckIn(cfg model.Config, currentSession int) bool {
	for _, c := range cfg.Channels {
		if c == currentSession {
			return true
		}
	}
	return false
}

func newReplyMessage(chatID, replyID int, text string) model.ReplyMessage {
	/*
		fmt.Sprintf("OK! you are checkin @%s", msg.From.Username),
	*/
	return model.ReplyMessage{
		BotMessage: model.BotMessage{
			ChatID: chatID,
			Text:   text,
		},
		ReplyToMessageID: replyID,
	}
}

func isChinaTimeZoneNewDay() bool {
	chinaNow := util.GetChinaTimeNow()
	if timeInRange(chinaNow, "00:00:00+800", "00:11:00+800") {
		return true
	}
	return false
}
