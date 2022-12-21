package main

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/cryptography"
	"liberty-town/node/settings"
	"syscall/js"
)

func cryptoRandomBytes(this js.Value, args []js.Value) any {
	b := cryptography.RandomBytes(args[0].Int())
	return webassembly_utils.ConvertBytes(b)
}

func sign(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		message := webassembly_utils.GetBytes(args[0])

		out, err := settings.Settings.Load().Account.PrivateKey.Sign(message)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(out), nil
	})
}

func verify(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		message := webassembly_utils.GetBytes(args[0])

		signature := webassembly_utils.GetBytes(args[1])

		address, err := addresses.DecodeAddr(args[2].String())
		if err != nil {
			return nil, errors.New("invalid address public key")
		}

		return address.VerifySignedMessage(message, signature), nil
	})
}
