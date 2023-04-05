//go:build !wasm
// +build !wasm

package api_method_search_threads

import (
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodSearchThreads(r *http.Request, args *APIMethodSearchThreadsRequest, reply *api_types.APIMethodFindListResult) error {

	results, err := federation_store.SearchThreads(args.Query, args.Type, args.Start)
	if err != nil {
		return err
	}

	reply.Results = results
	return nil
}
