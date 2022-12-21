package main

import (
	"encoding/base64"
	"encoding/json"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/settings"
	"syscall/js"
)

func settingsGet(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		settings := settings.Settings.Load()

		type outputAddressType struct {
			Address   string `json:"address"`
			PublicKey []byte `json:"publicKey"`
		}

		return webassembly_utils.ConvertJSONBytes(&struct {
			Name    string             `json:"name"`
			Contact *outputAddressType `json:"contact"`
			Account *outputAddressType `json:"account"`
		}{
			settings.Name,
			&outputAddressType{settings.Contact.Address.Encoded, settings.Contact.Address.PublicKey},
			&outputAddressType{settings.Account.Address.Encoded, settings.Account.Address.PublicKey},
		})
	})
}

func settingsGetSecretWords(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		return settings.Settings.Load().Mnemonic, nil
	})
}

func settingsGetSecretEntropy(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		entropy, err := settings.Settings.Load().GetSecretEntropy()
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertBytes(entropy), nil
	})
}

func settingsImportSecretWords(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		return true, settings.ImportMnemonic(args[0].String(), true)
	})
}

func settingsImportSecretEntropy(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		entropy, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}
		return true, settings.ImportEntropy(entropy, true)
	})
}

func settingsClear(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		return true, settings.ImportMnemonic("", true)
	})
}

func settingsExportJSON(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		out, err := settings.Settings.Load().ExportJSON()
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertBytes(out), nil
	})
}

func settingsExportJSONAll(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		s := settings.Settings.Load()
		out, err := json.Marshal(s)
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertBytes(out), nil
	})
}

func settingsRename(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		return true, settings.RenameSettings(args[0].String())
	})
}

func settingsImportJSON(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		return true, settings.ImportSettingsJSON(webassembly_utils.GetBytes(args[0]), true)
	})
}
