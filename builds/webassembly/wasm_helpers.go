package main

import (
	"encoding/base64"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/gui"
	"liberty-town/node/pandora-pay/helpers/identicon"
	"liberty-town/node/start"
	"syscall/js"
)

func test(this js.Value, args []js.Value) any {
	gui.GUI.Info("WASM is working")
	return true
}

func getIdenticon(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		publicKey, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		identicon, err := identicon.GenerateToBytes(publicKey, args[1].Int(), args[2].Int())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(identicon), nil
	})
}

func startLibrary(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		err := start.StartMainNow()
		return true, err
	})
}
