// Пакет timetz — разбор строк дат/времени для HTTP и gRPC слоёв (без зависимости от transport).
package timetz

import (
	"fmt"
	"time"
)

// ParseFlexible парсит RFC3339/RFC3339Nano либо дату вида 2006-01-02 (UTC полночь).
func ParseFlexible(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}

// StartOfDayUTC — начало календарного дня в UTC для выборки «на день».
func StartOfDayUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// StartOfMonthUTC — первое число месяца (UTC 00:00) для выборки «на месяц».
func StartOfMonthUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC)
}
