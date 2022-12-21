package api_code_websockets

import (
	msgpack "github.com/vmihailenco/msgpack/v5"
	"liberty-town/node/network/api_code/api_code_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers/multicast"
	"net/http"
)

var SubscriptionNotifications *multicast.MulticastChannel[*api_code_types.APISubscriptionNotification]

func HandleAuthenticated[T any, B any](callback func(r *http.Request, args *T, reply *B, authenticated bool) error) func(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	return func(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
		args := new(T)
		if err := msgpack.Unmarshal(values, args); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply, conn.Authenticated.IsSet())
	}
}

func Handle[T any, B any](callback func(r *http.Request, args *T, reply *B) error) func(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	return func(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
		args := new(T)
		if err := msgpack.Unmarshal(values, args); err != nil {
			return nil, err
		}
		reply := new(B)
		return reply, callback(nil, args, reply)
	}
}

func init() {
	SubscriptionNotifications = multicast.NewMulticastChannel[*api_code_types.APISubscriptionNotification]()
}
