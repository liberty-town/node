package node_http_rpc

import (
	"github.com/gorilla/rpc"
	"liberty-town/node/network/api_implementation/api_common"
	"net/http"
)

func InitializeRPC(apiCommon *api_common.APICommon) (err error) {

	s := rpc.NewServer()

	s.RegisterCodec(NewUpCodec(), "application/json")
	if err = s.RegisterService(apiCommon, "api"); err != nil {
		return
	}

	http.Handle("/rpc/api/v1", s)

	return
}
