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

func (c *SubscriptionClient) ListUpcomingForBilling(ctx context.Context, withinHours int32) ([]usecases.Subscription, error) {
	resp, err := c.rpc.GetUpcomingForBilling(ctx, &subscriptionv1.GetUpcomingForBillingRequest{WithinHours: withinHours})
	if err != nil {
		return nil, err
	}

	out := make([]usecases.Subscription, 0, len(resp.GetSubscription()))
	for _, sub := range resp.GetSubscription() {
		out = append(out, usecases.Subscription{UserID: sub.GetUserId(), Name: sub.GetName()})
	}
	return out, nil
}
