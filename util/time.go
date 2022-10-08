package util

import (
	"time"
)

func ToLongWeekDay(weekday time.Weekday) string {
	tm := time.Date(2001, 1, int(weekday), 0, 0, 0, 0, time.UTC)
	return tm.Format("Monday")
}

func ToShortWeekDay(weekday time.Weekday) string {
	tm := time.Date(2001, 1, int(weekday), 0, 0, 0, 0, time.UTC)
	return tm.Format("Mon")
}
