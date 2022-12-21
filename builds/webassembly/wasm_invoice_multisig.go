package main

import (
	"errors"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	pandora_pay_addresses "liberty-town/node/pandora-pay/addresses"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_type"
	pandora_pay_cryptography "liberty-town/node/pandora-pay/cryptography"
	"liberty-town/node/settings"
	"syscall/js"
)

func getMultisigPair(nonce []byte) (*pandora_pay_addresses.PrivateKey, error) {
	r := pandora_pay_cryptography.SHA3(pandora_pay_cryptography.SHA3(settings.Settings.Load().Multisig.PrivateKey.Key[:]))
	finalKey := pandora_pay_cryptography.SHA3(pandora_pay_cryptography.SHA3(append(nonce, r...)))
	finalPrivateKey, err := pandora_pay_addresses.NewPrivateKey(finalKey)
	if err != nil {
		return nil, err
	}

	return finalPrivateKey, nil
}

func invoiceMultisigCompute(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		req := &struct {
			Nonce []byte `json:"nonce"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		if len(req.Nonce) != 32 {
			return nil, errors.New("compute multisig nonce is invalid")
		}

		key, err := getMultisigPair(req.Nonce)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(key.GeneratePublicKey()), nil
	})
}

func invoiceMultisigSign(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		req := &struct {
			Nonce        []byte `json:"nonce"`
			TxId         []byte `json:"txId"`
			PayloadIndex byte   `json:"payloadIndex"`
			Resolution   bool   `json:"resolution"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		if len(req.Nonce) != 32 {
			return nil, errors.New("multisig sign nonce is invalid")
		}
		if len(req.TxId) != pandora_pay_cryptography.HashSize {
			return nil, errors.New("txId is invalid")
		}

		key, err := getMultisigPair(req.Nonce)
		if err != nil {
			return nil, err
		}

		extra := &transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment{nil,
			req.TxId,
			req.PayloadIndex,
			req.Resolution,
			nil, nil,
		}

		signature, err := key.Sign(extra.MessageForSigning())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(signature), nil
	})
}

func invoiceModeratorMultisigSign(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		req := &struct {
			PrivateKey   []byte `json:"privateKey"`
			TxId         []byte `json:"txId"`
			PayloadIndex byte   `json:"payloadIndex"`
			Resolution   bool   `json:"resolution"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		if len(req.PrivateKey) != pandora_pay_cryptography.PrivateKeySize {
			return nil, errors.New("private key is invalid")
		}
		if len(req.TxId) != pandora_pay_cryptography.HashSize {
			return nil, errors.New("txId is invalid")
		}

		key, err := pandora_pay_addresses.NewPrivateKey(req.PrivateKey)
		if err != nil {
			return nil, err
		}

		extra := &transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment{nil,
			req.TxId,
			req.PayloadIndex,
			req.Resolution,
			nil, nil,
		}

		signature, err := key.Sign(extra.MessageForSigning())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(signature), nil
	})
}

func invoiceMultisigVerify(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {
		req := &struct {
			TxId              []byte `json:"txId"`
			PayloadIndex      byte   `json:"payloadIndex"`
			Resolution        bool   `json:"resolution"`
			MultisigPublicKey []byte `json:"multisigPublicKey"`
			MultisigSignature []byte `json:"multisigSignature"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		if len(req.TxId) != 32 {
			return nil, errors.New("txId is invalid")
		}
		if len(req.MultisigPublicKey) != pandora_pay_cryptography.PublicKeySize {
			return nil, errors.New("invalid public key")
		}
		if len(req.MultisigSignature) != pandora_pay_cryptography.SignatureSize {
			return nil, errors.New("invalid signature")
		}

		extra := &transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment{nil,
			req.TxId,
			req.PayloadIndex,
			req.Resolution,
			[][]byte{req.MultisigPublicKey}, [][]byte{req.MultisigSignature},
		}

		return extra.VerifySignature(), nil
	})
}

func invoiceMultisigClaimTx(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		req := &struct {
			TxId               []byte   `json:"txId"`
			PayloadIndex       byte     `json:"payloadIndex"`
			Resolution         bool     `json:"resolution"`
			MultisigPublicKeys [][]byte `json:"multisigPublicKeys"`
			MultisigSignatures [][]byte `json:"multisigSignatures"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		if len(req.TxId) != 32 {
			return nil, errors.New("txId is invalid")
		}

		for i := range req.MultisigPublicKeys {
			if len(req.MultisigPublicKeys[i]) != pandora_pay_cryptography.PublicKeySize {
				return nil, errors.New("invalid public key")
			}
			if len(req.MultisigSignatures[i]) != pandora_pay_cryptography.SignatureSize {
				return nil, errors.New("invalid signature")
			}
		}

		extra := &transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment{nil,
			req.TxId,
			req.PayloadIndex,
			req.Resolution,
			req.MultisigPublicKeys, req.MultisigSignatures,
		}

		if extra.VerifySignature() == false {
			return nil, errors.New("signature is wrong")
		}

		tx := &transaction.Transaction{
			&transaction_simple.TransactionSimple{
				extra,
				transaction_simple.SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT,
				0,
				nil,
				0,
				0,
				nil,
			},
			transaction_type.TX_SIMPLE,
			0,
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Serialized []byte `json:"serialized"`
			Hash       []byte `json:"hash"`
		}{
			tx.SerializeManualToBytes(),
			tx.HashManual(),
		})
	})
}
