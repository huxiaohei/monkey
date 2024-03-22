package placement

import (
	"bytes"
	"encoding/json"
	"io"
	"monkey/actor"
	"monkey/logger"
	"monkey/utils"
	"net/http"
	"strings"
)

var (
	_          Placement = &PDPlacement{}
	pdloger, _           = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type PDPlacement struct {
	pdServerAddress string
	httpClient      *http.Client
	positionLRU     *utils.LRU[actor.ActorId, PlacementActorPosition]
	addServer       *utils.LRU[uint64, PlacementActorHostInfo]
	offlineServer   *utils.LRU[uint64, PlacementActorHostInfo]
	host            map[uint64]PlacementActorHostInfo
	curServerInfo   *PlacementActorHostInfo
	onAddServer     func(PlacementActorHostInfo)
	onRemoveServer  func(PlacementActorHostInfo)
	onServerOffline func(PlacementActorHostInfo)
	onFatalError    func(error)
	startPulling    bool
}

func NewPDPlacement(pdServerAddress string) *PDPlacement {
	pdServerAddress = strings.TrimSuffix(pdServerAddress, "/")
	if !strings.HasPrefix(pdServerAddress, "http") {
		pdServerAddress = "http://" + pdServerAddress
	}
	return &PDPlacement{
		pdServerAddress: pdServerAddress,
		httpClient:      &http.Client{},
		positionLRU:     utils.NewLRU[actor.ActorId, PlacementActorPosition](20000),
		addServer:       utils.NewLRU[uint64, PlacementActorHostInfo](1024),
		offlineServer:   utils.NewLRU[uint64, PlacementActorHostInfo](1024),
		host:            make(map[uint64]PlacementActorHostInfo),
		curServerInfo:   &PlacementActorHostInfo{},
		startPulling:    false,
	}
}

func (pdp *PDPlacement) get(path string) (int, []byte) {
	url := pdp.pdServerAddress + path
	resp, err := pdp.httpClient.Get(url)
	if err != nil {
		pdloger.Errorf("get %s failed, %v", url, err)
		return http.StatusInternalServerError, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		pdloger.Errorf("read response body failed, %v", err)
		return http.StatusInternalServerError, nil
	}
	return resp.StatusCode, body
}

func (pdp *PDPlacement) post(path string, body []byte) (int, []byte) {
	url := pdp.pdServerAddress + path
	resp, err := pdp.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		pdloger.Errorf("post %s failed, %v", url, err)
		return http.StatusInternalServerError, nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		pdloger.Errorf("read response body failed, %v", err)
		return http.StatusInternalServerError, nil
	}
	return resp.StatusCode, respBody
}

func (pdp *PDPlacement) IsServerValid(serverId uint64) bool {
	if info, ok := pdp.host[serverId]; ok {
		if info.DeadTime > utils.GetNowMs() {
			return true
		}
	}
	return false
}

func (pdp *PDPlacement) GenerateServerId() (uint64, error) {
	status, body := pdp.post("/pd/api/v1/id/newServerId", nil)
	if status != http.StatusOK {
		return 0, utils.NewError("generate server id failed, status: %d, body: %s", status, body)
	}
	var resp SequenceResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return 0, utils.NewError("unmarshal response body failed, err: %v body: %s", err, body)
	}
	return resp.Id, nil
}

func (pdp *PDPlacement) GenerateNewSequence(sequenceType string, step int) uint64 {
	status, body := pdp.post("/pd/api/v1/id/newSequence", []byte(`{"sequenceType":"`+sequenceType+`","step":`+string(rune(step))+`}`))
	if status != http.StatusOK {
		pdloger.Errorf("generate new sequence failed, status: %d, body: %s", status, body)
		return 0
	}
	var resp SequenceResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		pdloger.Errorf("unmarshal response body failed, err: %v body: %s", err, body)
		return 0
	}
	return resp.Id
}

func (pdp *PDPlacement) GenerateNewToken() (*GenerateNewTokenResponse, error) {
	status, body := pdp.post("/pd/api/v1/placement/newToken", nil)
	if status != http.StatusOK {
		pdloger.Errorf("generate new token failed, status: %d, body: %s", status, body)
		return nil, utils.NewError("generate new token failed, status: %d, body: %s", status, body)
	}
	var resp GenerateNewTokenResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		pdloger.Errorf("unmarshal response body failed, err: %v body: %s", err, body)
		return nil, utils.NewError("unmarshal response body failed, err: %v body: %s", err, body)
	}
	return &resp, nil
}

