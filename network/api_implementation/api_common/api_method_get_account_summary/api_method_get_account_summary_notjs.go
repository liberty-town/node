//go:build !wasm
// +build !wasm

package api_method_get_account_summary

import (
	"errors"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/http"
)

func MethodGetAccountSummary(r *http.Request, args *api_types.APIMethodGetRequest, reply *api_types.APIMethodGetResult) error {

	accountSummary, err := federation_store.GetAccountSummary(args.Identity)
	if err != nil {
		return err
	}

	if len(accountSummary) == 0 {
		return errors.New("not found")
	}

	reply.Result = accountSummary
	return nil
}
