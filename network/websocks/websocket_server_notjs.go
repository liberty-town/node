//go:build !js
// +build !js

package websocks

import (
	"liberty-town/node/network/connected_nodes"
	"liberty-town/node/network/known_nodes"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks/websock"
	"liberty-town/node/pandora-pay/helpers/recovery"
	"net/http"
	"sync/atomic"
)

func (this *websocketsType) HandleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	if atomic.LoadInt64(&connected_nodes.ConnectedNodes.ServerSockets) >= network_config.WEBSOCKETS_NETWORK_SERVER_MAX {
		http.Error(w, "Too many websockets", 400)
		return
	}

	c, err := websock.Upgrade(w, r)
	if err != nil {
		return
	}

	conn, err := Websockets.NewConnection(c, r.RemoteAddr, nil, true)
	if err != nil {
		return
	}

	if conn.Handshake.URL != "" {
		conn.KnownNode, err = known_nodes.KnownNodes.AddKnownNode(conn.Handshake.URL, false)
		if conn.KnownNode != nil {
			recovery.SafeGo(conn.IncreaseKnownNodeScore)
		}
	}

}
