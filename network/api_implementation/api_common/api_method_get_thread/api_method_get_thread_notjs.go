//go:build !wasm
// +build !wasm

package api_method_get_thread

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetThread(r *http.Request, args *api_types.APIMethodGetRequest, reply *api_types.APIMethodGetResult) error {

	thread, err := federation_store.GetData("threads:", args.Identity)
	if err != nil {
		return err
	}

	if len(thread) == 0 {
		return errors.New("not found")
	}

	reply.Result = thread
	return nil
}
