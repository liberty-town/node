package settings

import (
	"encoding/json"
)

func RenameSettings(newName string) (err error) {
	settings := Settings.Load().Clone()
	settings.Name = newName
	settings.Contact.Contact.Ownership.Signature = nil
	if err = settings.Init(true); err != nil {
		return
	}
	return
}

func ImportSettingsJSON(data []byte, save bool) (err error) {

	in := &settingsJSONType{}
	if err = json.Unmarshal(data, in); err != nil {
		return
	}

	settings := &SettingsType{
		Name: in.Name,
	}
	if err = settings.CreateEmptyMnemonic(in.Mnemonic, false); err != nil {
		return
	}
	if err = settings.Init(false); err != nil {
		return
	}
	if save {
		if err = settings.saveSettings(); err != nil {
			return
		}
	}
	return
}

func ImportEntropy(entropy []byte, save bool) (err error) {
	settings := &SettingsType{}
	if err = settings.CreateEmptyEntropy(entropy, false); err != nil {
		return
	}
	if err = settings.Init(false); err != nil {
		return
	}
	if save {
		if err = settings.saveSettings(); err != nil {
			return
		}
	}
	return
}

func ImportMnemonic(mnemonic string, save bool) (err error) {

	settings := &SettingsType{}
	if err = settings.CreateEmptyMnemonic(mnemonic, false); err != nil {
		return
	}
	if err = settings.Init(false); err != nil {
		return
	}
	if save {
		if err = settings.saveSettings(); err != nil {
			return
		}
	}
	return
}
