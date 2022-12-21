package network_config

import (
	"liberty-town/node/config/arguments"
	"liberty-town/node/network/network_config/network_config_auth"
	"strconv"
	"time"
)

var (
	WEBSOCKETS_NETWORK_CLIENTS_MAX       = int64(50)
	WEBSOCKETS_NETWORK_SERVER_MAX        = int64(500)
	NETWORK_ADDRESS_URL_STRING           string
	NETWORK_WEBSOCKET_ADDRESS_URL_STRING string
	NETWORK_KNOWN_NODES_LIMIT            int32 = 5000
	NETWORK_KNOWN_NODES_LIST_RETURN            = 100
	NETWORK_ENABLE_SUBSCRIPTIONS               = true
	NETWORK_CONNECTIONS_READY_THRESHOLD        = int64(1)
	STATIC_FILES                               = map[string]string{}
)

const (
	WEBSOCKETS_MAX_READ_THREADS                   = 5
	WEBSOCKETS_PONG_WAIT                          = 60 * time.Second // Time allowed to read the next pong message from the peer.
	WEBSOCKETS_PING_INTERVAL                      = (WEBSOCKETS_PONG_WAIT * 8) / 10
	WEBSOCKETS_MAX_READ                           = 50 * 1024
	WEBSOCKETS_MAX_SUBSCRIPTIONS                  = 30
	WEBSOCKETS_INCREASE_KNOWN_NODE_SCORE_INTERVAL = 1 * time.Minute
	WEBSOCKETS_CONCURRENT_NEW_CONENCTIONS         = 5
	WEBSOCKETS_CONCURRENT_SYNC_CONNECTIONS        = 5
	WEBSOCKETS_TIMEOUT                            = 15 * time.Second //seconds
	REQUEST_TIMEOUT                               = 15 * time.Second
)

func InitConfig() (err error) {

	if arguments.Arguments["--tcp-max-clients"] != nil {
		if WEBSOCKETS_NETWORK_CLIENTS_MAX, err = strconv.ParseInt(arguments.Arguments["--tcp-max-clients"].(string), 10, 64); err != nil {
			return
		}
	}

	if arguments.Arguments["--tcp-max-server-sockets"] != nil {
		if WEBSOCKETS_NETWORK_SERVER_MAX, err = strconv.ParseInt(arguments.Arguments["--tcp-max-server-sockets"].(string), 10, 64); err != nil {
			return
		}
	}

	if arguments.Arguments["--tcp-connections-ready=threshold"] != nil {
		if NETWORK_CONNECTIONS_READY_THRESHOLD, err = strconv.ParseInt(arguments.Arguments["--tcp-connections-ready"].(string), 10, 64); err != nil {
			return
		}
	}

	if err = network_config_auth.InitConfig(); err != nil {
		return
	}

	return
}
