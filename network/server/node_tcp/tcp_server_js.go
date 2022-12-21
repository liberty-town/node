//go:build wasm
// +build wasm

package node_tcp

import (
	"liberty-town/node/network/server/node_http"
)

type tcpServerType struct {
}

var TcpServer *tcpServerType

func NewTcpServer() error {
	TcpServer = &tcpServerType{}
	return node_http.NewHttpServer()
}
