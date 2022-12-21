package contact

import (
	"errors"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"net/url"
)

type ContactAddress struct {
	Type    ContactAddressType `json:"type"`
	Address string             `json:"address"`
}

func (c *ContactAddress) Validate() error {

	switch c.Type {
	case CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER, CONTACT_ADDRESS_TYPE_HTTP_SERVER:
		if _, err := url.Parse(c.Address); err != nil {
			return err
		}
	default:
		return errors.New("invalid")
	}

	return nil
}

func (c *ContactAddress) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteByte(byte(c.Type))
	switch c.Type {
	case CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER, CONTACT_ADDRESS_TYPE_HTTP_SERVER:
		w.WriteString(c.Address)
	}
}

func (c *ContactAddress) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var b byte
	if b, err = r.ReadByte(); err != nil {
		return
	}

	c.Type = ContactAddressType(b)

	switch c.Type {
	case CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER, CONTACT_ADDRESS_TYPE_HTTP_SERVER:
		if c.Address, err = r.ReadString(100); err != nil {
			return
		}
	default:
		return errors.New("invalid Contact Address Type ")
	}

	return
}
