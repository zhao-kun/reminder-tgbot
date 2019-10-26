package util

import "time"

var (
	chinaTime *time.Location
)

func init() {
	var err error
	chinaTime, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
}

// GetChinaTimeFromUnix return a `Asia/Shanghai` local time
func GetChinaTimeFromUnix(t int64) time.Time {
	checkTime := time.Unix(t, 0)
	return checkTime.In(chinaTime)
}

// GetChinaTimeNow return a `Asia/Shanghai` local time of current timestamp
func GetChinaTimeNow() time.Time {
	t := time.Now()
	return t.In(chinaTime)
}
