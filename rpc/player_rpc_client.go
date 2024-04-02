package rpc

import (
	"context"
	"monkey/rpc/pb"
	"time"

	"google.golang.org/grpc"
)

type PlayerRPCClient struct {
	conn   *grpc.ClientConn
	client pb.PlayerClient
}

func (p *PlayerRPCClient) Bind(msg *pb.BindMsg) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	p.client.Bind(ctx, msg)
}

func (p *PlayerRPCClient) ReceiveMessage(msg *pb.CommonMsg) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	p.client.ReceiveMessage(ctx, msg)
}
