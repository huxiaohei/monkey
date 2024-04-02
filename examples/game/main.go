package main

import (
	"monkey/logger"
	"monkey/placement"
	"monkey/rpc"
	"monkey/utils"
)

var (
	mlog, _ = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

func main() {

	pd := placement.NewPDPlacement("0.0.0.0:8000")
	serverId, err := pd.GenerateServerId()
	if err != nil {
		mlog.Error("generate server id error: ", err)
		return
	}
	leaseId := pd.RegisterServer(&placement.PlacementActorHostInfo{
		ServerId:  serverId,
		LeaseId:   0,
		Load:      0,
		StartTime: utils.GetNowMs(),
		TTL:       15,
		DeadTime:  0,
		Address:   "0.0.0.0:8000",
		Services:  map[string]string{"IPlayer": "0.0.0.0:8003"},
		Desc:      "IPlayer Server",
		Labels:    map[string]string{"IPlayer": "0.0.0.0:8003"},
	})
	if leaseId == 0 {
		mlog.Error("register server error")
		return
	}
	mlog.Infof("register server success, serverId: %d leaseId: %d", serverId, leaseId)
	pd.StartPulling()

	rpc.GetRPCClientManager().SetPlacement(pd)
	rpc.StartPlayerRPCServer("0.0.0.0:8003")

	for {
		utils.SleepSec(10)
	}
}
