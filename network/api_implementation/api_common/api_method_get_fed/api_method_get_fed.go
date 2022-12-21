package api_method_get_fed

import (
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/pandora-pay/helpers"
	"net/http"
)

type APIMethodGetFedRequest struct {
}

type APIMethodGetFedResult struct {
	Federation []byte `json:"federation" msgpack:"federation"`
}

func MethodGetFed(r *http.Request, args *APIMethodGetFedRequest, reply *APIMethodGetFedResult) (err error) {
	fed := federation_serve.ServeFederation.Load()
	reply.Federation = helpers.SerializeToBytes(fed.Federation)
	return
}
