package transaction_simple_extra

import "liberty-town/node/pandora-pay/helpers/advanced_buffers"

type TransactionSimpleExtraInterface interface {
	Serialize(w *advanced_buffers.BufferWriter, inclSignature bool)
	Deserialize(r *advanced_buffers.BufferReader) error
	Validate(fee uint64) error
}
