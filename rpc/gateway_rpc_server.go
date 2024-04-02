package rpc

import (
	"context"
	"monkey/network"
	"monkey/rpc/pb"
	"monkey/utils"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GatewayRPCServer struct {
	pb.UnimplementedGatewayServer
	sessionManager network.SessionManager
}

func StartGatewayRPCServer(address string, sessionManager network.SessionManager) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		mlog.Errorf("GatewayRPCServer listen error: %v", err)
		return err
	}

	s := grpc.NewServer()
	pb.RegisterGatewayServer(s, &GatewayRPCServer{sessionManager: sessionManager})
	go func() {
		if err := s.Serve(lis); err != nil {
			mlog.Errorf("GatewayRPCServer server error: %v", err)
		}
	}()
	return nil
}

func (p *GatewayRPCServer) SendMessage(ctx context.Context, msg *pb.ClientMsg) (*emptypb.Empty, error) {
	if session, ok := p.sessionManager.GetSession(utils.SessionId(msg.SessionId)); ok {
		session.SendMessage(msg.Payload)
	}
	return nil, nil
}
