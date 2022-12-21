package main

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/accounts_summaries"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/settings"
	"syscall/js"
)

func accountSummaryGet(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			Account *addresses.Address `json:"account,omitempty"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		if req.Account == nil {
			req.Account = settings.Settings.Load().Account.Address
		}

		var accountSummary *accounts_summaries.AccountSummary

		if err := federation_network.FetchData[api_types.APIMethodGetResult]("get-account-summary", &api_types.APIMethodGetRequest{
			req.Account.Encoded,
		}, func(data *api_types.APIMethodGetResult, b *connection.AdvancedConnection) bool {

			if data == nil || data.Result == nil {
				return true
			}
			temp := &accounts_summaries.AccountSummary{}
			if err := temp.Deserialize(advanced_buffers.NewBufferReader(data.Result)); err != nil {
				return true
			}
			if temp.Validate() != nil || temp.ValidateSignatures() != nil {
				return true
			}

			if temp.IsBetter(accountSummary) {
				accountSummary = temp
			}
			return true
		}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(accountSummary)

	})
}

func accountSummaryStore(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		var err error

		req := &struct {
			Identity        *addresses.Address `json:"identity"`
			SalesTotal      uint64             `json:"salesTotal"`
			SalesCount      uint64             `json:"salesCount"`
			SalesAmount     uint64             `json:"salesAmount"`
			PurchasesTotal  uint64             `json:"purchasesTotal"`
			PurchasesCount  uint64             `json:"purchasesCount"`
			PurchasesAmount uint64             `json:"purchasesAmount"`
		}{}

		if err = webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		it := &accounts_summaries.AccountSummary{
			accounts_summaries.ACCOUNT_SUMMARY_VERSION,
			f.Federation.Ownership.Address,
			req.Identity,
			req.SalesTotal,
			req.SalesCount,
			req.SalesAmount,
			req.PurchasesTotal,
			req.PurchasesCount,
			req.PurchasesAmount,
			nil,
			&ownership.Ownership{},
		}

		if it.Validation, err = federationValidate(f.Federation, it.GetMessageForSigningValidator, args[1]); err != nil {
			return nil, err
		}

		if err = it.Signer.Sign(settings.Settings.Load().Account.PrivateKey, it.GetMessageForSigningSigner); err != nil {
			return nil, err
		}

		if err = it.Validate(); err != nil {
			return nil, err
		}

		results := 0

		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-account-summary", api_types.APIMethodStoreRequest{helpers.SerializeToBytes(it)}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
			if a != nil && a.Result {
				results++
			}
			return true
		}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			AccountSummary *accounts_summaries.AccountSummary `json:"accountSummary"`
			Results        int                                `json:"results"`
		}{it, results})

	})
}
