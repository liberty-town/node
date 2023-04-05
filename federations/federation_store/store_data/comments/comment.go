package comments

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator/validation"
)

type Comment struct {
	Version            CommentVersion         `json:"version" msgpack:"version"`
	FederationIdentity *addresses.Address     `json:"federation"  msgpack:"federation"` //not serialized
	ParentIdentity     *addresses.Address     `json:"parent"  msgpack:"parent"`         //not serialized
	Content            string                 `json:"content"  msgpack:"content"`
	Validation         *validation.Validation `json:"validation"  msgpack:"validation"`
	Publisher          *ownership.Ownership   `json:"publisher"  msgpack:"publisher"`
	Identity           *addresses.Address     `json:"identity"  msgpack:"identity"`
}

func (this *Comment) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeUnserializedData, includeValidation, includePublisher, includePublisherSignature bool) {
	w.WriteUvarint(uint64(this.Version))

	if includeUnserializedData {
		this.FederationIdentity.Serialize(w)
		this.ParentIdentity.Serialize(w)
	}

	w.WriteString(this.Content)

	if includeValidation {
		this.Validation.Serialize(w)
	}

	if includeValidation && includePublisher {
		this.Publisher.AdvancedSerialize(w, includePublisherSignature)
	}

}

func (this *Comment) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, false, true, true, true)
}

func (this *Comment) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var version uint64

	if version, err = r.ReadUvarint(); err != nil {
		return err
	}

	this.Version = CommentVersion(version)

	switch this.Version {
	case COMMENT_VERSION:
	default:
		return errors.New("invalid comment version")
	}

	if this.Content, err = r.ReadString(config.COMMENT_CONTENT_MAX_LENGTH); err != nil {
		return
	}

	this.Validation = &validation.Validation{}
	if err = this.Validation.Deserialize(r, this.GetMessageForSigningValidator, nil); err != nil {
		return
	}

	this.Publisher = &ownership.Ownership{}
	if err = this.Publisher.Deserialize(r, this.GetMessageForSigningPublisher); err != nil {
		return
	}

	return this.SetIdentityNow()
}

func (this *Comment) SetIdentityNow() (err error) {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, false, true, true)
	this.Identity, err = addresses.CreateAddr(cryptography.SHA3(w.Bytes()))
	return err
}

func (this *Comment) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, false, false, false)
	return w.Bytes()
}

func (this *Comment) GetMessageForSigningPublisher() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, true, false)
	return w.Bytes()
}

func (this *Comment) IsDeletable() bool {
	return false
}

func (this *Comment) Validate() error {

	switch this.Version {
	case COMMENT_VERSION:
	default:
		return errors.New("invalid comment")
	}

	if len(this.Content) > config.COMMENT_CONTENT_MAX_LENGTH {
		return errors.New("comment content is too large")
	}
	if len(this.Content) < config.COMMENT_CONTENT_MIN_LENGTH {
		return errors.New("comment content is too small")
	}

	if this.FederationIdentity == nil {
		return errors.New("federation is not set")
	}
	if this.ParentIdentity == nil {
		return errors.New("parent identity is not set")
	}

	return nil
}

func (this *Comment) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator, nil) {
		return errors.New("listing validation signature failed")
	}
	if !this.Publisher.Verify(this.GetMessageForSigningPublisher) {
		return errors.New("listing publisher signature failed")
	}
	return nil
}
