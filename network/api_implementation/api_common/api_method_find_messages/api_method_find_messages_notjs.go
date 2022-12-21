//go:build !wasm
// +build !wasm

package api_method_find_messages

import (
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodFindMessages(r *http.Request, args *APIMethodFindMessagesRequest, reply *api_types.APIMethodFindListResult) error {

	results, err := federation_store.GetChatMessages(args.First, args.Second, args.Start)
	if err != nil {
		return err
	}

	reply.Results = results
	return nil
}
