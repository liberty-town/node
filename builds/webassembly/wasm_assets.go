package main

import (
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/config/globals"
	"syscall/js"
)

func appGetAssets(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		return webassembly_utils.ConvertJSONBytes(globals.Assets)
	})
}

func appConvertCurrencyToAsset(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			Currency string `json:"currency"`
			Asset    string `json:"asset"`
			Amount   uint64 `json:"amount"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		result, err := globals.ConvertCurrencyToAsset(req.Currency, req.Asset, req.Amount)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Result uint64 `json:"result"`
		}{result})
	})
}

func appConvertAssetToCurrency(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			Currency string `json:"currency"`
			Asset    string `json:"asset"`
			Amount   uint64 `json:"amount"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		result, err := globals.ConvertAssetToCurrency(req.Asset, req.Currency, req.Amount)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Result uint64 `json:"result"`
		}{result})
	})
}
