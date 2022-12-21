package transaction

import (
	"errors"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_type"
	"liberty-town/node/pandora-pay/cryptography"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type Transaction struct {
	transaction_base_interface.TransactionBaseInterface
	Version    transaction_type.TransactionVersion
	SpaceExtra uint64
}

func (tx *Transaction) SerializeManualToBytes() []byte {
	writer := advanced_buffers.NewBufferWriter()
	tx.SerializeAdvanced(writer, true)
	return writer.Bytes()
}

func (tx *Transaction) HashManual() []byte {
	serialized := tx.SerializeManualToBytes()
	return cryptography.SHA3(serialized)
}

func (tx *Transaction) SerializeAdvanced(w *advanced_buffers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(uint64(tx.Version))
	w.WriteUvarint(tx.SpaceExtra)
	tx.TransactionBaseInterface.SerializeAdvanced(w, inclSignature)
}

func (tx *Transaction) Serialize(w *advanced_buffers.BufferWriter) {
	tx.SerializeAdvanced(w, true)
}

func (tx *Transaction) validate() error {
	if tx.Version >= transaction_type.TX_END {
		return errors.New("VersionType is invalid")
	}
	return tx.TransactionBaseInterface.Validate()
}
