//go:build !wasm
// +build !wasm

package api_method_get_last_msg

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetLastMessage(r *http.Request, args *APIMethodGetLastMessageRequest, reply *api_types.APIMethodGetResult) error {

	found, err := federation_store.GetChatLastMessage(args.First, args.Second)
	if err != nil {
		return err
	}

	if len(found) == 0 {
		return errors.New("not found")
	}

	reply.Result = found
	return nil
}
