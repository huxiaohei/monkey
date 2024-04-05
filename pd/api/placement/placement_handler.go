package placement

import (
	"encoding/json"
	"fmt"
	"io"
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

func (ph *PlacementHandler) GenNewActor(actorType string, id uint64, ttl int64) (*placement.PlacementActorPosition, error) {
	hosts, err := ph.serverStorgate.GetAllServerInfo()
	if err != nil {
		mlog.Error("GenNewActor GetAllServerInfo failed ", err)
		return nil, err
	}
	var bestHost *placement.PlacementHostInfo
	for _, host := range hosts {
		if _, ok := host.Services[actorType]; !ok {
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
		mlog.Errorf("GenNewActor no host found %s_%d", actorType, id)
		return nil, fmt.Errorf("no host found %s_%d", actorType, id)
	}
	token, err := utils.GenerateToken(32)
	if err != nil {
		mlog.Error("GenNewActor GenerateToken failed ", err)
		return nil, err
	}
	pos := &placement.PlacementActorPosition{
		ActorType:  actorType,
		Id:         id,
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
	pos, ok, err := ph.actorStorage.GetActorInfo(req.ActorType, req.Id)
	if err != nil {
		mlog.Error("FindPosition GetActorInfo failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ok && (!ph.serverStorgate.IsServerValid(pos.ServerId) || pos.DeadTime < utils.GetNowSec()) {
		ok = false
		ph.actorStorage.DeleteActor(req.ActorType, req.Id)
	}
	if !ok {
		pos, err = ph.GenNewActor(req.ActorType, req.Id, req.TTL)
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
	pos, ok, err := ph.actorStorage.GetActorInfo(req.ActorType, req.Id)
	if err != nil {
		mlog.Error("KeepAlive GetActorInfo failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		mlog.Errorf("KeepAlive Actor not found %s_%d", req.ActorType, req.Id)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Actor not found %s_%d", req.ActorType, req.Id)})
		return
	}
	if pos.Token != req.Token {
		mlog.Errorf("KeepAlive Token not match %s_%d %s_%s", req.ActorType, req.Id, req, pos)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Token not match %s_%d", req.ActorType, req.Id)})
		return
	}
	if pos.DeadTime < utils.GetNowSec() {
		mlog.Error("KeepAlive Actor expired ", req, pos)
		ph.actorStorage.DeleteActor(req.ActorType, req.Id)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Actor expired"})
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
