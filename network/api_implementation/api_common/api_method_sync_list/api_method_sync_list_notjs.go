//go:build !wasm
// +build !wasm

package api_method_sync_list

import (
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/federations/federation_store"
	"net/http"
)

type APIMethodSyncListRequest struct {
	Type sync_type.SyncVersion `json:"type" msgpack:"type"`
}

type APIMethodSyncListResult struct {
	Keys         []string `json:"keys" msgpack:"keys"`
	BetterScores []uint64 `json:"betterScores" msgpack:"betterScores"`
}

func MethodSyncList(r *http.Request, args *APIMethodSyncListRequest, reply *APIMethodSyncListResult) (err error) {
	reply.Keys, reply.BetterScores, err = federation_store.GetSyncList(args.Type, 20)
	return
}
