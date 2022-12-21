package main

import (
	"errors"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/config"
	pandora_pay_addresses "liberty-town/node/pandora-pay/addresses"
	"syscall/js"
)

func pandoraPayDecodeAddress(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		addr, err := pandora_pay_addresses.DecodeAddr(args[0].String())
		if err != nil {
			return nil, err
		}

		if addr.Network != config.NETWORK_SELECTED {
			return nil, errors.New("address invalid network")
		}

		return webassembly_utils.ConvertJSONBytes(addr)
	})
}

func pandoraPayCreateAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		parameters := struct {
			PublicKey     []byte `json:"publicKey"`
			Registration  []byte `json:"registration"`
			PaymentID     []byte `json:"paymentID"`
			PaymentAmount uint64 `json:"paymentAmount"`
			PaymentAsset  []byte `json:"paymentAsset"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], &parameters); err != nil {
			return nil, err
		}

		addr, err := pandora_pay_addresses.CreateAddr(parameters.PublicKey, false, nil, parameters.Registration, parameters.PaymentID, parameters.PaymentAmount, parameters.PaymentAsset)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes([]interface{}{
			addr,
			addr.EncodeAddr(),
		})

	})
}
