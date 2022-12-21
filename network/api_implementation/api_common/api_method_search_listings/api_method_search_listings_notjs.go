//go:build !wasm
// +build !wasm

package api_method_search_listings

import (
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodSearchListings(r *http.Request, args *APIMethodSearchListingsRequest, reply *api_types.APIMethodFindListResult) error {

	results, err := federation_store.SearchListings(args.Query, args.Type, args.QueryType, args.Start)
	if err != nil {
		return err
	}

	reply.Results = results
	return nil
}
