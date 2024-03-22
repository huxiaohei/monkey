package main

import (
	"fmt"
	"monkey/conf"
	"monkey/gateway"
)

func main() {
	gatewayConf, err := conf.NewGatewayConf("conf.json")
	if err != nil {
		panic(err)
	}
	gateway.Start(gatewayConf)

	fmt.Println("Gateway started")
}