func (pdp *PDPlacement) RegisterServer(info *PlacementActorHostInfo) uint64 {
	if info.TTL <= 0 {
		info.TTL = 15
	}
	data, err := json.Marshal(info)
	if err != nil {
		pdloger.Errorf("marshal register server info failed, err: %v", err)
		return 0
	}
	status, body := pdp.post("/pd/api/v1/membership/registerServer", data)
	if status != http.StatusOK {
		pdloger.Errorf("register server failed, status: %d, body: %s", status, body)
		return 0
	}
	var resp RegisterServerResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		pdloger.Errorf("unmarshal response body failed, err: %v body: %s", err, body)
		return 0
	}
	if resp.LeaseId != 0 {
		pdp.curServerInfo.ServerId = info.ServerId
		pdp.curServerInfo.LeaseId = resp.LeaseId
		pdp.curServerInfo.Load = info.Load
		pdp.curServerInfo.StartTime = info.StartTime
		pdp.curServerInfo.TTL = info.TTL
		pdp.curServerInfo.DeadTime = info.DeadTime
		pdp.curServerInfo.Address = info.Address
		pdp.curServerInfo.Services = info.Services
		pdp.curServerInfo.Desc = info.Desc
		pdp.curServerInfo.Labels = info.Labels
	}
	pdloger.Infof("register server success, serverId: %d leaseId: %d", info.ServerId, resp.LeaseId)
	return resp.LeaseId
}

func (pdp *PDPlacement) KeepAliveServer(serverId uint64, leaseId uint64, load uint64) *PlacementKeepAliveResponse {
	args := &ServerKeepAliveArgs{
		ServerId: serverId,
		LeaseId:  leaseId,
		Load:     load,
	}
	data, err := json.Marshal(args)
	if err != nil {
		pdloger.Errorf("marshal keep alive server args failed, err: %v", err)
		return nil
	}
	status, body := pdp.post("/pd/api/v1/membership/keepAliveServer", data)
	if status != http.StatusOK {
		pdloger.Errorf("keep alive server failed, status: %d, body: %s", status, body)
		return nil
	}
	var resp PlacementKeepAliveResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		pdloger.Errorf("unmarshal response body failed, err: %v body: %s", err, body)
		return nil
	}
	return &resp
}

func (pdp *PDPlacement) FindActorPositionInCache(request *PlacementFindActorPositionArgs) *PlacementActorPosition {
	resp, ok := pdp.positionLRU.Get(request.ActorId)
	if ok {
		return resp
	}
	return nil
}

func (pdp *PDPlacement) FindActorPositon(request *PlacementFindActorPositionArgs) *PlacementActorPosition {
	pos, ok := pdp.positionLRU.Get(request.ActorId)
	if ok && pdp.IsServerValid(pos.ServerId) {
		return pos
	}
	data, err := json.Marshal(request)
	if err != nil {
		pdloger.Errorf("marshal find actor position request failed, err: %v", err)
		return nil
	}
	status, body := pdp.post("/pd/api/v1/placement/findPosition", data)
	if status != http.StatusOK {
		pdloger.Errorf("find actor position failed, status: %d, body: %s", status, body)
		return nil
	}
	var resp PlacementActorPosition
	err = json.Unmarshal(body, &resp)
	if err != nil {
		pdloger.Errorf("unmarshal response body failed, err: %v body: %s", err, body)
		return nil
	}
	_, ok = pdp.offlineServer.Get(resp.ServerId)
	if !ok {
		pdp.positionLRU.Put(request.ActorId, resp)
	}
	return &resp
}

func (pdp *PDPlacement) ClearActorPositionCache(request *PlacementFindActorPositionArgs) {
	pdp.positionLRU.Remove(request.ActorId)
	pdloger.Infof("clear actor position cache, actorId: %v", request.ActorId)
}

func (pdp *PDPlacement) GetVersionInfo() *PlacementVersionInfo {
	status, body := pdp.get("/pd/api/v1/version")
	if status != http.StatusOK {
		pdloger.Errorf("get version info failed, status: %d, body: %s", status, body)
		return nil
	}
	var resp PlacementVersionInfo
	err := json.Unmarshal(body, &resp)
	if err != nil {
		pdloger.Errorf("unmarshal response body failed, err: %v body: %s", err, body)
		return nil
	}
	return &resp
}

