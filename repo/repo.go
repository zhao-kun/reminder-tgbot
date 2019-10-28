package repo

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/zhao-kun/reminder-tgbot/model"
	"github.com/zhao-kun/reminder-tgbot/util"
)

var (
	// ErrAlreadyCheckedIn represent you have already checked
	ErrAlreadyCheckedIn = fmt.Errorf("Have checked in already")
)

// Repo is a interface which operation check history
type Repo interface {
	model.Cfg
	// CheckIn record a information of checking in according to message
	CheckIn(model.Message) error
	// IsUserNeedCheckIn jude whether the `user` need to check in today
	IsUserNeedCheckIn(user string) bool
}

type repo struct {
	cfg model.Config
}

var _ Repo = repo{}

func (r repo) Cfg() model.Config {
	return r.cfg
}

// CheckIn executed check in by some one
func (r repo) CheckIn(message model.Message) error {
	checkTime := util.GetChinaTimeFromUnix(int64(message.Date))
	return checkIn(checkTime, message.From.Username)
}

func (r repo) IsUserNeedCheckIn(user string) bool {
	now := util.GetChinaTimeNow()
	file, _ := checkInFilePath(now, user)
	if util.IsFileExist(file) {
		return false
	}
	return true
}

func checkIn(checkTime time.Time, user string) error {
	path, file := checkInFilePath(checkTime, user)
	log.Printf("file is %s", file)
	if util.IsFileExist(file) {
		return ErrAlreadyCheckedIn
	}

	os.MkdirAll(path, 0755)
	h, err := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer h.Close()
	_, err = h.WriteString(fmt.Sprintf("%s checkin at %+v", user, checkTime))
	return err

}

// New return a Repo interface
func New(cfg model.Config) Repo {
	return repo{cfg}
}

func checkInFilePath(checkTime time.Time, user string) (checkinFilePath string, checkinFile string) {
	year, mon, day := checkTime.Date()
	bdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	checkRepoDir := fmt.Sprintf("%s/checkin_history", bdir)
	checkinFilePath = fmt.Sprintf("%s/%s/%04d/%02d/%02d", checkRepoDir, user, year, mon, day)
	checkinFile = fmt.Sprintf("%s/checkin", checkinFilePath)
	return
}
