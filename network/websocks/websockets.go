package websocks

import (
	"context"
	"errors"
	"github.com/tevino/abool"
	msgpack "github.com/vmihailenco/msgpack/v5"
	"liberty-town/node/config"
	"liberty-town/node/config/globals"
	"liberty-town/node/gui"
	"liberty-town/node/network/banned_nodes"
	"liberty-town/node/network/connected_nodes"
	"liberty-town/node/network/known_nodes"
	"liberty-town/node/network/known_nodes/known_node"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/network/websocks/connection/advanced_connection_types"
	"liberty-town/node/network/websocks/websock"
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/pandora-pay/helpers/multicast"
	"liberty-town/node/pandora-pay/helpers/recovery"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
)

type SocketEvent struct {
	Type         string
	Conn         *connection.AdvancedConnection
	TotalSockets int64
}

type websocketsType struct {
	apiGetMap                    map[string]func(conn *connection.AdvancedConnection, values []byte) (any, error)
	UpdateNewConnectionMulticast *multicast.MulticastChannel[*connection.AdvancedConnection]
	subscriptions                *WebsocketSubscriptions
	UpdateSocketEventMulticast   *multicast.MulticastChannel[*SocketEvent]
	ReadyCn                      *generics.Value[chan struct{}]
	ReadyCnClosed                *abool.AtomicBool
}

var Websockets *websocketsType

func (websockets *websocketsType) GetClients() int64 {
	return atomic.LoadInt64(&connected_nodes.ConnectedNodes.Clients)
}

func (websockets *websocketsType) GetServerSockets() int64 {
	return atomic.LoadInt64(&connected_nodes.ConnectedNodes.ServerSockets)
}

func (websockets *websocketsType) GetAllSockets() []*connection.AdvancedConnection {
	return connected_nodes.ConnectedNodes.AllList.Get()
}

func (websockets *websocketsType) GetRandomSocket() *connection.AdvancedConnection {
	list := websockets.GetAllSockets()
	if len(list) > 0 {
		index := rand.Intn(len(list))
		return list[index]
	}
	return nil
}

func (websockets *websocketsType) Disconnect() int {
	list := websockets.GetAllSockets()
	for _, sock := range list {
		sock.Close()
	}
	return len(list)
}

