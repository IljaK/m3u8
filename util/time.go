package util

import (
	"context"
	"time"
)

type DurationCounter struct {
	startTS time.Time
	timeout time.Duration
}

func (t *DurationCounter) Start(timeout time.Duration) {
	t.startTS = time.Now()
	t.timeout = timeout
}
func (t *DurationCounter) RemainAt(ts time.Time) time.Duration {
	return t.timeout - t.ElapsedAt(ts)
}
func (t *DurationCounter) ElapsedAt(ts time.Time) time.Duration {
	if t.startTS.IsZero() {
		return 0
	}
	if ts.After(t.startTS.Add(t.timeout)) {
		return t.timeout
	}
	return ts.Sub(t.startTS)
}
func (t *DurationCounter) IsCompletedAt(ts time.Time) bool {
	return t.ElapsedAt(ts) == t.timeout
}
func (t *DurationCounter) Remain() time.Duration {
	return t.RemainAt(time.Now())
}
func (t *DurationCounter) RemainSeconds() uint32 {
	return uint32(t.Remain() / time.Second)
}
func (t *DurationCounter) Elapsed() time.Duration {
	return t.ElapsedAt(time.Now())
}
func (t *DurationCounter) IsCompleted() bool {
	return t.IsCompletedAt(time.Now())
}
func (t *DurationCounter) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), t.Remain())
}
func (t *DurationCounter) RemainWithBuffer(buffer time.Duration) time.Duration {
	remain := t.RemainAt(time.Now())
	if remain > buffer {
		return remain - buffer
	}
	return remain
}

func StartCounter(timeout time.Duration) *DurationCounter {
	c := DurationCounter{}
	c.Start(timeout)
	return &c
}

func ToLongWeekDay(weekday time.Weekday) string {
	tm := time.Date(2001, 1, int(weekday), 0, 0, 0, 0, time.UTC)
	return tm.Format("Monday")
}

func ToShortWeekDay(weekday time.Weekday) string {
	tm := time.Date(2001, 1, int(weekday), 0, 0, 0, 0, time.UTC)
	return tm.Format("Mon")
}
