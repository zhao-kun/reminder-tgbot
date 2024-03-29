package server

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/zhao-kun/reminder-tgbot/client"
	"github.com/zhao-kun/reminder-tgbot/model"
	"github.com/zhao-kun/reminder-tgbot/repo"
	"github.com/zhao-kun/reminder-tgbot/task"
	"github.com/zhao-kun/reminder-tgbot/telegram"
	"github.com/zhao-kun/reminder-tgbot/util"
)

// wrapWithRepoAndTelegramClient wrap function with model.Config and
// telegram.Client function to a TaskCallbackFunc
func wrapWithRepoAndTelegramClient(tgClient telegram.Client, r repo.Repo,
	c task.Context, f func(telegram.Client, repo.Repo, task.Context) bool) task.CallbackFunc {
	return func() bool {
		return f(tgClient, r, c)
	}
}

func getChineseFestivalCalendar(c telegram.Client, r repo.Repo, context task.Context) bool {
	if !isChinaTimeZoneNewDay() {
		return true
	}
	context[contextTodayIsFestivalKey] = 0
	context[contextTodayIsFestivalKey] = todayIsFestival(r.Cfg().CNCalendarServiceEndpoint)
	log.Printf("today is [%d] day", context[contextTodayIsFestivalKey])
	return true
}

func todayIsFestival(calendarServiceEndpoint string) int {
	today := util.GetDate(util.GetChinaTimeNow())
	url := fmt.Sprintf("%s?date=%s", calendarServiceEndpoint, today)
	resp, err := client.HandleRequest("GET", url, nil)
	if err != nil {
		log.Printf("request %s failed: %s", url, err)
		return 0
	}

	var cal calendarResp
	err = json.Unmarshal(resp, &cal)
	if err != nil {
		log.Printf("unmarsh resp %+v failed: %s", resp, err)
		return 0
	}

	return cal.Data
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

	context[contextTodayIsFestivalKey] = todayIsFestival(r.Cfg().CNCalendarServiceEndpoint)
	log.Printf("Today is %d", context[contextTodayIsFestivalKey])

	calendarTask, err := task.New("get_chinese_festival_task", "2m",
		wrapWithRepoAndTelegramClient(c, r, context, getChineseFestivalCalendar))
	if err != nil {
		return fmt.Errorf("create calendarTask error: %s", err)
	}

	remindTask, err := task.New("remind_task", r.Cfg().Remind.RemindInterval,
		wrapWithRepoAndTelegramClient(c, r, context, reminder))
	if err != nil {
		return fmt.Errorf("create remindTask error: %s", err)
	}

	registry := task.NewTaskRegistry()

	err = registry.AddTask(calendarTask)
	if err != nil {
		return fmt.Errorf("Add %s task error: %s", calendarTask.Name(), err)
	}
	err = registry.AddTask(remindTask)
	if err != nil {
		return fmt.Errorf("Add %s task error: %s", remindTask.Name(), err)
	}
	registry.StartAllTask()
	return nil
}
