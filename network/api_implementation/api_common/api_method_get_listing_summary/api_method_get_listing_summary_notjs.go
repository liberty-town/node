//go:build !wasm
// +build !wasm

package api_method_get_listing_summary

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetListingSummary(r *http.Request, args *api_types.APIMethodGetRequest, reply *api_types.APIMethodGetResult) error {

	listingSummary, err := federation_store.GetData("listings_summaries:", args.Identity)
	if err != nil {
		return err
	}

	if listingSummary == nil {
		return errors.New("not found")
	}

	reply.Result = listingSummary
	return nil
}
