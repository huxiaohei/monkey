package gateway

import (
	"monkey/conf"
	"monkey/gateway/codec"
	"monkey/gateway/handler"
	"monkey/logger"
	"monkey/network"
	"monkey/placement"
	"monkey/rpc"
	"monkey/utils"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	receiveTimeout int64                  = 30
	sessionManager network.SessionManager = network.NewSessionManager()
	mlog, _                               = logger.GetLoggerManager().GetLogger(logger.MainTag)
	pd             *placement.PDPlacement = nil
)

func accept(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		mlog.Error("upgrade:", err)
		return
	}
	s := network.NewWebSocketSession(c, &codec.GatewayCodec{}, &handler.GatewayMessageHandler{
		PD: pd,
	}, sessionManager)
	s.ReceiveMessage(receiveTimeout)
}

func registerServer(cfg *conf.GatewayConf) uint64 {
	serverId, err := pd.GenerateServerId()
	if err != nil {
		mlog.Error("generate server id error: ", err)
		return 0
	}
	leaseId := pd.RegisterServer(&placement.PlacementHostInfo{
		ServerId:  serverId,
		LeaseId:   0,
		Load:      cfg.Load,
		StartTime: utils.GetNowMs(),
		TTL:       cfg.ServerTTL,
		DeadTime:  0,
		Address:   cfg.ListenAddress,
		Services:  cfg.Services,
		Desc:      cfg.Desc,
		Labels:    cfg.Labels,
	})
	if leaseId == 0 {
		mlog.Error("register server error")
		return 0
	}
	mlog.Infof("register server success, serverId: %d leaseId: %d", serverId, leaseId)
	pd.StartPulling()
	return serverId
}

func Start(cfg *conf.GatewayConf) {
	mlog.Infof("Gateway starting %v", cfg)

	pd = placement.NewPDPlacement(cfg.PdAddress)
	rpc.GetRPCClientManager().SetPlacement(pd)
	serverId := registerServer(cfg)
	if serverId == 0 {
		mlog.Errorf("register server error pdAddress:%v", cfg.PdAddress)
		return
	}

	rpc.StartGatewayRPCServer(cfg.Services["IGatewayRPC"], sessionManager)

	receiveTimeout = cfg.ReceiveTimeout
	http.HandleFunc(cfg.ListenPath, accept)
	err := http.ListenAndServe(cfg.ListenAddress, nil)

	mlog.Info("Gateway end with error: ", err)
}
