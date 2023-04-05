package chat_message

import (
	"encoding/base64"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

type ChatMessage struct {
	Version            ChatMessageVersion     `json:"version"`
	FederationIdentity *addresses.Address     `json:"federationIdentity"`
	First              *addresses.Address     `json:"first"`
	Second             *addresses.Address     `json:"second"`
	Nonce              []byte                 `json:"nonce"`
	FirstMessage       []byte                 `json:"firstMessage"`
	SecondMessage      []byte                 `json:"secondMessage"`
	Validation         *validation.Validation `json:"validation"`
}

func (this *ChatMessage) GetUniqueId() string {
	w := advanced_buffers.NewBufferWriter()
	this.Serialize(w)
	return base64.StdEncoding.EncodeToString(cryptography.SHA3(w.Bytes()))
}

func (this *ChatMessage) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeValidation bool) {

	w.WriteUvarint(uint64(this.Version))
	this.FederationIdentity.Serialize(w)
	this.First.Serialize(w)
	this.Second.Serialize(w)
	w.Write(this.Nonce)
	w.WriteVariableBytes(this.FirstMessage)
	w.WriteVariableBytes(this.SecondMessage)

	if includeValidation {
		this.Validation.Serialize(w)
	}

}

func (this *ChatMessage) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true)
}

func (this *ChatMessage) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var version uint64

	if version, err = r.ReadUvarint(); err != nil {
		return err
	}

	this.Version = ChatMessageVersion(version)

	switch this.Version {
	case CHAT_MESSAGE:
	default:
		return errors.New("invalid listing version")
	}

	this.FederationIdentity = &addresses.Address{}
	if err = this.FederationIdentity.Deserialize(r); err != nil {
		return
	}

	this.First = &addresses.Address{}
	if err = this.First.Deserialize(r); err != nil {
		return err
	}

	this.Second = &addresses.Address{}
	if err = this.Second.Deserialize(r); err != nil {
		return err
	}

	if this.Nonce, err = r.ReadBytes(cryptography.HashSize); err != nil {
		return err
	}

	if this.FirstMessage, err = r.ReadVariableBytes(config.CHAT_MESSAGE_MAX_LENGTH); err != nil {
		return err
	}
	if this.SecondMessage, err = r.ReadVariableBytes(config.CHAT_MESSAGE_MAX_LENGTH); err != nil {
		return err
	}

	this.Validation = &validation.Validation{}
	if err = this.Validation.Deserialize(r, this.GetMessageForSigningValidator, nil); err != nil {
		return
	}

	return
}

func (this *ChatMessage) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false)
	return w.Bytes()
}

func (this *ChatMessage) IsBetter(old *ChatMessage) bool {
	return old == nil || old.GetBetterScore() < this.GetBetterScore()
}

func (this *ChatMessage) GetBetterScore() uint64 {
	return this.Validation.Timestamp
}

func (this *ChatMessage) IsDeletable() bool {
	return false
}

func (this *ChatMessage) Validate() error {

	_, _, ok := SortKeys(this.First.Encoded, this.Second.Encoded)
	if ok {
		return errors.New("keys are not sorted")
	}

	switch this.Version {
	case CHAT_MESSAGE:
	default:
		return errors.New("invalid message")
	}

	if this.First.Network != config.NETWORK_SELECTED || this.Second.Network != config.NETWORK_SELECTED {
		return errors.New("invalid network")
	}

	if len(this.FirstMessage) > config.CHAT_MESSAGE_MAX_LENGTH {
		return errors.New("invalid first message length")
	}
	if len(this.SecondMessage) > config.CHAT_MESSAGE_MAX_LENGTH {
		return errors.New("invalid second message length")
	}

	return nil
}

func (this *ChatMessage) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator, nil) {
		return errors.New("ACCOUNT VALIDATION FAILED")
	}
	return nil
}
