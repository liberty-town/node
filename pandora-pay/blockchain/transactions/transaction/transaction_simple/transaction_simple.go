package transaction_simple

import (
	"errors"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_data"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type TransactionSimple struct {
	Extra       transaction_simple_extra.TransactionSimpleExtraInterface
	TxScript    ScriptType
	DataVersion transaction_data.TransactionDataVersion
	Data        []byte
	Nonce       uint64
	Fee         uint64
	Vin         *transaction_simple_parts.TransactionSimpleInput
}

func (tx *TransactionSimple) VerifySignatureManually(hashForSignature []byte) bool {

	if tx.TxScript == SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT {
		extra := tx.Extra.(*transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment)
		if !extra.VerifySignature() {
			return false
		}
	}

	return true
}

func (tx *TransactionSimple) Validate() (err error) {

	if tx.HasVin() {
		if err = tx.Vin.Validate(); err != nil {
			return
		}
	}

	switch tx.TxScript {
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY, SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT:
		if tx.Extra == nil {
			return errors.New("extra is not assigned")
		}
		if err = tx.Extra.Validate(tx.Fee); err != nil {
			return
		}
	default:
		return errors.New("Invalid Simple TxScript")
	}

	return
}

func (tx *TransactionSimple) SerializeAdvanced(w *advanced_buffers.BufferWriter, inclSignature bool) {

	w.WriteUvarint(uint64(tx.TxScript))

	w.WriteByte(byte(tx.DataVersion))
	if tx.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT || tx.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
		w.WriteVariableBytes(tx.Data)
	}

	if tx.HasVin() {
		w.WriteUvarint(tx.Nonce)
		w.WriteUvarint(tx.Fee)
		tx.Vin.Serialize(w, inclSignature)
	}

	if tx.Extra != nil {
		tx.Extra.Serialize(w, inclSignature)
	}
}

func (tx *TransactionSimple) Serialize(w *advanced_buffers.BufferWriter) {
	tx.SerializeAdvanced(w, true)
}

func (tx *TransactionSimple) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	return
}

func (tx *TransactionSimple) HasVin() bool {
	switch tx.TxScript {
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		return true
	default:
		return false
	}
}
