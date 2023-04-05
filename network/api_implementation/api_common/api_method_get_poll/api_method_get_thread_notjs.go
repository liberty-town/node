//go:build !wasm
// +build !wasm

package api_method_get_poll

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetPoll(r *http.Request, args *api_types.APIMethodGetRequest, reply *api_types.APIMethodGetResult) error {

	poll, err := federation_store.GetData("polls:", args.Identity)
	if err != nil {
		return err
	}

	if len(poll) == 0 {
		return errors.New("not found")
	}

	reply.Result = poll
	return nil
}
