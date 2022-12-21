//go:build !wasm
// +build !wasm

package validator

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/browser"
	"liberty-town/node/contact"
	"liberty-town/node/gui"
)

func (this *Validator) processValidate(validate func([]byte) []byte, challengeUri string, proof *validatorProof, data []byte) ([]byte, error) {

	proofSerialized, err := json.Marshal(proof)
	if err != nil {
		return nil, err
	}

	addr := this.Contact.GetAddress(contact.CONTACT_ADDRESS_TYPE_HTTP_SERVER) + challengeUri + "?showResult=true&proof=" + string(proofSerialized)
	if err := browser.OpenURL(addr); err != nil {
		return nil, err
	}

	gui.GUI.Info("Paste the result")

	var s string
	fmt.Scan(&s)

	return []byte(s), nil
}
