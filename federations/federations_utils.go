package federations

import (
	"encoding/base64"
	"liberty-town/node/contact"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

func ContactDeserializeForced(data string) *contact.Contact {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		panic(err)
	}

	ct := new(contact.Contact)
	if err := ct.Deserialize(advanced_buffers.NewBufferReader(b)); err != nil {
		panic(err)
	}
	return ct
}

func OwnershipDeserializedForced(o *ownership.Ownership, data string, cb func() []byte) {
	if len(data) == 0 {
		return
	}

	r := advanced_buffers.NewBufferReader(helpers.DecodeBase64(data))
	if err := o.Deserialize(r, cb); err != nil {
		panic(err)
	}
}

func ValidationDeserialized(v *validation.Validation, data string, cb func() []byte) error {
	if len(data) == 0 {
		return nil
	}

	r := advanced_buffers.NewBufferReader(helpers.DecodeBase64(data))
	return v.Deserialize(r, cb, nil)
}