func (pdp *PDPlacement) GetCurServerId() uint64 {
	return pdp.curServerInfo.ServerId
}

func (pdp *PDPlacement) RegisterServerChangedEvent(onAddServer func(PlacementActorHostInfo), onRemoveServer func(PlacementActorHostInfo), onServerOffline func(PlacementActorHostInfo), onFatalError func(error)) {
	pdp.onAddServer = onAddServer
	pdp.onRemoveServer = onRemoveServer
	pdp.onServerOffline = onServerOffline
	pdp.onFatalError = onFatalError
}

func (pdp *PDPlacement) SetServerLoad(load uint64) {
	pdp.curServerInfo.Load = load
}

func (pdp *PDPlacement) processAddServerEvent(newServers map[uint64]PlacementActorHostInfo, events PlacementEvents) {
	if len(events.Add) == 0 {
		return
	}
	for _, serverId := range events.Add {
		server, ok := newServers[serverId]
		if !ok || pdp.addServer.Contain(serverId) {
			continue
		}
		pdp.addServer.Put(server.ServerId, server)
		if pdp.onAddServer != nil {
			pdp.onAddServer(server)
		}
	}
}

func (pdp *PDPlacement) processRemoveServerEvent(events PlacementEvents) {
	if len(events.Remove) == 0 {
		return
	}
	if pdp.onRemoveServer == nil {
		return
	}
	for _, serverId := range events.Remove {
		if server, ok := pdp.host[serverId]; ok {
			pdp.onRemoveServer(server)
		}
	}
}

func (pdp *PDPlacement) processServerOfflineEvent(newServers map[uint64]PlacementActorHostInfo) {
	for serverId, server := range newServers {
		if server.DeadTime <= utils.GetNowMs() && !pdp.offlineServer.Contain(serverId) {
			pdp.offlineServer.Put(serverId, server)
			if pdp.onServerOffline != nil {
				pdp.onServerOffline(server)
			}
		}
	}
}

func (pdp *PDPlacement) processDiffTwoServerList(newServers map[uint64]PlacementActorHostInfo) {
	for serverId, server := range newServers {
		if _, ok := pdp.host[serverId]; ok {
			continue
		}
		if pdp.addServer.Contain(serverId) {
			continue
		}
		pdp.addServer.Put(serverId, server)
		if pdp.onAddServer != nil {
			pdp.onAddServer(server)
		}
	}
	pdp.host = newServers
}

func (pdp *PDPlacement) pullOnce() (ok bool) {
	defer func() {
		if err := recover(); err != nil {
			pdloger.Errorf("pullOnce failed, err: %v", err)
			ok = false
		}
	}()

	resp := pdp.KeepAliveServer(pdp.curServerInfo.ServerId, pdp.curServerInfo.LeaseId, pdp.curServerInfo.Load)
	if resp == nil {
		pdloger.Errorf("pullOnce failed resp is nil")
		return false
	}
	pdloger.Infof("pullOnce Host:%d Event:%d", len(resp.Events), len(resp.Hosts))
	for _, event := range resp.Events {
		pdp.processAddServerEvent(resp.Hosts, event)
		pdp.processRemoveServerEvent(event)
	}
	pdp.processServerOfflineEvent(resp.Hosts)
	pdp.processDiffTwoServerList(resp.Hosts)
	return true
}

func (pdp *PDPlacement) pullServer() {
	timerInterval := pdp.curServerInfo.TTL / 3
	timerCount, failedCount := 0, 0
	for pdp.startPulling {
		ok := pdp.pullOnce()
		if ok {
			timerCount++
		} else {
			failedCount++
		}
		utils.SleepSec(timerInterval)
		if failedCount > 3 {
			pdloger.Errorf("pullServer failed, failedCount: %d", failedCount)
			if pdp.onFatalError != nil {
				pdp.onFatalError(utils.NewError("pullServer failed, failedCount: %d", failedCount))
			}
			break
		}
		failedCount = 0
	}
}

func (pdp *PDPlacement) StartPulling() {
	if pdp.startPulling {
		return
	}
	pdp.startPulling = true
	go pdp.pullServer()
}

func (pdp *PDPlacement) StopPulling() {
	pdp.startPulling = false
}
