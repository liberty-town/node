package network

import (
	"liberty-town/node/gui"
	"liberty-town/node/network/banned_nodes"
	"liberty-town/node/network/connected_nodes"
	"liberty-town/node/network/known_nodes"
	"liberty-town/node/network/known_nodes/known_node"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks"
	"liberty-town/node/pandora-pay/helpers/recovery"
	"time"
)

func (network *networkType) continuouslyConnectingNewPeers() {

	for i := 0; i < network_config.WEBSOCKETS_CONCURRENT_NEW_CONENCTIONS; i++ {
		index := i
		recovery.SafeGo(func() {

			for {

				if websocks.Websockets.GetClients() >= network_config.WEBSOCKETS_NETWORK_CLIENTS_MAX {
					time.Sleep(500 * time.Millisecond)
					continue
				}

				var knownNode *known_node.KnownNodeScored
				if index == 0 {
					knownNode = known_nodes.KnownNodes.GetBestNotConnectedKnownNode()
				} else {
					knownNode = known_nodes.KnownNodes.GetRandomKnownNode()
				}
				if knownNode != nil {

					if _, loaded := connected_nodes.ConnectedNodes.AllAddresses.Load(knownNode.URL); loaded {
						time.Sleep(100 * time.Millisecond)
						continue
					}

					//gui.GUI.Log("connecting to", knownNode.URL, atomic.LoadInt32(&knownNode.Score))

					if banned_nodes.BannedNodes.IsBanned(knownNode.URL) {
						known_nodes.KnownNodes.DecreaseKnownNodeScore(knownNode, -10, false)
					} else {
						_, err := websocks.Websockets.NewWebsocketClient(knownNode)
						if err != nil {

							//gui.GUI.Error("error connecting", knownNode.URL, err)

							if err.Error() != "Already connected" {
								known_nodes.KnownNodes.DecreaseKnownNodeScore(knownNode, -20, false)
							}

						} else {
							gui.GUI.Log("connected to: " + knownNode.URL)
						}
					}
				}

				time.Sleep(100 * time.Millisecond)
			}
		})
	}

}
