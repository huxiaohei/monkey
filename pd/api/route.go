package api

import (
	"monkey/pd/api/id"
	"monkey/pd/api/membership"
	"monkey/pd/api/placement"
	"monkey/pd/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func makeRouteHandler(handler func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		handler(c)
	}
}

func RegisterRouteV1(engine *gin.Engine, sequenceStorage storage.SequenceStorage, serverStorgate storage.ServerStorage, actorStorage storage.ActorStorage, version string) {

	engine.GET("/pd/api/v1/ping", Ping)

	idV1 := engine.Group("/pd/api/v1/id")
	{
		idHandler := id.NewIdHandler(sequenceStorage)

		idV1.POST("newServerId", makeRouteHandler(idHandler.NewServerId))
		idV1.POST("newSequence", makeRouteHandler(idHandler.NewSequence))
	}

	membershipV1 := engine.Group("/pd/api/v1/membership")
	{
		membershipHandler := membership.NewMembershipHandler(sequenceStorage, serverStorgate, version)

		membershipV1.POST("registerServer", makeRouteHandler(membershipHandler.RegisterServer))
		membershipV1.POST("keepAliveServer", makeRouteHandler(membershipHandler.KeepAliveServer))
		membershipV1.GET("allServers", makeRouteHandler(membershipHandler.AllServers))
		membershipV1.GET("version", makeRouteHandler(membershipHandler.Version))
	}

	placementV1 := engine.Group("/pd/api/v1/placement")
	{
		placementHandler := placement.NewPlacementHandler(serverStorgate, actorStorage)

		placementV1.POST("findPosition", makeRouteHandler(placementHandler.FindPosition))
		placementV1.POST("actorKeepAlive", makeRouteHandler(placementHandler.KeepAlive))
	}

}
