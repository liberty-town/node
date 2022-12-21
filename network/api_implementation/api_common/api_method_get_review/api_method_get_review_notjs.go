//go:build !wasm
// +build !wasm

package api_method_get_review

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetReview(r *http.Request, args *api_types.APIMethodGetRequest, reply *api_types.APIMethodGetResult) error {

	review, err := federation_store.GetReview(args.Identity)
	if err != nil {
		return err
	}

	if len(review) == 0 {
		return errors.New("not found")
	}

	reply.Result = review
	return nil
}
