package conf

import (
	"encoding/json"
	"fmt"
	"os"
)

type GatewayConf struct {
	PdAddress      string            `json:"pdAddress"`
	ListenAddress  string            `json:"listenAddress"`
	ListenPath     string            `json:"listenPath"`
	Load           uint64            `json:"load"`
	ServerTTL      int64             `json:"serverTTL"`
	ReceiveTimeout int64             `json:"receiveTimeout"`
	Services       map[string]string `json:"services" description:"服务器能提供的Actor对象类型"`
	Desc           string            `json:"desc"`
	Labels         map[string]string `json:"labels"`
}

func (g GatewayConf) String() string {
	return fmt.Sprintf("PdAddress: %s, ListenAddress: %s, ListenPath: %s, Load: %d, ServerTTL: %d, ReceiveTimeout: %d, Services: %s, Desc: %s, Labels: %v",
		g.PdAddress, g.ListenAddress, g.ListenPath, g.Load, g.ServerTTL, g.ReceiveTimeout, g.Services, g.Desc, g.Labels)
}

func NewGatewayConf(path string) (*GatewayConf, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	conf := &GatewayConf{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(conf); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return conf, nil
}
