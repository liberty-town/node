package category

import (
	"fmt"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type Category struct {
	Id  uint64      `json:"id"`
	N   string      `json:"n"`
	Sub []*Category `json:"sub,omitempty"`
}

func (this *Category) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteUvarint(this.Id)
	w.WriteString(this.N)
	w.WriteByte(byte(len(this.Sub)))
	for i := range this.Sub {
		this.Sub[i].Serialize(w)
	}
}

func (this *Category) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if this.Id, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.N, err = r.ReadString(100); err != nil {
		return
	}
	var n byte
	if n, err = r.ReadByte(); err != nil {
		return
	}
	this.Sub = make([]*Category, n)
	for i := range this.Sub {
		this.Sub[i] = &Category{}
		if err = this.Sub[i].Deserialize(r); err != nil {
			return
		}
	}
	return nil
}

func (this *Category) Process(uniqueCategories map[uint64]bool) error {
	if uniqueCategories[this.Id] {
		return fmt.Errorf("duplicate %d id found", this.Id)
	}

	uniqueCategories[this.Id] = true
	for i := range this.Sub {
		if err := this.Sub[i].Process(uniqueCategories); err != nil {
			return err
		}
	}
	return nil
}
