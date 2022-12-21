//go:build wasm
// +build wasm

package node_http

import (
	"liberty-town/node/network/api_implementation/api_common"
	"liberty-town/node/network/api_implementation/api_websockets"
	"liberty-town/node/network/websocks"
)

type httpServerType struct {
	ApiWebsockets *api_websockets.APIWebsockets
}

var HttpServer *httpServerType

func NewHttpServer() error {

	apiCommon, err := api_common.NewAPICommon()
	if err != nil {
		return err
	}

	apiWebsockets := api_websockets.NewWebsocketsAPI(apiCommon)
	websocks.NewWebsockets(apiWebsockets.GetMap)

	HttpServer = &httpServerType{
		apiWebsockets,
	}

	return nil
}
