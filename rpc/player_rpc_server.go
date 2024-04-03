package rpc

import (
	"context"
	"monkey/rpc/pb"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PlayerRPCServer struct {
	pb.UnimplementedPlayerServer
}

func StartPlayerRPCServer(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		mlog.Errorf("start player rpc server error: %v", err)
		return err
	}

	s := grpc.NewServer()
	pb.RegisterPlayerServer(s, &PlayerRPCServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			mlog.Errorf("start player rpc server error: %v", err)
		}
	}()
	return nil
}

func (p *PlayerRPCServer) Bind(ctx context.Context, msg *pb.BindMsg) (*emptypb.Empty, error) {
	mlog.Infof("bind %d %d %d %d %s %s", msg.ServerId, msg.UserId, msg.SessionId, msg.MsgSeq, msg.AccountToken, msg.TtlToken)
	return nil, nil
}

func (p *PlayerRPCServer) ReceiveMessage(ctx context.Context, msg *pb.CommonMsg) (*emptypb.Empty, error) {
	mlog.Infof("receive message %d_%d %v", msg.ServerId, msg.SessionId, msg.Payload)
	client, err := GetRPCClientManager().GetGatewayClient(msg.ServerId)
	if err != nil {
		return nil, err
	}
	client.SendMessage(&pb.ClientMsg{
		SessionId: msg.SessionId,
		Payload:   msg.Payload,
	})
	return nil, nil
}
