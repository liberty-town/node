package known_nodes_sync

import (
	"liberty-town/node/network/api_implementation/api_common"
	"liberty-town/node/network/known_nodes"
	"liberty-town/node/network/websocks/connection"
)

type KnownNodesSyncType struct {
}

var KnownNodesSync *KnownNodesSyncType

func (self *KnownNodesSyncType) DownloadNetworkNodes(conn *connection.AdvancedConnection) error {

	data, err := connection.SendJSONAwaitAnswer[api_common.APINetworkNodesReply](conn, []byte("network/nodes"), nil, nil, 0)
	if err != nil {
		return err
	}

	if data == nil || data.Nodes == nil {
		return nil
	}

	for _, node := range data.Nodes {
		if node != nil {
			known_nodes.KnownNodes.AddKnownNode(node.URL, false)
		}
	}

	return nil
}

func init() {
	KnownNodesSync = &KnownNodesSyncType{}
}
