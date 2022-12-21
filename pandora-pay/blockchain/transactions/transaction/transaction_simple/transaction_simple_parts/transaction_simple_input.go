package transaction_simple_parts

import (
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type TransactionSimpleInput struct {
}

func (vin *TransactionSimpleInput) Validate() error {
	return nil
}

func (vin *TransactionSimpleInput) Serialize(w *advanced_buffers.BufferWriter, inclSignature bool) {
}

func (vin *TransactionSimpleInput) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	return
}
