//go:build !wasm
// +build !wasm

package api_method_get_listing_data

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetListingData(r *http.Request, args *api_types.APIMethodGetRequest, reply *APIMethodGetListingDataReply) error {

	found, accountSummary, listingSummary, err := federation_store.GetListingData(args.Identity)
	if err != nil {
		return err
	}

	if len(found) == 0 {
		return errors.New("not found")
	}

	reply.Listing = found
	reply.AccountSummary = accountSummary
	reply.ListingSummary = listingSummary

	return nil
}
