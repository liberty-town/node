package api_method_ping

import (
	"net/http"
)

type APIPingReply struct {
	Ping string `json:"ping" msgpack:"ping"`
}

func GetPing(r *http.Request, args *struct{}, reply *APIPingReply) error {
	reply.Ping = "pong"
	return nil
}
