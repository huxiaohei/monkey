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
	Services       map[string]string `json:"services"`
	Desc           string            `json:"desc"`
	Labels         map[string]string `json:"labels"`
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
