package invoices

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

//产品详情
type InvoiceItem struct {
	Version  InvoiceItemVersion `json:"version" msgpack:"version"`
	Address  *addresses.Address `json:"address" msgpack:"address"`
	Name     string             `json:"name" msgpack:"name"`
	Offer    string             `json:"offer" msgpack:"offer"`
	Price    uint64             `json:"price" msgpack:"price"`
	Quantity uint64             `json:"quantity" msgpack:"quantity"`
}

func (this *InvoiceItem) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteUvarint(uint64(this.Version))
	if this.Version == INVOICE_ITEM_ID {
		this.Address.Serialize(w)
	}
	w.WriteString(this.Name)
	w.WriteString(this.Offer)
	w.WriteUvarint(this.Price)
	w.WriteUvarint(this.Quantity)
}

func (this *InvoiceItem) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	this.Version = InvoiceItemVersion(n)

	switch this.Version {
	case INVOICE_ITEM_ID:
		this.Address = &addresses.Address{}
		if err = this.Address.Deserialize(r); err != nil {
			return
		}
	case INVOICE_ITEM_NEW:
		this.Address = nil
	default:
		errors.New("invalid invoice item")
	}

	if this.Name, err = r.ReadString(config.LISTING_TITLE_MAX_LENGTH); err != nil {
		return
	}
	if this.Offer, err = r.ReadString(config.LISTING_OFFER_MAX_LENGTH); err != nil {
		return
	}
	if this.Price, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Quantity, err = r.ReadUvarint(); err != nil {
		return
	}

	return nil
}

func (this *InvoiceItem) Validate() error {
	switch this.Version {
	case INVOICE_ITEM_ID:
		if this.Address == nil {
			return errors.New("it should have id")
		}
	case INVOICE_ITEM_NEW:
		if this.Address != nil {
			return errors.New("it should not have id")
		}
	default:
		return errors.New("invoice item version is invalid")
	}
	if len(this.Name) < config.LISTING_TITLE_MIN_LENGTH || len(this.Name) > config.LISTING_TITLE_MAX_LENGTH {
		return errors.New("invoice item name is invalid")
	}
	if len(this.Offer) < config.LISTING_OFFER_MIN_LENGTH || len(this.Offer) > config.LISTING_OFFER_MAX_LENGTH {
		return errors.New("invoice offer is invalid")
	}
	if this.Quantity == 0 {
		return errors.New("invoice item quantity must be greater than zero")
	}
	return nil
}
