package small_sorted_set

import (
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type SmallSortedSetNode struct {
	Key   string
	Score float64
}

func (this *SmallSortedSetNode) Validate() error {
	return nil
}

func (this *SmallSortedSetNode) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteString(this.Key)
	w.WriteFloat64(this.Score)
}

func (this *SmallSortedSetNode) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if this.Key, err = r.ReadString(255); err != nil {
		return
	}
	if this.Score, err = r.ReadFloat64(); err != nil {
		return
	}
	return nil
}
