//go:build !wasm
// +build !wasm

package api_method_get_listing

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetListing(r *http.Request, args *api_types.APIMethodGetRequest, reply *api_types.APIMethodGetResult) error {

	listing, err := federation_store.GetListing(args.Identity)
	if err != nil {
		return err
	}

	if listing == nil {
		return errors.New("not found")
	}

	reply.Result = listing
	return nil
}
