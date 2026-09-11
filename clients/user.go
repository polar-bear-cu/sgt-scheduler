package clients

import (
	"context"

	userv1 "github.com/polar-bear-cu/sgt-proto/gen/go/user/v1"
	"google.golang.org/grpc"
)

type UserClient struct {
	rpc userv1.UserServiceClient
}

func NewUserClient(conn *grpc.ClientConn) *UserClient {
	return &UserClient{rpc: userv1.NewUserServiceClient(conn)}
}

func (c *UserClient) GetEmail(ctx context.Context, userID string) (string, error) {
	resp, err := c.rpc.GetUser(ctx, &userv1.GetUserRequest{Id: userID})
	if err != nil {
		return "", err
	}
	return resp.GetUser().GetEmail(), nil
}
