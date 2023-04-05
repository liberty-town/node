package network

import (
	"context"
	"liberty-town/node/network/connected_nodes"
	"liberty-town/node/network/known_nodes"
	"liberty-town/node/network/server/node_tcp"
	"liberty-town/node/network/websocks"
	"liberty-town/node/network/websocks/connection/advanced_connection_types"
	"liberty-town/node/pandora-pay/helpers/msgpack"
	"time"
)

type networkType struct {
}

var Network *networkType

func (this *networkType) ImportSeeds(seeds []string) error {
	return known_nodes.KnownNodes.Reset(seeds, true)
}

func (this *networkType) Send(name, data []byte, ctxDuration time.Duration) error {

	for {

		<-websocks.Websockets.ReadyCn.Load()
		list := connected_nodes.ConnectedNodes.AllList.Get()
		if len(list) > 0 {
			sock := list[0]
			if err := sock.Send(name, data, ctxDuration); err != nil {
				return err
			}
			return nil
		}
	}

}

func (this *networkType) SendJSON(name, data []byte, ctxDuration time.Duration) error {
	out, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}

	return this.Send(name, out, ctxDuration)
}

func (this *networkType) SendAwaitAnswer(name, data []byte, ctxParent context.Context, ctxDuration time.Duration) *advanced_connection_types.AdvancedConnectionReply {
	for {
		<-websocks.Websockets.ReadyCn.Load()
		list := connected_nodes.ConnectedNodes.AllList.Get()
		if len(list) > 0 {
			sock := list[0]
			result := sock.SendAwaitAnswer(name, data, ctxParent, ctxDuration)
			if result.Timeout {
				continue
			}
			return result
		}
	}
}

func SendJSONAwaitAnswer[T any](name []byte, data any, ctxParent context.Context, ctxDuration time.Duration) (*T, error) {

	out, err := msgpack.Marshal(data)
	if err != nil {
		return nil, err
	}

	for {
		<-websocks.Websockets.ReadyCn.Load()
		list := connected_nodes.ConnectedNodes.AllList.Get()
		if len(list) > 0 {
			sock := list[0]

			out := sock.SendAwaitAnswer(name, out, ctxParent, ctxDuration)
			if out.Err != nil {
				if out.Timeout {
					continue
				}
				return nil, out.Err
			}

			final := new(T)
			if err = msgpack.Unmarshal(out.Out, final); err != nil {
				return nil, err
			}
			return final, nil
		}
	}
}

func NewNetwork() error {

	if err := node_tcp.NewTcpServer(); err != nil {
		return err
	}

	Network = &networkType{}

	Network.continuouslyConnectingNewPeers()
	Network.continuouslyDownloadNetworkNodes()

	return nil
}
