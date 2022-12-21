//go:build !wasm
// +build !wasm

package api_method_find_reviews

import (
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodFindReviews(r *http.Request, args *APIMethodFindReviewsRequest, reply *api_types.APIMethodFindListResult) error {

	results, err := federation_store.GetReviews(args.Identity, args.Type, args.Start)
	if err != nil {
		return err
	}

	reply.Results = results
	return nil
}
