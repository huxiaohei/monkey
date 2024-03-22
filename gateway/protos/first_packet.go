package protos

import "fmt"

type FirstPacket struct {
	UserId     uint64 `json:"userId"`
	Token      string `json:"token"`
	ServerType string `json:"serverType"`
	ClientTs   int64  `json:"clientTs"`
}

func (fp *FirstPacket) String() string {
	return fmt.Sprintf("userId: %d, token: %s, serverType: %s, clientTs: %d", fp.UserId, fp.Token, fp.ServerType, fp.ClientTs)
}

type FirstPacketResponse struct {
	UserId       uint64 `json:"userId"`
	ReloginToken string `json:"reloginToken"`
	Nonce        string `json:"nonce"`
	ServerTs     int64  `json:"serverTs"`
}

func (fpr *FirstPacketResponse) String() string {
	return fmt.Sprintf("userId: %d, reloginToken: %s, nonce: %s, serverTs: %d", fpr.UserId, fpr.ReloginToken, fpr.Nonce, fpr.ServerTs)
}

func NewFirstPacketResponse(userId uint64, reloginToken string, nonce string, serverTs int64) *FirstPacketResponse {
	return &FirstPacketResponse{
		UserId:       userId,
		ReloginToken: reloginToken,
		Nonce:        nonce,
		ServerTs:     serverTs,
	}
}

type SessionCloseResponse struct {
	UserId uint64 `json:"userId"`
	Reason string `json:"reason"`
}

func NewSessionCloseResponse(userId uint64, reason string) *SessionCloseResponse {
	return &SessionCloseResponse{
		UserId: userId,
		Reason: reason,
	}
}
