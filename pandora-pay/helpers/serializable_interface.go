package helpers

import "liberty-town/node/pandora-pay/helpers/advanced_buffers"

type SerializableInterface interface {
	Serialize(w *advanced_buffers.BufferWriter)
	Deserialize(r *advanced_buffers.BufferReader) error
}

func SerializeToBytes(self SerializableInterface) []byte {
	w := advanced_buffers.NewBufferWriter()
	self.Serialize(w)
	return w.Bytes()
}