func (websockets *websocketsType) Broadcast(name []byte, data []byte, consensusTypeAccepted map[config.NodeConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctxDuration time.Duration) {

	if exceptSocketUUID == advanced_connection_types.UUID_SKIP_ALL {
		return
	}

	all := websockets.GetAllSockets()

	for i, conn := range all {
		if conn.UUID != exceptSocketUUID && consensusTypeAccepted[conn.Handshake.Consensus] {
			go func(conn *connection.AdvancedConnection, i int) {
				conn.Send(name, data, ctxDuration)
			}(conn, i)
		}
	}

}

func (websockets *websocketsType) BroadcastAwaitAnswer(name, data []byte, consensusTypeAccepted map[config.NodeConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context, ctxDuration time.Duration) []*advanced_connection_types.AdvancedConnectionReply {

	if exceptSocketUUID == advanced_connection_types.UUID_SKIP_ALL {
		return nil
	}

	all := websockets.GetAllSockets()

	t := time.Now().Unix()
	index := rand.Int()
	gui.GUI.Log("Propagating", index, len(all), string(name), t)

	chans := make(chan *advanced_connection_types.AdvancedConnectionReply, len(all)+1)
	for i, conn := range all {
		if conn.UUID != exceptSocketUUID && consensusTypeAccepted[conn.Handshake.Consensus] {
			go func(conn *connection.AdvancedConnection, i int) {
				answer := conn.SendAwaitAnswer(name, data, ctx, ctxDuration)
				chans <- answer
			}(conn, i)
		} else {
			chans <- nil
		}
	}

	out := make([]*advanced_connection_types.AdvancedConnectionReply, len(all))
	for i := range all {
		out[i] = <-chans
		if out[i] != nil && out[i].Err != nil {
			gui.GUI.Error("Error propagating", index, out[i].Err, len(all), string(name), all[i].RemoteAddr, all[i].UUID, time.Now().Unix()-t)
		}
	}

	return out
}

func (websockets *websocketsType) BroadcastJSON(name []byte, data interface{}, consensusTypeAccepted map[config.NodeConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctxDuration time.Duration) {
	out, _ := msgpack.Marshal(data)
	websockets.Broadcast(name, out, consensusTypeAccepted, exceptSocketUUID, ctxDuration)
}

func (websockets *websocketsType) BroadcastJSONAwaitAnswer(name []byte, data interface{}, consensusTypeAccepted map[config.NodeConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context, ctxDuration time.Duration) []*advanced_connection_types.AdvancedConnectionReply {
	out, _ := msgpack.Marshal(data)
	return websockets.BroadcastAwaitAnswer(name, out, consensusTypeAccepted, exceptSocketUUID, ctx, ctxDuration)
}
func (websockets *websocketsType) closedConnection(conn *connection.AdvancedConnection) {

	if conn.KnownNode != nil {
		known_nodes.KnownNodes.MarkKnownNodeDisconnected(conn.KnownNode)
	}
	connected_nodes.ConnectedNodes.JustDisconnected(conn)

	conn.InitializedStatusMutex.Lock()

	if conn.InitializedStatus != connection.INITIALIZED_STATUS_INITIALIZED {
		conn.InitializedStatusMutex.Unlock()
		return
	}

	conn.InitializedStatus = connection.INITIALIZED_STATUS_CLOSED
	conn.InitializedStatusMutex.Unlock()

	totalSockets := connected_nodes.ConnectedNodes.Disconnected(conn)

	if network_config.NETWORK_ENABLE_SUBSCRIPTIONS {
		websockets.subscriptions.websocketClosedCn <- conn
	}

	globals.MainEvents.BroadcastEvent("sockets/totalSocketsChanged", totalSockets)
	websockets.UpdateSocketEventMulticast.Broadcast(&SocketEvent{"disconnected", conn, totalSockets})

	if totalSockets < network_config.NETWORK_CONNECTIONS_READY_THRESHOLD {
		if websockets.ReadyCnClosed.SetToIf(true, false) {
			websockets.ReadyCn.Store(make(chan struct{}))
		}
	}
}

func (websockets *websocketsType) increaseScoreKnownNode(knownNode *known_node.KnownNodeScored, delta int32, isServer bool) bool {
	return known_nodes.KnownNodes.IncreaseKnownNodeScore(knownNode, delta, isServer)
}

func (websockets *websocketsType) NewConnection(c *websock.Conn, remoteAddr string, knownNode *known_node.KnownNodeScored, connectionType bool) (*connection.AdvancedConnection, error) {

	conn, err := connection.NewAdvancedConnection(c, remoteAddr, knownNode, websockets.apiGetMap, connectionType, websockets.subscriptions.newSubscriptionCn, websockets.subscriptions.removeSubscriptionCn, websockets.closedConnection, websockets.increaseScoreKnownNode)
	if err != nil {
		return nil, err
	}

	if !connected_nodes.ConnectedNodes.JustConnected(conn, remoteAddr) {
		return nil, errors.New("Already connected")
	}

	recovery.SafeGo(conn.ReadPump)
	recovery.SafeGo(conn.SendPings)

	if knownNode != nil {
		known_nodes.KnownNodes.MarkKnownNodeConnected(knownNode)
		recovery.SafeGo(conn.IncreaseKnownNodeScore)
	}

	if err = websockets.InitializeConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func (websockets *websocketsType) InitializeConnection(conn *connection.AdvancedConnection) (err error) {

	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	out := conn.SendAwaitAnswer([]byte("handshake"), nil, nil, 0)

	if out.Err != nil {
		return errors.New("Error sending handshake")
	}
	if out.Out == nil {
		return errors.New("Handshake was not received")
	}

	handshakeReceived := &connection.ConnectionHandshake{}
	if err := msgpack.Unmarshal(out.Out, handshakeReceived); err != nil {
		return errors.New("Handshake received was invalid")
	}

	version, err := handshakeReceived.ValidateHandshake()
	if err != nil {
		return errors.New("Handshake is invalid")
	}

	if handshakeReceived.URL != "" && banned_nodes.BannedNodes.IsBanned(handshakeReceived.URL) {
		return errors.New("Socket is banned")
	}

	conn.Handshake = handshakeReceived
	conn.Version = version

	if conn.IsClosed.IsSet() {
		return
	}

	conn.InitializedStatusMutex.Lock()
	conn.InitializedStatus = connection.INITIALIZED_STATUS_INITIALIZED
	conn.InitializedStatusMutex.Unlock()

	totalSockets := connected_nodes.ConnectedNodes.ConnectedHandshakeValidated(conn)
	globals.MainEvents.BroadcastEvent("sockets/totalSocketsChanged", totalSockets)
	websockets.UpdateSocketEventMulticast.Broadcast(&SocketEvent{"connected", conn, totalSockets})
	websockets.UpdateNewConnectionMulticast.Broadcast(conn)

	if totalSockets >= network_config.NETWORK_CONNECTIONS_READY_THRESHOLD {
		cn := websockets.ReadyCn.Load()
		if websockets.ReadyCnClosed.SetToIf(false, true) {
			close(cn)
		}
	}

	return nil
}

func NewWebsockets(apiGetMap map[string]func(conn *connection.AdvancedConnection, values []byte) (any, error)) *websocketsType {

	Websockets = &websocketsType{
		apiGetMap,
		multicast.NewMulticastChannel[*connection.AdvancedConnection](),
		nil,
		multicast.NewMulticastChannel[*SocketEvent](),
		&generics.Value[chan struct{}]{},
		abool.NewBool(false),
	}

	Websockets.subscriptions = newWebsocketSubscriptions()
	Websockets.ReadyCn.Store(make(chan struct{}))

	recovery.SafeGo(func() {
		for {
			gui.GUI.InfoUpdate("sockets", strconv.FormatInt(atomic.LoadInt64(&connected_nodes.ConnectedNodes.Clients), 32)+" "+strconv.FormatInt(atomic.LoadInt64(&connected_nodes.ConnectedNodes.ServerSockets), 32))
			time.Sleep(1 * time.Second)
		}
	})

	return Websockets
}
