package placement

import (
	"fmt"
)

type SequenceResponse struct {
	Id uint64 `json:"id" description:"新的Id"`
}

func (s SequenceResponse) String() string {
	return fmt.Sprintf("Id: %d", s.Id)
}

type GenerateNewTokenResponse struct {
	Token       string `json:"token" description:"新的Token"`
	InvalidTime int64  `json:"invalidTime" description:"Token的失效时间"`
}

func (g GenerateNewTokenResponse) String() string {
	return fmt.Sprintf("Token: %s, InvalidTime: %d", g.Token, g.InvalidTime)
}

type RegisterServerResponse struct {
	LeaseId uint64 `json:"leaseId" description:"租约Id"`
}

func (r RegisterServerResponse) String() string {
	return fmt.Sprintf("LeaseId: %d", r.LeaseId)
}

// PD服务器上
type PlacementHostInfo struct {
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

func (p PlacementHostInfo) String() string {
	return fmt.Sprintf("ServerId: %d, LeaseId: %d, Load: %d, StartTime: %d, TTL: %d, DeadTime: %d, Address: %s, Services: %v, Desc: %s, Labels: %v", p.ServerId, p.LeaseId, p.Load, p.StartTime, p.TTL, p.DeadTime, p.Address, p.Services, p.Desc, p.Labels)
}

// PD上最近发生的事件
type PlacementEvents struct {
	Time   uint64   `json:"time" description:"事件发生的时间"`
	Add    []uint64 `json:"add" description:"添加的服务器ID"`
	Remove []uint64 `json:"remove" description:"删除的服务器ID"`
}

func (p PlacementEvents) String() string {
	return fmt.Sprintf("Time: %d, Add: %v, Remove: %v", p.Time, p.Add, p.Remove)
}

// 服务器续约请求
type ServerKeepAliveArgs struct {
	ServerId uint64 `json:"serverId" description:"服务器唯一Id"`
	LeaseId  uint64 `json:"leaseId" description:"租约Id"`
	Load     uint64 `json:"load" description:"负载"`
}

func (s ServerKeepAliveArgs) String() string {
	return fmt.Sprintf("ServerId: %d, LeaseId: %d, Load: %d", s.ServerId, s.LeaseId, s.Load)
}

// 服务器续约返回
type ServerKeepAliveResponse struct {
	Hosts  map[uint64]PlacementHostInfo `json:"hosts" description:"每次续约PD会将所有的服务器信息下发"`
	Events []PlacementEvents            `json:"events" description:"服务器最近的事件(增减和删除)"`
}

// 服务对象定位的请求
type PlacementFindActorPositionArgs struct {
	ActorType string `json:"actorType" description:"actor类型"`
	Id        uint64 `json:"id" description:"id"`
	TTL       int64  `json:"ttl" description:"租约时间"`
}

func (p PlacementFindActorPositionArgs) String() string {
	return fmt.Sprintf("ActorType: %s, Id: %d, TTL: %d", p.ActorType, p.Id, p.TTL)
}

// 服务对象定位的返回
type PlacementActorPosition struct {
	ActorType  string `json:"actorType" description:"actor类型"`
	Id         uint64 `json:"id" description:"id"`
	TTL        int64  `json:"ttl" description:"租约时间"`
	CreateTime int64  `json:"createTime" description:"创建时间"`
	DeadTime   int64  `json:"deadTime" description:"租约过期时间"`
	ServerId   uint64 `json:"serverId" description:"宿主的唯一id"`
	Token      string `json:"token" description:"写入Token"`
}

func (p PlacementActorPosition) String() string {
	return fmt.Sprintf("ActorType: %s, Id: %d, TTL: %d, CreateTime: %d, DeadTime: %d, ServerId: %d, Token: %s", p.ActorType, p.Id, p.TTL, p.CreateTime, p.DeadTime, p.ServerId, p.Token)
}

// Actor续约请求
type ActorKeepAliveArgs struct {
	ActorType string `json:"actorType" description:"服务类型"`
	Id        uint64 `json:"id" description:"id"`
	Token     string `json:"token" description:"写入Token"`
}

func (a ActorKeepAliveArgs) String() string {
	return fmt.Sprintf("ActorType: %s, Id: %d, Token: %s", a.ActorType, a.Id, a.Token)
}

// Actor续约返回
type ActorKeepAliveResponse struct {
	ActorType  string `json:"actorType" description:"服务类型"`
	Id         uint64 `json:"id" description:"id"`
	CreateTime int64  `json:"createTime" description:"创建时间"`
	DeadTime   int64  `json:"deadTime" description:"租约过期时间"`
}

func (a ActorKeepAliveResponse) String() string {
	return fmt.Sprintf("ActorType: %s, Id: %d, CreateTime: %d, DeadTime: %d", a.ActorType, a.Id, a.CreateTime, a.DeadTime)
}

// PD的版本信息
type PlacementVersionInfo struct {
	Version           string `json:"version" description:"PD的版本"`
	LastHeartBeatTime int64  `json:"lastHeartBeatTime" description:"PD最后一次心跳时间"`
}

func (p PlacementVersionInfo) String() string {
	return fmt.Sprintf("Version: %s, LastHeartBeatTime: %d", p.Version, p.LastHeartBeatTime)
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
	RegisterServer(info *PlacementHostInfo) uint64

	// 获取服务器的信息
	GetServerInfo(serverId uint64) *PlacementHostInfo

	// 给当前服务器续约, 维持其生命
	KeepAliveServer(serverId uint64, leaseId uint64, load uint64) *ServerKeepAliveResponse

	// 在内存中找Actor所在的服务器信息
	FindActorPositionInCache(request *PlacementFindActorPositionArgs) *PlacementActorPosition

	// 找到Actor所在的服务器信息
	FindActorPositon(request *PlacementFindActorPositionArgs) *PlacementActorPosition

	// 续约Actor的生命
	ActorKeepAliveActor(actorType string, id uint64, token string) *ActorKeepAliveResponse

	// 清空Actor的位置缓存
	ClearActorPositionCache(request *PlacementFindActorPositionArgs)

	// 获取PD的版本信息
	GetVersionInfo() *PlacementVersionInfo

	// 获取当前服务器的信息
	GetCurServerId() uint64

	// 获取服务器列表变动事件
	RegisterServerChangedEvent(onAddServer func(PlacementHostInfo), onRemoveServer func(PlacementHostInfo), onServerOffline func(PlacementHostInfo), onFatalError func(error))

	// 设置服务器的负载, 0表示无负载, 数字越大表示负载越大, -1表示服务器将要下线
	SetServerLoad(load uint64)

	// 开启轮训续约的异步任务
	StartPulling()

	// 停止轮训续约
	StopPulling()
}
