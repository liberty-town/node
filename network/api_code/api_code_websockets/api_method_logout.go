package api_code_websockets

import (
	"liberty-town/node/network/websocks/connection"
)

type APILogoutReply struct {
	Status bool `json:"status" msgpack:""`
}

func Logout(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	reply := &APILogoutReply{}

	if conn.Authenticated.IsNotSet() {
		return reply, nil
	}

	conn.Authenticated.UnSet()
	reply.Status = true

	return reply, nil
}
