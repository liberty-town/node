package api_websockets

import (
	"liberty-town/node/config"
	"liberty-town/node/network/api_code/api_code_websockets"
	"liberty-town/node/network/api_implementation/api_common"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_fed"
	"liberty-town/node/network/api_implementation/api_common/api_method_ping"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_fed"
	"liberty-town/node/network/websocks/connection"
)

type APIWebsockets struct {
	GetMap    map[string]func(conn *connection.AdvancedConnection, values []byte) (interface{}, error)
	apiCommon *api_common.APICommon
}

var ConfigureAPIRoutes func(api *APIWebsockets)

func NewWebsocketsAPI(apiCommon *api_common.APICommon) *APIWebsockets {

	api := &APIWebsockets{
		nil,
		apiCommon,
	}

	api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) (interface{}, error){
		"":              api_code_websockets.Handle[struct{}, api_common.APIInfoReply](api.apiCommon.GetInfo),
		"ping":          api_code_websockets.Handle[struct{}, api_method_ping.APIPingReply](api_method_ping.GetPing),
		"network/nodes": api_code_websockets.Handle[struct{}, api_common.APINetworkNodesReply](api.apiCommon.GetNetworkNodes),
		//below are ONLY websockets API
		"handshake": api_code_websockets.Handshake,
		"login":     api_code_websockets.Login,
		"logout":    api_code_websockets.Logout,
		"sub":       api_code_websockets.Subscribe,
		"unsub":     api_code_websockets.Unsubscribe,
	}

	if config.NODE_CONSENSUS == config.NODE_CONSENSUS_TYPE_APP {
		api.GetMap["sub/notify"] = api_code_websockets.SubscribedNotificationReceived
	}

	api.initApi()

	api.GetMap["sync-fed"] = api_code_websockets.Handle[api_method_sync_fed.APIMethodSyncFedRequest, api_method_sync_fed.APIMethodSyncFedResult](api_method_sync_fed.MethodSyncFed)
	api.GetMap["get-fed"] = api_code_websockets.Handle[api_method_get_fed.APIMethodGetFedRequest, api_method_get_fed.APIMethodGetFedResult](api_method_get_fed.MethodGetFed)

	if ConfigureAPIRoutes != nil {
		ConfigureAPIRoutes(api)
	}

	return api
}
