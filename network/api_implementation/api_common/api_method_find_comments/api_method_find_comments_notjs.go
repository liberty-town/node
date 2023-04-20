//go:build !wasm
// +build !wasm

package api_method_find_comments

import (
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodFindComments(r *http.Request, args *APIMethodFindCommentsRequest, reply *api_types.APIMethodFindListResult) error {

	results, err := federation_store.GetComments(args.Identity, args.Type, args.Start)
	if err != nil {
		return err
	}

	reply.Results = results
	return nil
}
