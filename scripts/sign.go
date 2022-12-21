package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/config/arguments"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations"
	"liberty-town/node/federations/federation"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/moderator"
	"liberty-town/node/gui"
	"liberty-town/node/network"
	"liberty-town/node/settings"
	"liberty-town/node/validator/validation"
	"os"
)

func readJson(obj any) error {

	file, err := ioutil.ReadFile("../../../scripts/input.txt")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(file, obj); err != nil {
		return err
	}

	return nil
}

func main() {

	var err error
	if err = arguments.InitArguments(os.Args[1:]); err != nil {
		panic(err)
	}
	if err = gui.InitGUI(); err != nil {
		panic(err)
	}
	if err = config.InitConfig(); err != nil {
		panic(err)
	}
	if err = federations.InitializeFederations(); err != nil {
		return
	}
	if err = network.NewNetwork(); err != nil {
		panic(err)
	}

	if err = settings.ImportMnemonic("", false); err != nil {
		return
	}

	reader := bufio.NewReaderSize(os.Stdin, 1024*1024)

	gui.GUI.Info("selection options: sign-fed, sign-moderator")

	answer, _ := reader.ReadString('\n')

	var obj any
	var owner *ownership.Ownership
	var valid *validation.Validation
	var getData func() []byte

	switch answer {
	case "sign-fed\n":
		fed := &federation.Federation{}
		if err = readJson(fed); err != nil {
			panic(err)
		}
		obj = fed
		fed.Ownership = &ownership.Ownership{}

		owner = fed.Ownership
		getData = fed.GetMessageToSign
	case "sign-moderator\n":
		mod := &moderator.Moderator{}
		if err := readJson(mod); err != nil {
			panic(err)
		}

		federations.FederationsDict.Range(func(key string, fed *federation.Federation) bool {

			if fed.FindModerator(mod.Ownership.Address.Encoded) != nil {
				mod.Validation = &validation.Validation{}
				if mod.Validation, err = fed.SignValidation(mod.GetMessageForSigningValidator, nil); err != nil {
					panic(err)
				}
			}

			return true
		})

		mod.Ownership = &ownership.Ownership{}

		obj = mod
		owner = mod.Ownership
		valid = mod.Validation
		getData = mod.GetMessageForSigningOwnership
	default:
		panic("Invalid command")
	}

	gui.GUI.Info("---------------------------------------")
	gui.GUI.Info("Write Ownership Private Key")
	answer, _ = reader.ReadString('\n')
	data, err := base64.StdEncoding.DecodeString(answer[0 : len(answer)-1])
	if err != nil {
		panic(err)
	}

	var privateKey *addresses.PrivateKey
	if len(data) == cryptography.PrivateKeySize {
		if privateKey, err = addresses.NewPrivateKey(data); err != nil {
			return
		}
	} else {
		privateKey = &addresses.PrivateKey{}
		if err = privateKey.Deserialize(data); err != nil {
			return
		}
	}

	if owner.Address, err = privateKey.GenerateAddress(); err != nil {
		panic(err)
	}

	if !bytes.Equal(privateKey.GeneratePublicKey(), owner.Address.PublicKey) {
		panic("Ownership Public key is not matching")
	}

	if err = owner.Sign(privateKey, getData); err != nil {
		panic(err)
	}

	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	gui.GUI.Log("---------------------------------------")
	gui.GUI.Log("---------------------------------------")
	gui.GUI.Log("json", string(b))
	gui.GUI.Log("---------------------------------------")

	if valid != nil {
		gui.GUI.Log("validation", base64.StdEncoding.EncodeToString(valid.SerializeToBytes()))
		gui.GUI.Log("---------------------------------------")
	}
	if owner != nil {
		gui.GUI.Log("ownership", base64.StdEncoding.EncodeToString(owner.SerializeToBytes()))
		gui.GUI.Log("---------------------------------------")
	}

}
