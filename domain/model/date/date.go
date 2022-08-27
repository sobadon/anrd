package date

import (
	"time"

	"github.com/sobadon/anrd/internal/timeutil"
)

// 年月日
type Date time.Time

func New(year int, month time.Month, day int) Date {
	return Date(time.Date(year, month, day, 0, 0, 0, 0, timeutil.LocationJST()))
}

func NewFromToday(today time.Time) Date {
	return Date(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, timeutil.LocationJST()))
}
