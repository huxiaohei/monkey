package main

import (
	"encoding/json"
	"fmt"
	"log"
	"monkey/gateway/protos"
	"monkey/utils"
	"net/url"

	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8001", Path: "/ws"}
	fmt.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	msg := protos.FirstPacket{
		UserId:     10001,
		Token:      "abcde",
		ServerType: "IPlayer",
		ClientTs:   utils.GetNowSec(),
		MsgSeq:     0,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("json marshal error: ", err)
		return
	}
	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Println("write:", err)
		return
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		return
	}
	log.Printf("recv: %s", message)

	utils.SleepMs(10)

}
