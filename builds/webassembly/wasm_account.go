package main

import (
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/accounts"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/settings"
	"sync/atomic"
	"syscall/js"
)

func accountStore(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		var err error

		req := &struct {
			Description string `json:"description"`
			Country     uint64 `json:"country"`
		}{}

		if err = json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		s := settings.Settings.Load()

		it := &accounts.Account{
			accounts.ACCOUNT_VERSION,
			f.Federation.Ownership.Address,
			s.Account.Address,
			req.Description,
			req.Country,
			nil,
			&ownership.Ownership{},
		}

		if it.Validation, _, err = federationValidate(f.Federation, it.GetMessageForSigningValidator, args[1], nil); err != nil {
			return nil, err
		}

		if err = it.Ownership.Sign(s.Account.PrivateKey, it.GetMessageForSigningOwnership); err != nil {
			return nil, err
		}

		if err = it.Validate(); err != nil {
			return nil, err
		}

		results := atomic.Int32{}
		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-account", &api_types.APIMethodStoreRequest{helpers.SerializeToBytes(it)}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
			if a != nil && a.Result {
				results.Add(1)
			}
			return true
		}, &generics.Map[string, bool]{}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Account *accounts.Account `json:"account"`
			Results int32             `json:"results"`
		}{it, results.Load()})

	})
}

func accountGet(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Account *addresses.Address `json:"account,omitempty"`
		}{}

		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		if req.Account == nil {
			account := settings.Settings.Load().Account
			req.Account = account.Address
		}

		var account *accounts.Account

		if err := federation_network.FetchData[api_types.APIMethodGetResult]("get-account", &api_types.APIMethodGetRequest{
			req.Account.Encoded,
		}, func(data *api_types.APIMethodGetResult, b *connection.AdvancedConnection) bool {
			if data == nil || data.Result == nil {
				return true
			}
			temp := &accounts.Account{}
			if err := temp.Deserialize(advanced_buffers.NewBufferReader(data.Result)); err != nil {
				return true
			}
			if temp.Validate() != nil || temp.ValidateSignatures() != nil {
				return true
			}
			if temp.IsBetter(account) {
				account = temp
			}
			return true
		}, &generics.Map[string, bool]{}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(account)

	})
}
