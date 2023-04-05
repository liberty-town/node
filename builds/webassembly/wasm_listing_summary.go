package main

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/listings_summaries"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/settings"
	"syscall/js"
)

func listingSummaryGet(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Listing *addresses.Address `json:"listing,omitempty"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		var listingSummary *listings_summaries.ListingSummary

		if err := federation_network.FetchData[api_types.APIMethodGetResult]("get-listing-summary", &api_types.APIMethodGetRequest{
			req.Listing.Encoded,
		}, func(data *api_types.APIMethodGetResult, b *connection.AdvancedConnection) bool {

			if data == nil || data.Result == nil {
				return true
			}
			temp := &listings_summaries.ListingSummary{}
			if err := temp.Deserialize(advanced_buffers.NewBufferReader(data.Result)); err != nil {
				return true
			}
			if temp.Validate() != nil || temp.ValidateSignatures() != nil {
				return true
			}

			if temp.IsBetter(listingSummary) {
				listingSummary = temp
			}
			return true
		}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(listingSummary)

	})
}

func listingSummaryStore(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		var err error

		req := &struct {
			Listing *addresses.Address `json:"listing"`
			Total   uint64             `json:"Total"`
			Count   uint64             `json:"Count"`
			Amount  uint64             `json:"Amount"`
		}{}

		if err = webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		it := &listings_summaries.ListingSummary{
			listings_summaries.LISTING_SUMMARY_VERSION,
			f.Federation.Ownership.Address,
			req.Listing,
			req.Total,
			req.Count,
			req.Amount,
			nil,
			&ownership.Ownership{},
		}

		if it.Validation, _, err = federationValidate(f.Federation, it.GetMessageForSigningValidator, args[1], nil); err != nil {
			return nil, err
		}

		if err = it.Signer.Sign(settings.Settings.Load().Account.PrivateKey, it.GetMessageForSigningSigner); err != nil {
			return nil, err
		}

		if err = it.Validate(); err != nil {
			return nil, err
		}

		results := 0

		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-listing-summary", &api_types.APIMethodStoreRequest{helpers.SerializeToBytes(it)}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
			if a != nil && a.Result {
				results++
			}
			return true
		}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			ListingSummary *listings_summaries.ListingSummary `json:"listingSummary"`
			Results        int                                `json:"results"`
		}{it, results})

	})
}
