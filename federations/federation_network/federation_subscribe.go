package federation_network

import (
	"liberty-town/node/config"
	"liberty-town/node/network/api_code/api_code_types"
	"liberty-town/node/network/websocks"
	"liberty-town/node/network/websocks/connection/advanced_connection_types"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/recovery"
	"liberty-town/node/settings"
)

//检查新通知
func SubscribeToChat() {
	recovery.SafeGo(func() {

		newConnectionCn := websocks.Websockets.UpdateNewConnectionMulticast.AddListener()
		defer websocks.Websockets.UpdateNewConnectionMulticast.RemoveChannel(newConnectionCn)

		for {

			newConn := <-newConnectionCn

			addr := settings.Settings.Load().Account.Address

			req := &api_code_types.APISubscriptionRequest{helpers.SerializeToBytes(addr), api_code_types.SUBSCRIPTION_CHAT_ACCOUNT, api_code_types.RETURN_SERIALIZED}
			newConn.SendJSON([]byte("sub"), req, 0)

		}

	})

	recovery.SafeGo(func() {

		changedCn := settings.ChangedEvents.AddListener()
		defer settings.ChangedEvents.RemoveChannel(changedCn)

		var old *settings.SettingsType

		for {

			s := <-changedCn

			if old != nil {
				addr := old.Account.Address

				req := &api_code_types.APISubscriptionRequest{helpers.SerializeToBytes(addr), api_code_types.SUBSCRIPTION_CHAT_ACCOUNT, api_code_types.RETURN_SERIALIZED}
				websocks.Websockets.BroadcastJSON([]byte("unsub"), req, map[config.NodeConsensusType]bool{config.NODE_CONSENSUS_TYPE_FULL: true}, advanced_connection_types.UUID_ALL, 0)

			}

			addr := s.Account.Address

			req := &api_code_types.APISubscriptionRequest{helpers.SerializeToBytes(addr), api_code_types.SUBSCRIPTION_CHAT_ACCOUNT, api_code_types.RETURN_SERIALIZED}
			websocks.Websockets.BroadcastJSON([]byte("sub"), req, map[config.NodeConsensusType]bool{config.NODE_CONSENSUS_TYPE_FULL: true}, advanced_connection_types.UUID_ALL, 0)

			old = s
		}

	})

}
