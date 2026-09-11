package usecases

import "context"

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
