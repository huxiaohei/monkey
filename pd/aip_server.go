package pd

import (
	"monkey/logger"
	"monkey/pd/api"
	"monkey/pd/storage"

	"github.com/gin-gonic/gin"
)

var (
	mlog, _ = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

func Start(sequenceStorage storage.SequenceStorage, serverStorgate storage.ServerStorage, actorStorage storage.ActorStorage, address string, version string) {
	engine := gin.Default()

	api.RegisterRouteV1(engine, sequenceStorage, serverStorgate, actorStorage, version)

	err := engine.Run(address)

	if err != nil {
		mlog.Error("pd start failed", err)
	}
}
