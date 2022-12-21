package api_method_sync_fed

import (
	"liberty-town/node/federations/federation_serve"
	"net/http"
)

type APIMethodSyncFedRequest struct {
}

type APIMethodSyncFedResult struct {
	BetterScore uint64 `json:"betterScore" msgpack:"betterScore"`
}

func MethodSyncFed(r *http.Request, args *APIMethodSyncFedRequest, reply *APIMethodSyncFedResult) (err error) {
	fed := federation_serve.ServeFederation.Load()
	reply.BetterScore = fed.Federation.GetBetterScore()
	return
}
