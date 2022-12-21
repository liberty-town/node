package websocks

import (
	"liberty-town/node/network/known_nodes/known_node"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/network/websocks/websock"
)

type WebsocketClient struct {
	knownNode *known_node.KnownNodeScored
	conn      *connection.AdvancedConnection
}

func (this *websocketsType) NewWebsocketClient(knownNode *known_node.KnownNodeScored) (*WebsocketClient, error) {

	wsClient := &WebsocketClient{
		knownNode, nil,
	}

	c, err := websock.Dial(knownNode.URL)
	if err != nil {
		return nil, err
	}

	if wsClient.conn, err = this.NewConnection(c, knownNode.URL, knownNode, false); err != nil {
		return nil, err
	}

	return wsClient, nil
}
