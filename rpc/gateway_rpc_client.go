package rpc

import (
	"context"
	"monkey/rpc/pb"
	"time"

	"google.golang.org/grpc"
)

type GatewayRPCClient struct {
	conn   *grpc.ClientConn
	client pb.GatewayClient
}

func (g *GatewayRPCClient) SendMessage(msg *pb.ClientMsg) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	g.client.SendMessage(ctx, msg)
}
