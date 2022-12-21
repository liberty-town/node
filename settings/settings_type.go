package settings

import (
	"liberty-town/node/addresses"
	"liberty-town/node/contact"
)

type settingsContact struct {
	PrivateKey *addresses.PrivateKey `json:"privateKey"`
	Address    *addresses.Address    `json:"-"`
	Contact    *contact.Contact      `json:"-"`
}

type settingsAccount struct {
	PrivateKey *addresses.PrivateKey `json:"privateKey"`
	Address    *addresses.Address    `json:"-"`
}

type SettingsType struct {
	Name       string                  `json:"name"`
	Mnemonic   string                  `json:"mnemonic"`
	Seed       *addresses.SeedExtended `json:"seed"`
	Contact    *settingsContact        `json:"contact"`
	Account    *settingsAccount        `json:"account"`
	Multisig   *settingsAccount        `json:"multisig"`
	Validation *settingsAccount        `json:"validator"`
}

type settingsJSONType struct {
	Name     string `json:"name"`
	Mnemonic string `json:"mnemonic"`
}
