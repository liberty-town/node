package main

import (
	"encoding/json"
	"errors"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/cryptography"
	"liberty-town/node/invoices"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/settings"
	"syscall/js"
)

func invoiceCreateId(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &invoices.InvoiceId{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		if err := req.Validate(); err != nil {
			return nil, err
		}

		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(cryptography.SHA3(cryptography.SHA3(b))), nil
	})
}

func invoiceSign(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			Invoice *invoices.Invoice `json:"invoice"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		signature, err := settings.Settings.Load().Account.PrivateKey.Sign(req.Invoice.MessageForSignature())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(signature), nil
	})
}

func invoiceSerialize(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			Invoice *invoices.Invoice `json:"invoice"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		serialized := helpers.SerializeToBytes(req.Invoice)
		return webassembly_utils.ConvertBytes(serialized), nil
	})
}

func invoiceDeserialize(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		bytes := webassembly_utils.GetBytes(args[0])

		invoice := &invoices.Invoice{}
		if err := invoice.Deserialize(advanced_buffers.NewBufferReader(bytes)); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(invoice)
	})
}

func invoiceMessageToSignItems(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			Id       []byte                  `json:"id"`
			Items    []*invoices.InvoiceItem `json:"items"`
			Notes    string                  `json:"notes"`
			Shipping uint64                  `json:"shipping"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		if len(req.Id) != cryptography.HashSize {
			return nil, errors.New("invalid id")
		}

		for i := range req.Items {
			if err := req.Items[i].Validate(); err != nil {
				return nil, err
			}
		}

		if len(req.Notes) > 512 {
			return nil, errors.New("invalid notes")
		}

		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(b), nil
	})
}

func invoiceValidate(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			Invoice                 *invoices.Invoice `json:"invoice"`
			ValidateBuyer           bool              `json:"validateBuyer"`
			ValidateSeller          bool              `json:"validateSeller"`
			ValidateBuyerSignature  bool              `json:"validateBuyerSignature"`
			ValidateSellerSignature bool              `json:"validateSellerSignature"`
		}{}
		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}
		if err := req.Invoice.Validate(); err != nil {
			return nil, err
		}

		if req.ValidateBuyer {
			if err := req.Invoice.ValidateBuyer(req.ValidateBuyerSignature); err != nil {
				return nil, err
			}
		} else if req.Invoice.Buyer.Multisig != nil || req.Invoice.Buyer.Nonce != nil || req.Invoice.Buyer.Signature != nil || req.Invoice.Buyer.ConversionAsset != "" {
			return nil, errors.New("invoice buyer has invalid data")
		}

		if !req.ValidateBuyerSignature && req.Invoice.Buyer.Signature != nil {
			return nil, errors.New("invoice buyer signature has invalid data")
		}

		if req.ValidateSeller {
			if err := req.Invoice.ValidateSeller(req.ValidateSellerSignature); err != nil {
				return nil, err
			}
		} else if req.Invoice.Seller.Multisig != nil || req.Invoice.Seller.Recipient != "" || req.Invoice.Seller.Nonce != nil || req.Invoice.Seller.Signature != nil {
			return nil, errors.New("invoice seller has invalid data")
		}

		if !req.ValidateSellerSignature && req.Invoice.Seller.Signature != nil {
			return nil, errors.New("invoice buyer signature has invalid data")
		}

		return true, nil
	})
}

func invoiceValidateConfirmed(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			Invoice *invoices.Invoice `json:"invoice"`
		}{}
		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}
		if err := req.Invoice.Validate(); err != nil {
			return nil, err
		}
		if err := req.Invoice.ValidateBuyer(true); err != nil {
			return nil, err
		}
		if err := req.Invoice.ValidateSeller(true); err != nil {
			return nil, err
		}

		return true, nil
	})
}
