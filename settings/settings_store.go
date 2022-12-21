package settings

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/contact"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/store"
	"liberty-town/node/store/store_db/store_db_interface"
)

func (this *SettingsType) saveSettings() error {

	return store.StoreSettings.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		writer.Put("exists", []byte{1})

		writer.Put("name", []byte(this.Name))
		writer.Put("mnemonic", []byte(this.Mnemonic))
		writer.Put("seed", this.Seed.Serialize())

		writer.Put("contact", helpers.SerializeToBytes(this.Contact.Contact))

		return
	})

}

func (this *SettingsType) loadSettings() error {

	return store.StoreSettings.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(reader.Get("exists")) == 0 {
			return errors.New("settings don't exist")
		}

		this.Name = string(reader.Get("name"))
		this.Mnemonic = string(reader.Get("mnemonic"))

		this.Seed = &addresses.SeedExtended{}
		if err = this.Seed.Deserialize(reader.Get("seed")); err != nil {
			return
		}

		this.Contact = &settingsContact{
			&addresses.PrivateKey{}, nil, nil,
		}

		if err = this.CreateEmptyMnemonic(this.Mnemonic, false); err != nil {
			return
		}

		this.Contact.Contact = &contact.Contact{}
		if err = this.Contact.Contact.Deserialize(advanced_buffers.NewBufferReader(reader.Get("contact"))); err != nil {
			return
		}

		return
	})
}
