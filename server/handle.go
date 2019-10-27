package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
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

	//processCommandFunc is func which process dedicated command sent by tg
	processCommandFunc func(repo.Repo, model.Message) model.ReplyMessage

	// commandFunc wrap ProcesschatFunc
	commandFunc func(telegram.Client, repo.Repo, rest.ResponseWriter) error

	validateFunc func(model.Config, model.Message) (bool, string)
)

var (
	chatFuncs = map[string]processCommandFunc{
		checkInCommand: processCheckIn,
		noneOpsCommand: processNone,
	}
)

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

func getChatFuncs(funcName string, funcs map[string]processCommandFunc) processCommandFunc {
	return funcs[funcName]
}

func isSessionAllowToCheckIn(cfg model.Config, currentSession int64) bool {
	for _, c := range cfg.Channels {
		if c == currentSession {
			return true
		}
	}
	return false
}

func isChinaTimeZoneNewDay() bool {
	chinaNow := util.GetChinaTimeNow()
	if timeInRange(chinaNow, "00:00:00+08:00", "00:11:00+08:00") {
		return true
	}
	return false
}

func dispatch(cfg model.Config, messages []model.TgMessage,
	chatFuncs map[string]processCommandFunc,
	validFuncs ...validateFunc) (commandFunc, error) {
	var pcf processCommandFunc = nil
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
			for _, validFunc := range validFuncs {
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

// TelegramServerHandle served `/checkin` command sent by user from tgchannel
func TelegramServerHandle(c telegram.Client, r repo.Repo, w rest.ResponseWriter, req *rest.Request) {

	ok := func() {
		w.WriteJson(response{true})
		w.WriteHeader(http.StatusOK)
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rest.Error(w, "read request body error", http.StatusBadGateway)
		return
	}

	log.Printf("Request %s is comming body is:\n%s\n", req.URL.Path[1:], body)
	var messages model.TgMessage
	err = json.Unmarshal(body, &messages)
	if err != nil {
		log.Printf("Can't unmarshal body [%s] to message", body)
		ok()
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
		ok()
		return
	}

	if respFunc == nil {
		ok()
		return
	}

	err = respFunc(c, r, w)
	if err != nil {
		log.Printf("respFunc run error %s", err)
		ok()
		return
	}

	return
}
