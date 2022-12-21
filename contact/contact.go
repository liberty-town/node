package contact

import (
	"errors"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/network/request"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"strings"
)

type Contact struct {
	ProtocolVersion uint64               `json:"protocolVersion"`
	Version         string               `json:"version"`
	Addresses       []*ContactAddress    `json:"addresses"`
	Ownership       *ownership.Ownership `json:"ownership"`
}

func (this *Contact) Validate() error {
	for i := range this.Addresses {
		if err := this.Addresses[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (this *Contact) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeOwnershipSignature bool) {
	w.WriteUvarint(this.ProtocolVersion)
	w.WriteString(this.Version)
	w.WriteByte(byte(len(this.Addresses)))
	for _, address := range this.Addresses {
		address.Serialize(w)
	}
	this.Ownership.AdvancedSerialize(w, includeOwnershipSignature)
}

func (this *Contact) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true)
}

func (this *Contact) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if this.ProtocolVersion, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.Version, err = r.ReadString(20); err != nil {
		return
	}

	var n byte
	if n, err = r.ReadByte(); err != nil {
		return
	}

	this.Addresses = make([]*ContactAddress, n)
	for i := range this.Addresses {
		this.Addresses[i] = &ContactAddress{}
		if err = this.Addresses[i].Deserialize(r); err != nil {
			return
		}
	}

	this.Ownership = &ownership.Ownership{}
	if err = this.Ownership.Deserialize(r, this.GetMessageForSigningOwnership); err != nil {
		return
	}

	return nil
}

func (this *Contact) GetMessageForSigningOwnership() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false)
	return w.Bytes()
}

func (this *Contact) GetAddress(addressType ContactAddressType) string {

	for _, v := range this.Addresses {

		if addressType == CONTACT_ADDRESS_TYPE_HTTP_SERVER && v.Type == CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER {
			addr := v.Address
			addr = strings.Replace(addr, "/ws", "", -1)
			addr = strings.Replace(addr, "ws://", "http://", -1)
			addr = strings.Replace(addr, "wss://", "https://", -1)
			return addr
		}

		if addressType == addressType {
			return v.Address
		}
	}

	return ""
}

func Send[T any](c *Contact, method string, data []byte) (*T, error) {

	if addr := c.GetAddress(CONTACT_ADDRESS_TYPE_HTTP_SERVER); addr != "" {

		received := new(T)
		if err := request.RequestPost(addr+"/"+method, data, received, "application/json"); err != nil {
			return nil, err
		}

		return received, nil

	}

	return nil, errors.New("contact can not be contacted")
}
