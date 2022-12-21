package main

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/config"
	"syscall/js"
)

func decodeAddress(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		addr, err := addresses.DecodeAddr(args[0].String())
		if err != nil {
			return nil, err
		}

		if addr.Network != config.NETWORK_SELECTED {
			return nil, errors.New("address invalid network")
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Network   uint64                   `json:"network" msgpack:"network"`
			Version   addresses.AddressVersion `json:"version" msgpack:"version"`
			PublicKey []byte                   `json:"publicKey" msgpack:"publicKey"`
			Encoded   string                   `json:"encoded" msgpack:"encoded"`
		}{
			addr.Network,
			addr.Version,
			addr.PublicKey,
			addr.Encoded,
		})
	})
}
