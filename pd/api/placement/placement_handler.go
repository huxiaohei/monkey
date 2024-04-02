package placement

import (
	"encoding/json"
	"fmt"
	"io"
	"monkey/actor"
	"monkey/logger"
	"monkey/pd/storage"
	"monkey/placement"
	"monkey/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	mlog, _ = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type PlacementHandler struct {
	actorStorage   storage.ActorStorage
	serverStorgate storage.ServerStorage
}

func NewPlacementHandler(serverStorgate storage.ServerStorage, actorStorage storage.ActorStorage) *PlacementHandler {
	return &PlacementHandler{
		serverStorgate: serverStorgate,
		actorStorage:   actorStorage,
	}
}

func (ph *PlacementHandler) GenNewActor(actorId *actor.ActorId, ttl int64) (*placement.PlacementActorPosition, error) {
	hosts, err := ph.serverStorgate.GetAllServerInfo()
	if err != nil {
		mlog.Error("GenNewActor GetAllServerInfo failed ", err)
		return nil, err
	}
	var bestHost *placement.PlacementActorHostInfo
	for _, host := range hosts {
		if _, ok := host.Services[actorId.ActorType]; !ok {
			continue
		}
		if bestHost == nil {
			bestHost = host
			continue
		}
		if bestHost.Load > host.Load {
			bestHost = host
		}
	}
	if bestHost == nil {
		mlog.Error("GenNewActor no host found ", actorId)
		return nil, fmt.Errorf("no host found %v", actorId)
	}
	token, err := utils.GenerateToken(32)
	if err != nil {
		mlog.Error("GenNewActor GenerateToken failed ", err)
		return nil, err
	}
	pos := &placement.PlacementActorPosition{
		ActorId:    *actorId,
		TTL:        ttl,
		CreateTime: utils.GetNowSec(),
		DeadTime:   utils.GetNowSec() + ttl,
		ServerId:   bestHost.ServerId,
		Token:      token,
	}
	err = ph.actorStorage.PutActorInfo(pos)
	if err != nil {
		mlog.Error("GenNewActor PutActorInfo failed ", err)
		return nil, err
	}
	return pos, nil
}

func (ph *PlacementHandler) FindPosition(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		mlog.Error("FindPosition Error reading request body ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}

	var req placement.PlacementFindActorPositionArgs
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		mlog.Error("FindPosition json.Unmarshal failed ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TTL == 0 {
		req.TTL = 1800
	}
	pos, ok, err := ph.actorStorage.GetActorInfo(&req.ActorId)
	if err != nil {
		mlog.Error("FindPosition GetActorInfo failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ok && !ph.serverStorgate.IsServerValid(pos.ServerId) {
		ok = false
		ph.actorStorage.DeleteActor(&req.ActorId)
	}
	if !ok {
		pos, err = ph.GenNewActor(&req.ActorId, req.TTL)
		if err != nil {
			mlog.Error("FindPosition GenNewActor failed ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		pos.DeadTime = utils.GetNowSec() + pos.TTL
		err = ph.actorStorage.PutActorInfo(pos)
		if err != nil {
			mlog.Error("KeepAlive PutActorInfo failed ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	mlog.Debugf("FindPosition %s", pos)
	c.JSON(200, pos)
}

func (ph *PlacementHandler) KeepAlive(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		mlog.Error("KeepAlive Error reading request body ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}

	var req placement.ActorKeepAliveArgs
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		mlog.Error("KeepAlive json.Unmarshal failed ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pos, ok, err := ph.actorStorage.GetActorInfo(&req.ActorId)
	if err != nil {
		mlog.Error("KeepAlive GetActorInfo failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		mlog.Error("KeepAlive Actor not found ", req.ActorId)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Actor not found"})
		return
	}
	if pos.Token != req.Token {
		mlog.Error("KeepAlive Token not match ", req, pos)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token not match"})
		return
	}
	pos.DeadTime = utils.GetNowSec() + pos.TTL
	err = ph.actorStorage.PutActorInfo(pos)
	if err != nil {
		mlog.Error("KeepAlive PutActorInfo failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, pos)
}
