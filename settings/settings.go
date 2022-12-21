package settings

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	bip32 "github.com/tyler-smith/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/config/arguments"
	"liberty-town/node/config/globals"
	"liberty-town/node/contact"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/gui"
	"liberty-town/node/network/network_config"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/pandora-pay/helpers/multicast"
)

var Settings = &generics.Value[*SettingsType]{}
var ChangedEvents = multicast.NewMulticastChannel[*SettingsType]()

// 读取配置
func Load() (err error) {

	s := &SettingsType{
		"Account",
		"",
		nil, nil, nil, nil, nil,
	}

	if err = s.loadSettings(); err != nil {
		if err.Error() != "settings don't exist" {
			return
		}
		if err = s.CreateEmptyMnemonic("", true); err != nil {
			return
		}
	}

	if err = s.Init(true); err != nil {
		return
	}

	if err = s.ProcessArguments(); err != nil {
		return err
	}

	return
}

func (this *SettingsType) Init(save bool) (err error) {

	changed := false

	b := helpers.SerializeToBytes(this.Contact.Contact)
	c2 := &contact.Contact{}

	if err = c2.Deserialize(advanced_buffers.NewBufferReader(b)); err != nil {
		return
	}

	if network_config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING != "" {

		url := network_config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING

		newIpDetected := -1

		for i, a := range c2.Addresses {
			if a.Type == contact.CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER && a.Address != url {
				newIpDetected = i
				break
			}
		}

		if newIpDetected >= 0 {
			c2.Addresses[newIpDetected] = &contact.ContactAddress{
				contact.CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER,
				url,
			}
			changed = true
		} else if c2.GetAddress(contact.CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER) == "" {
			c2.Addresses = append(c2.Addresses, &contact.ContactAddress{
				contact.CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER,
				url,
			})
			changed = true
		}

	}

	if len(c2.Ownership.Signature) == 0 || changed {
		if err = c2.Ownership.Sign(this.Contact.PrivateKey, c2.GetMessageForSigningOwnership); err != nil {
			return
		}
		if !c2.Ownership.Verify(c2.GetMessageForSigningOwnership) {
			return errors.New("verification returned an error")
		}

		this.Contact.Contact = c2
		if save {
			if err = this.saveSettings(); err != nil {
				return
			}
		}
	}

	Settings.Store(this)
	ChangedEvents.Broadcast(this)

	if arguments.Arguments["--display-identity"] != nil {
		gui.GUI.Info("contact identity", this.Contact.Contact.Ownership.Address.Encoded)
		fmt.Printf("contact dump %s \n\n", base64.StdEncoding.EncodeToString(helpers.SerializeToBytes(this.Contact.Contact)))

		b, err := json.Marshal(this.Contact.Contact)
		if err != nil {
			return err
		}
		gui.GUI.Info("contact json", string(b))
	}

	globals.MainEvents.BroadcastEvent("settings", "changed")

	return
}

func (this *SettingsType) CreateEmptyEntropy(entropy []byte, save bool) (err error) {
	var mnemonic string
	if mnemonic, err = bip39.NewMnemonic(entropy); err != nil {
		return
	}
	return this.CreateEmptyMnemonic(mnemonic, save)
}

func (this *SettingsType) CreateEmptyMnemonic(mnemonic string, save bool) (err error) {

	if mnemonic == "" {
		var entropy []byte
		if entropy, err = bip39.NewEntropy(256); err != nil {
			return
		}

		if mnemonic, err = bip39.NewMnemonic(entropy); err != nil {
			return
		}
	} else {
		if !bip39.IsMnemonicValid(mnemonic) {
			return errors.New("Invalid mnemonic")
		}
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "SEED Secret Passphrase")
	if err != nil {
		return
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return
	}

	//contact
	start := bip32.FirstHardenedChild
	key2, err := masterKey.NewChildKey(start + 0)
	if err != nil {
		return
	}

	this.Mnemonic = mnemonic
	if this.Seed, err = addresses.NewSeedExtended(seed); err != nil {
		return
	}

	this.Contact = &settingsContact{}
	if this.Contact.PrivateKey, err = addresses.NewPrivateKey(key2.Key); err != nil {
		return
	}
	if this.Contact.Address, err = this.Contact.PrivateKey.GenerateAddress(); err != nil {
		return
	}

	this.Contact.Contact = &contact.Contact{
		config.PROTOCOL_VERSION,
		config.VERSION_STRING,
		[]*contact.ContactAddress{},
		&ownership.Ownership{
			0,
			nil,
			this.Contact.Address,
		},
	}
	if err = this.Contact.Contact.Ownership.Sign(this.Contact.PrivateKey, this.Contact.Contact.GetMessageForSigningOwnership); err != nil {
		panic(err)
	}

	//account
	key3, err := masterKey.NewChildKey(start + 1)
	if err != nil {
		return err
	}

	this.Account = &settingsAccount{}

	if this.Account.PrivateKey, err = addresses.NewPrivateKey(key3.Key); err != nil {
		return
	}
	if this.Account.Address, err = this.Account.PrivateKey.GenerateAddress(); err != nil {
		return
	}

	//multisig
	key4, err := masterKey.NewChildKey(start + 2)
	if err != nil {
		return err
	}

	this.Multisig = &settingsAccount{}
	if this.Multisig.PrivateKey, err = addresses.NewPrivateKey(key4.Key); err != nil {
		return
	}
	if this.Multisig.Address, err = this.Multisig.PrivateKey.GenerateAddress(); err != nil {
		return
	}

	//validation
	key5, err := masterKey.NewChildKey(start + 3)
	if err != nil {
		return err
	}

	this.Validation = &settingsAccount{}
	if this.Validation.PrivateKey, err = addresses.NewPrivateKey(key5.Key); err != nil {
		return
	}
	if this.Validation.Address, err = this.Multisig.PrivateKey.GenerateAddress(); err != nil {
		return
	}

	if save {
		if err = this.saveSettings(); err != nil {
			return
		}
	}

	return
}

func (this *SettingsType) Clone() *SettingsType {
	return &SettingsType{
		this.Name,
		this.Mnemonic,
		this.Seed,
		this.Contact,
		this.Account,
		this.Multisig,
		this.Validation,
	}
}

func (this *SettingsType) ExportJSON() ([]byte, error) {
	return json.Marshal(&settingsJSONType{
		this.Name,
		this.Mnemonic,
	})
}

func (this *SettingsType) GetSecretEntropy() ([]byte, error) {
	return bip39.EntropyFromMnemonic(this.Mnemonic)
}
