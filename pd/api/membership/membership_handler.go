package membership

import (
	"encoding/json"
	"io"
	"monkey/logger"
	"monkey/pd/storage"
	"monkey/placement"
	"monkey/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	mhloger, _ = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type MembershipHandler struct {
	sequenceStorage   storage.SequenceStorage
	serverStorgate    storage.ServerStorage
	version           string
	lastHeartBeatTime int64
}

func NewMembershipHandler(sequenceStorage storage.SequenceStorage, serverStorgate storage.ServerStorage, version string) *MembershipHandler {
	return &MembershipHandler{
		sequenceStorage:   sequenceStorage,
		serverStorgate:    serverStorgate,
		version:           version,
		lastHeartBeatTime: utils.GetNowSec(),
	}
}

func (mh *MembershipHandler) updateLastHeartBeatTime() {
	mh.lastHeartBeatTime = utils.GetNowSec()
}

func (mh *MembershipHandler) RegisterServer(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		mhloger.Error("RegisterServer ReadAll failed ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}

	var info placement.PlacementHostInfo
	err = json.Unmarshal(bodyBytes, &info)
	if err != nil {
		mhloger.Error("RegisterServer Unmarshal failed ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = mh.serverStorgate.RegisterServer(&info)
	if err != nil {
		mhloger.Error("RegisterServer RegisterServer failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := placement.RegisterServerResponse{
		LeaseId: info.LeaseId,
	}
	mh.updateLastHeartBeatTime()
	c.JSON(200, resp)
}

func (mh *MembershipHandler) KeepAliveServer(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		mhloger.Error("KeepAliveServer ReadAll failed ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}

	var args placement.ServerKeepAliveArgs
	err = json.Unmarshal(bodyBytes, &args)
	if err != nil {
		mhloger.Error("KeepAliveServer Unmarshal failed ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = mh.serverStorgate.KeepAliveServer(args.ServerId, args.LeaseId, args.Load)
	if err != nil {
		mhloger.Error("KeepAliveServer KeepAliveServer failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hosts, err := mh.serverStorgate.GetAllServerInfo()
	if err != nil {
		mhloger.Error("KeepAliveServer GetAllServerInfo failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := placement.ServerKeepAliveResponse{
		Hosts:  make(map[uint64]placement.PlacementHostInfo, 0),
		Events: make([]placement.PlacementEvents, 0),
	}
	for _, h := range hosts {
		resp.Hosts[h.ServerId] = *h
	}
	mh.updateLastHeartBeatTime()
	c.JSON(http.StatusOK, resp)
}

func (mh *MembershipHandler) AllServers(c *gin.Context) {
	hosts, err := mh.serverStorgate.GetAllServerInfo()
	if err != nil {
		mhloger.Error("AllServers GetAllServerInfo failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hosts": hosts})
}

func (mh *MembershipHandler) Version(c *gin.Context) {
	resp := placement.PlacementVersionInfo{
		Version:           mh.version,
		LastHeartBeatTime: mh.lastHeartBeatTime,
	}
	c.JSON(http.StatusOK, resp)
}
