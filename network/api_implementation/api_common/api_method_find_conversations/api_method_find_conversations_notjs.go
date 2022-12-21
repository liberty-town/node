//go:build !wasm
// +build !wasm

package api_method_find_conversations

import (
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodFindConversations(r *http.Request, args *APIMethodFindConversationsRequest, reply *api_types.APIMethodFindListResult) error {

	results, err := federation_store.GetChatConversations(args.First, args.Start)
	if err != nil {
		return err
	}

	reply.Results = results
	return nil
}
