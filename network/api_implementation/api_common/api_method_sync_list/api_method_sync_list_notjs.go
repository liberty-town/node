//go:build !wasm
// +build !wasm

package api_method_sync_list

import (
	"liberty-town/node/federations/federation_store"
	"net/http"
)

func MethodSyncList(r *http.Request, args *APIMethodSyncListRequest, reply *APIMethodSyncListResult) (err error) {
	reply.Keys, reply.BetterScores, err = federation_store.GetSyncList(args.Type, 20)
	return
}
