//go:build !wasm
// +build !wasm

package federations

import (
	"liberty-town/node/config/arguments"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/gui"
)

func readArgument() (err error) {

	if arguments.Arguments["--serve-federation"] != nil {

		fedKey := arguments.Arguments["--serve-federation"].(string)

		if len(fedKey) > 0 {
			fed, _ := FederationsDict.Load(fedKey)
			if fed != nil {
				if err = federation_serve.SetServeFederation(fed, false); err != nil {
					return
				}
				gui.GUI.Warning("Federation Serving", fedKey)
				return
			}

			gui.GUI.Warning("Federation for serving was not found", fedKey)
		}

	}

	return
}
