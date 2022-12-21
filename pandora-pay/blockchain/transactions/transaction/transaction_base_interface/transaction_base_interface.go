package transaction_base_interface

import (
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type TransactionBaseInterface interface {
	Validate() error
	SerializeAdvanced(w *advanced_buffers.BufferWriter, inclSignature bool)
	VerifySignatureManually(hashForSignature []byte) bool
}
