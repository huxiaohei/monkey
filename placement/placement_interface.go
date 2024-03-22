package placement

import (
	"monkey/actor"
)

type SequenceResponse struct {
	Id uint64 `json:"id" description:"新的ID"`
}

type GenerateNewTokenResponse struct {
	Token       string `json:"token" description:"新的Token"`
	InvalidTime int64  `json:"invalidTime" description:"Token的失效时间"`
}

type RegisterServerResponse struct {
	LeaseId uint64 `json:"leaseId" description:"租约ID"`
}

// PD服务器上, Actor宿主服务器的信息
type PlacementActorHostInfo struct {
	ServerId  uint64            `json:"serverId" description:"服务器唯一ID"`
	LeaseId   uint64            `json:"leaseId" description:"租约ID"`
	Load      uint64            `json:"load" description:"负载"`
	StartTime int64             `json:"startTime" description:"启动时间"`
	TTL       int64             `json:"ttl" description:"租约时间"`
	DeadTime  int64             `json:"deadTime" description:"租约过期时间"`
	Address   string            `json:"address" description:"服务器的地址"`
	Services  map[string]string `json:"services" description:"服务器能提供的Actor对象类型"`
	Desc      string            `json:"desc" description:"服务器的描述"`
	Labels    map[string]string `json:"labels" description:"服务器的额外属性, 用来表示网关等信息"`
}

// PD上最近发生的事件
type PlacementEvents struct {
	Time   uint64   `json:"time" description:"事件发生的时间"`
	Add    []uint64 `json:"add" description:"添加的服务器ID"`
	Remove []uint64 `json:"remove" description:"删除的服务器ID"`
}

// 服务器续约请求
type ServerKeepAliveArgs struct {
	ServerId uint64 `json:"serverId" description:"服务器唯一ID"`
	LeaseId  uint64 `json:"leaseId" description:"租约ID"`
	Load     uint64 `json:"load" description:"负载"`
}

// 服务器续约返回
type PlacementKeepAliveResponse struct {
	Hosts  map[uint64]PlacementActorHostInfo `json:"hosts" description:"每次续约PD会将所有的服务器信息下发"`
	Events []PlacementEvents                 `json:"events" description:"服务器最近的事件(增减和删除)"`
}

// PD上对于Actor定位的请求
type PlacementFindActorPositionArgs struct {
	ActorId actor.ActorId `json:"actorId" description:"ActorID"`
	TTL     int64         `json:"ttl" description:"租约时间"`
}

// PD上对于Actor定位的返回
type PlacementActorPosition struct {
	ActorId    actor.ActorId `json:"actorId" description:"ActorID"`
	TTL        int64         `json:"ttl" description:"租约时间"`
	CreateTime int64         `json:"createTime" description:"创建时间"`
	DeadTime   int64         `json:"deadTime" description:"租约过期时间"`
	ServerId   uint64        `json:"serverId" description:"宿主的唯一ID"`
	Token      string        `json:"token" description:"写入Token"`
}

// Actor续约请求
type ActorKeepAliveArgs struct {
	ActorId actor.ActorId `json:"actorId" description:"ActorID"`
	Token   string        `json:"token" description:"写入Token"`
}

// PD的版本信息
type PlacementVersionInfo struct {
	Version           string `json:"version" description:"PD的版本"`
	LastHeartBeatTime int64  `json:"lastHeartBeatTime" description:"PD最后一次心跳时间"`
}

type Placement interface {
	// 判断服务器是否有效
	IsServerValid(serverId uint64) bool

	// 生成一个新的服务器ID, 服务器每次启动的时候都需要向PD去申请新的ID
	GenerateServerId() (uint64, error)

	// 获取一个新的ID, 可以提供比较频繁的调用
	GenerateNewSequence(sequenceType string, step int) uint64

	// 生成一个新的写入Token, 用来做Actor写入权限的判断
	GenerateNewToken() (*GenerateNewTokenResponse, error)

	// 注册当前服务器到PD里面去
	RegisterServer(info *PlacementActorHostInfo) uint64

	// 给当前服务器续约, 维持其生命
	KeepAliveServer(serverId uint64, leaseId uint64, load uint64) *PlacementKeepAliveResponse

	// 在内存中找Actor所在的服务器信息
	FindActorPositionInCache(request *PlacementFindActorPositionArgs) *PlacementActorPosition

	// 找到Actor所在的服务器信息
	FindActorPositon(request *PlacementFindActorPositionArgs) *PlacementActorPosition

	// 清空Actor的位置缓存
	ClearActorPositionCache(request *PlacementFindActorPositionArgs)

	// 获取PD的版本信息
	GetVersionInfo() *PlacementVersionInfo

	// 获取当前服务器的信息
	GetCurServerId() uint64

	// 获取服务器列表变动事件
	RegisterServerChangedEvent(onAddServer func(PlacementActorHostInfo), onRemoveServer func(PlacementActorHostInfo), onServerOffline func(PlacementActorHostInfo), onFatalError func(error))

	// 设置服务器的负载, 0表示无负载, 数字越大表示负载越大, -1表示服务器将要下线
	SetServerLoad(load uint64)

	// 开启轮训续约的异步任务
	StartPulling()

	// 停止轮训续约
	StopPulling()
}
