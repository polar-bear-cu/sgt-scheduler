package usecases

import (
	"context"
	"errors"
	"fmt"
)

type Subscription struct {
	UserID string
	Name   string
}

type SubscriptionSource interface {
	ListUpcomingForBilling(ctx context.Context, withinHours int32) ([]Subscription, error)
}

type UserDirectory interface {
	GetEmail(ctx context.Context, userID string) (string, error)
}

type Publisher interface {
	Publish(ctx context.Context, to, subject, body string) error
}

type BillingReminderUsecase struct {
	subs  SubscriptionSource
	users UserDirectory
	pub   Publisher
}

func NewBillingReminder(subs SubscriptionSource, users UserDirectory, pub Publisher) *BillingReminderUsecase {
	return &BillingReminderUsecase{subs: subs, users: users, pub: pub}
}

// Run finds subscriptions billing within withinHours and publishes a reminder
// email for each. It keeps going on a per-subscription failure and returns
// the count actually sent plus every error joined together.
func (u *BillingReminderUsecase) Run(ctx context.Context, withinHours int32) (int, error) {
	subs, err := u.subs.ListUpcomingForBilling(ctx, withinHours)
	if err != nil {
		return 0, err
	}

	var errs error
	sent := 0
	for _, sub := range subs {
		email, err := u.users.GetEmail(ctx, sub.UserID)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		subject := fmt.Sprintf("Upcoming charge: %s", sub.Name)
		body := fmt.Sprintf("Your subscription %q is due for billing soon.", sub.Name)
		if err := u.pub.Publish(ctx, email, subject, body); err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		sent++
	}
	return sent, errs
}
