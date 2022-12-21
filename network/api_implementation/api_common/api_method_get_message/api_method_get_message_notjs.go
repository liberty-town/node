//go:build !wasm
// +build !wasm

package api_method_get_message

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetMessage(r *http.Request, args *api_types.APIMethodGetRequest, reply *api_types.APIMethodGetResult) error {

	msg, err := federation_store.GetChatMessage(args.Identity)
	if err != nil {
		return err
	}

	if len(msg) == 0 {
		return errors.New("not found")
	}

	reply.Result = msg
	return nil
}
