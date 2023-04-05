package shipping

import (
	"errors"
	"liberty-town/node/config"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type Shipping struct {
	Option string `json:"option" msgpack:"option"`
	Price  uint64 `json:"price" msgpack:"option"`
}

func (this *Shipping) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteString(this.Option)
	w.WriteUvarint(this.Price)
}

func (this *Shipping) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if this.Option, err = r.ReadString(config.LISTING_SHIPPING_MAX_LENGTH); err != nil {
		return
	}
	if this.Price, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}

func (this *Shipping) Validate() error {
	if len(this.Option) < config.LISTING_SHIPPING_MIN_LENGTH || len(this.Option) > config.LISTING_SHIPPING_MAX_LENGTH {
		return errors.New("invalid shipping option")
	}
	return nil
}
