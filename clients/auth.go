package clients

import (
	"context"

	authv1 "github.com/polar-bear-cu/sgt-proto/gen/go/auth/v1"
	"google.golang.org/grpc"
)

type AuthClient struct {
	rpc authv1.AuthServiceClient
}

func NewAuthClient(conn *grpc.ClientConn) *AuthClient {
	return &AuthClient{rpc: authv1.NewAuthServiceClient(conn)}
}

func (c *AuthClient) DeleteExpiredRefreshTokens(ctx context.Context) (int64, error) {
	resp, err := c.rpc.DeleteExpiredRefreshTokens(ctx, &authv1.DeleteExpiredRefreshTokensRequest{})
	if err != nil {
		return 0, err
	}
	return resp.GetDeletedCount(), nil
}
