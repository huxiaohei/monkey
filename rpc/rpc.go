package rpc

import (
	"fmt"
	"monkey/logger"
	"monkey/placement"
	"monkey/rpc/pb"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	instance *RPCClientManager
	once     sync.Once
	mlog, _  = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type RPCClientManager struct {
	pd             placement.Placement
	gatewayClients map[uint64]*GatewayRPCClient
	playerClients  map[string]*PlayerRPCClient
}

func GetRPCClientManager() *RPCClientManager {
	once.Do(func() {
		instance = &RPCClientManager{
			gatewayClients: make(map[uint64]*GatewayRPCClient),
			playerClients:  make(map[string]*PlayerRPCClient),
		}
	})
	return instance
}

func (m *RPCClientManager) SetPlacement(pd placement.Placement) {
	m.pd = pd
}

func (m *RPCClientManager) GetGatewayClient(serverId uint64) (*GatewayRPCClient, error) {
	if client, ok := m.gatewayClients[serverId]; ok {
		return client, nil
	}

	hostInfo := m.pd.GetServerInfo(serverId)
	if hostInfo == nil {
		mlog.Errorf("get server info error: %d", serverId)
		return nil, fmt.Errorf("get server info error: %d", serverId)
	}

	address, ok := hostInfo.Services["IGateway"]
	if !ok {
		mlog.Errorf("get gateway rpc address error: %d", serverId)
		return nil, fmt.Errorf("get gateway rpc address error: %d", serverId)
	}

	mlog.Debug("dial to gateway ", address)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		mlog.Errorf("dial to gateway %s error: %v", address, err)
		return nil, err
	}

	client := pb.NewGatewayClient(conn)
	m.gatewayClients[serverId] = &GatewayRPCClient{
		conn:   conn,
		client: client,
	}
	return m.gatewayClients[serverId], nil
}

func (m *RPCClientManager) GetPlayerClient(id uint64) (*PlayerRPCClient, error) {
	actorInfo := m.pd.FindActorPositon(&placement.PlacementFindActorPositionArgs{
		ActorType: "IPlayer",
		Id:        id,
		TTL:       1800,
	})
	if actorInfo == nil {
		mlog.Errorf("find actor position error: player_%v", id)
		return nil, fmt.Errorf("find actor position error: player_%v", id)
	}
	if client, ok := m.playerClients[fmt.Sprintf("%d_%s", actorInfo.ServerId, actorInfo.ActorType)]; ok {
		return client, nil
	}

	hostInfo := m.pd.GetServerInfo(actorInfo.ServerId)
	if hostInfo == nil {
		mlog.Errorf("get server info error: %d", actorInfo.ServerId)
		return nil, fmt.Errorf("get server info error: %d", actorInfo.ServerId)
	}

	address, ok := hostInfo.Services["IPlayer"]
	if !ok {
		mlog.Errorf("get player rpc address error: %d", actorInfo.ServerId)
		return nil, fmt.Errorf("get player rpc address error: %d", actorInfo.ServerId)
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		mlog.Errorf("dial to player %s error: %v", address, err)
		return nil, err
	}

	key := fmt.Sprintf("%d_%s", actorInfo.ServerId, actorInfo.ActorType)
	client := pb.NewPlayerClient(conn)
	m.playerClients[key] = &PlayerRPCClient{
		conn:   conn,
		client: client,
	}
	return m.playerClients[key], nil
}
