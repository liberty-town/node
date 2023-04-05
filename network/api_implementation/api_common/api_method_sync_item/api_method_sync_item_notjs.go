//go:build !wasm
// +build !wasm

package api_method_sync_item

import (
	"liberty-town/node/federations/federation_store"
	"net/http"
)

func MethodSyncItem(r *http.Request, args *APIMethodSyncItemRequest, reply *APIMethodSyncItemResult) (err error) {
	reply.BetterScore, err = federation_store.GetSyncItem(args.Type, args.Key)
	return
}
