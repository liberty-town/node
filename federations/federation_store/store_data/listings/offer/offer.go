package offer

import (
	"errors"
	"fmt"
	"liberty-town/node/config"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type Offer struct {
	Amount string `json:"amount" msgpack:"amount"`
	Price  uint64 `json:"price" msgpack:"price"`
}

func (this *Offer) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteString(this.Amount)
	w.WriteUvarint(this.Price)
}

func (this *Offer) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if this.Amount, err = r.ReadString(config.LISTING_OFFER_MAX_LENGTH); err != nil {
		return
	}
	if this.Price, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}

func (this *Offer) Validate() error {
	if this.Price == 0 {
		return errors.New("offer price should not be zero")
	}
	if len(this.Amount) < config.LISTING_OFFER_MIN_LENGTH {
		return errors.New("offer amount should not be empty")
	}
	if len(this.Amount) > config.LISTING_OFFER_MAX_LENGTH {
		return fmt.Errorf("offer amount should not exceed the limit %d", config.LISTING_OFFER_MAX_LENGTH)
	}
	return nil
}
