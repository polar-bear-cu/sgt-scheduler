package clients

import (
	"context"

	subscriptionv1 "github.com/polar-bear-cu/sgt-proto/gen/go/subscription/v1"
	"google.golang.org/grpc"

	"github.com/polar-bear-cu/sgt-scheduler/usecases"
)

type SubscriptionClient struct {
	rpc subscriptionv1.SubscriptionServiceClient
}

func NewSubscriptionClient(conn *grpc.ClientConn) *SubscriptionClient {
	return &SubscriptionClient{rpc: subscriptionv1.NewSubscriptionServiceClient(conn)}
}

func (c *SubscriptionClient) ListDueReminders(ctx context.Context, date string) ([]usecases.DueReminder, error) {
	resp, err := c.rpc.ListDueReminders(ctx, &subscriptionv1.ListDueRemindersRequest{Date: date})
	if err != nil {
		return nil, err
	}

	out := make([]usecases.DueReminder, 0, len(resp.GetReminders()))
	for _, r := range resp.GetReminders() {
		sub := r.GetSubscription()
		d := usecases.DueReminder{
			SubscriptionID:  sub.GetId(),
			UserID:          sub.GetUserId(),
			Name:            sub.GetName(),
			Cost:            sub.GetCost(),
			NextBillingDate: sub.GetBillingDate().AsTime(),
			BillingDue:      r.GetBillingDue(),
			TrialEndDue:     r.GetTrialEndDue(),
		}
		if sub.GetFtEndDate() != nil {
			ft := sub.GetFtEndDate().AsTime()
			d.FtEndDate = &ft
		}
		out = append(out, d)
	}
	return out, nil
}

func (c *SubscriptionClient) AdvanceBillingDates(ctx context.Context, date string) (advanced, converted int64, err error) {
	resp, err := c.rpc.AdvanceBillingDates(ctx, &subscriptionv1.AdvanceBillingDatesRequest{Date: date})
	if err != nil {
		return 0, 0, err
	}
	return resp.GetAdvancedCount(), resp.GetTrialsConvertedCount(), nil
}
