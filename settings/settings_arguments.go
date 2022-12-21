package settings

import (
	"encoding/base64"
	"github.com/tyler-smith/go-bip39"
	"liberty-town/node/config/arguments"
)

func (this *SettingsType) ProcessArguments() (err error) {

	if mnemonic := arguments.Arguments["--settings-import-secret-mnemonic"]; mnemonic != nil {
		if this.Mnemonic != mnemonic.(string) {
			if err = ImportMnemonic(mnemonic.(string), true); err != nil {
				return
			}
		}
	}

	if entropy := arguments.Arguments["--settings-import-secret-entropy"]; entropy != nil {
		var b []byte
		var mnemonic string

		if b, err = base64.StdEncoding.DecodeString(entropy.(string)); err != nil {
			return
		}
		if mnemonic, err = bip39.NewMnemonic(b); err != nil {
			return
		}

		if mnemonic != this.Mnemonic {
			if err = ImportEntropy(b, true); err != nil {
				return
			}
		}
	}

	return
}
