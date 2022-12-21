package api_code_websockets

import (
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks/connection"
)

func Handshake(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	var fed *addresses.Address
	f := federation_serve.ServeFederation.Load()
	if f != nil {
		fed = f.Federation.Ownership.Address
	}
	return &connection.ConnectionHandshake{config.NAME, config.VERSION_STRING, config.NETWORK_SELECTED, config.NODE_CONSENSUS, fed, network_config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING}, nil
}
