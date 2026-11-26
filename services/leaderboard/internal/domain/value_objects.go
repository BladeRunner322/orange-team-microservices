package domain

import (
	"errors"
	"time"
)

// PeriodKind задаёт вид календарного периода.
type PeriodKind string

const (
	PeriodDaily   PeriodKind = "daily"
	PeriodWeekly  PeriodKind = "weekly"
	PeriodMonthly PeriodKind = "monthly"
)

// Period — интервал в UTC с включённым началом и исключённым концом.
type Period struct {
	Start time.Time
	End   time.Time
}

// Bounds определяет границы периода, в который попадает now.
func (kind PeriodKind) Bounds(now time.Time) (Period, error) {
	// TODO: Вычислить календарные границы периода в UTC и обработать неизвестный вид периода.
	return Period{}, errors.ErrUnsupported
}
