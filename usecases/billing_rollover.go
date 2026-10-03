package usecases

import (
	"context"
	"time"
)

type BillingDateAdvancer interface {
	AdvanceBillingDates(ctx context.Context, date string) (advanced, converted int64, err error)
}

type RolloverResult struct {
	Date            string
	Advanced        int64
	TrialsConverted int64
}

type BillingRolloverUsecase struct {
	subs BillingDateAdvancer
	loc  *time.Location
}

func NewBillingRollover(subs BillingDateAdvancer, loc *time.Location) *BillingRolloverUsecase {
	return &BillingRolloverUsecase{subs: subs, loc: loc}
}

// Run must happen before the day's reminder check so reminders see the next cycle's dates.
func (u *BillingRolloverUsecase) Run(ctx context.Context, now time.Time) (RolloverResult, error) {
	res := RolloverResult{Date: now.In(u.loc).Format(time.DateOnly)}
	advanced, converted, err := u.subs.AdvanceBillingDates(ctx, res.Date)
	res.Advanced, res.TrialsConverted = advanced, converted
	return res, err
}
