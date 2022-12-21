package invoices

import (
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	pandora_pay_cryptography "liberty-town/node/pandora-pay/cryptography"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

//买家信息
type InvoiceBuyerAccount struct {
	Address          *addresses.Address `json:"address" msgpack:"address"`
	Nonce            []byte             `json:"nonce" msgpack:"nonce"`
	Multisig         []byte             `json:"multisig" msgpack:"multisig"`
	ConversionAsset  string             `json:"conversionAsset" msgpack:"conversionAsset"`
	ConversionAmount uint64             `json:"conversionAmount" msgpack:"conversionAmount"`
	Signature        []byte             `json:"signature" msgpack:"signature"`
}

func (this *InvoiceBuyerAccount) Serialize(w *advanced_buffers.BufferWriter) {
	this.Address.Serialize(w)
	w.WriteBool(len(this.Nonce) > 0)
	if len(this.Nonce) > 0 {
		w.Write(this.Nonce)
		w.Write(this.Multisig)
		w.WriteString(this.ConversionAsset)
		w.WriteUvarint(this.ConversionAmount)
		w.WriteBool(len(this.Signature) > 0)
		if len(this.Signature) > 0 {
			w.Write(this.Signature)
		}
	}
}

func (this *InvoiceBuyerAccount) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	this.Address = &addresses.Address{}
	if err = this.Address.Deserialize(r); err != nil {
		return
	}

	var b bool
	if b, err = r.ReadBool(); err != nil {
		return
	}
	if b {
		if this.Nonce, err = r.ReadBytes(cryptography.HashSize); err != nil {
			return
		}
		if this.Multisig, err = r.ReadBytes(pandora_pay_cryptography.PublicKeySize); err != nil {
			return
		}
		if this.ConversionAsset, err = r.ReadString(config.ACCEPTED_ASSET_LENGTH); err != nil {
			return
		}
		if this.ConversionAmount, err = r.ReadUvarint(); err != nil {
			return
		}

		if b, err = r.ReadBool(); err != nil {
			return
		}
		if b {
			if this.Signature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
				return
			}
		} else {
			this.Signature = nil
		}
	} else {
		this.Nonce = nil
		this.ConversionAsset = ""
		this.ConversionAmount = 0
		this.Multisig = nil
		this.Signature = nil
	}

	return
}
