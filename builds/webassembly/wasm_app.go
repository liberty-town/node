package main

import (
	"errors"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/contact"
	"liberty-town/node/federations"
	"liberty-town/node/federations/federation"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator"
	"syscall/js"
)

func appFederationReplaceValidatorContactAddresses(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		validatorAddress := args[1].String()
		url := args[2].String()

		fed, _ := federations.FederationsDict.Load(args[0].String())
		if fed == nil {
			return nil, errors.New("federation not found")
		}

		w := advanced_buffers.NewBufferWriter()
		fed.Serialize(w)

		fed2 := &federation.Federation{}
		if err := fed2.Deserialize(advanced_buffers.NewBufferReader(w.Bytes())); err != nil {
			return nil, err
		}

		for _, v := range fed2.Validators {
			if v.Ownership.Address.Encoded == validatorAddress {
				v.Contact.Addresses = []*contact.ContactAddress{{
					contact.CONTACT_ADDRESS_TYPE_HTTP_SERVER,
					url,
				}}
				fed2.Validators = []*validator.Validator{v}
				federations.FederationsDict.Store(args[0].String(), fed2)
				federation_serve.SetServeFederation(fed2, false)
				return true, nil
			}
		}

		return false, nil
	})
}

func appGetFederations(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		dict := make(map[string]*federation.Federation)
		federations.FederationsDict.Range(func(key string, fed *federation.Federation) bool {
			dict[key] = fed
			return true
		})
		return webassembly_utils.ConvertJSONBytes(dict)
	})
}

func appSetSelectedFederation(this js.Value, args []js.Value) any {

	fed, _ := federations.FederationsDict.Load(args[0].String())
	if fed == nil {
		return errors.New("federation not found")
	}

	federation_serve.SetServeFederation(fed, false)

	return true
}
