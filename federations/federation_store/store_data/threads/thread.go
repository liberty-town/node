package threads

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/threads/thread_type"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/advanced_strings"
	"liberty-town/node/validator/validation"
	"math"
	"net/url"
	"strings"
)

type Thread struct {
	Version            ThreadVersion          `json:"version" msgpack:"version"`
	FederationIdentity *addresses.Address     `json:"federation" msgpack:"federation"` //not serialized
	Type               thread_type.ThreadType `json:"type" msgpack:"type"`
	Title              string                 `json:"title" msgpack:"title"`
	Keywords           []string               `json:"keywords" msgpack:"keywords"`
	Content            string                 `json:"content"  msgpack:"content"`
	Links              []string               `json:"links"  msgpack:"links"`
	Validation         *validation.Validation `json:"validation"  msgpack:"validation"`
	Publisher          *ownership.Ownership   `json:"publisher"  msgpack:"publisher"`
	Identity           *addresses.Address     `json:"identity"  msgpack:"identity"`
}

func (this *Thread) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeUnserializedData, includeValidation, includePublisher, includePublisherSignature bool) {
	w.WriteUvarint(uint64(this.Version))

	if includeUnserializedData {
		this.FederationIdentity.Serialize(w)
	}

	w.WriteByte(byte(this.Type))

	w.WriteString(this.Title)

	w.WriteByte(byte(len(this.Keywords)))
	for i := range this.Keywords {
		w.WriteString(this.Keywords[i])
	}

	switch this.Type {
	case thread_type.THREAD_TEXT:
		w.WriteString(this.Content)
	case thread_type.THREAD_LINK, thread_type.THREAD_IMAGE:
		w.WriteByte(byte(len(this.Links)))
		for i := range this.Links {
			w.WriteString(this.Links[i])
		}
	}

	if includeValidation {
		this.Validation.Serialize(w)
	}

	if includeValidation && includePublisher {
		this.Publisher.AdvancedSerialize(w, includePublisherSignature)
	}

}

func (this *Thread) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, false, true, true, true)
}

func (this *Thread) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var version uint64
	var b byte

	if version, err = r.ReadUvarint(); err != nil {
		return err
	}

	this.Version = ThreadVersion(version)

	switch this.Version {
	case THREAD_VERSION:
	default:
		return errors.New("invalid thread version")
	}

	if b, err = r.ReadByte(); err != nil {
		return
	}
	this.Type = thread_type.ThreadType(b)

	if this.Title, err = r.ReadString(config.THREAD_TITLE_MAX_LENGTH); err != nil {
		return
	}

	if b, err = r.ReadByte(); err != nil {
		return
	}
	if b == 0 || b > config.THREAD_KEYWORDS_MAX_COUNT {
		return errors.New("invalid number of keywords")
	}
	this.Keywords = make([]string, b)
	for i := range this.Keywords {
		if this.Keywords[i], err = r.ReadString(config.THREAD_KEYWORD_MAX_LENGTH); err != nil {
			return
		}
	}

	switch this.Type {
	case thread_type.THREAD_TEXT:
		if this.Content, err = r.ReadString(config.THREAD_CONTENT_MAX_LENGTH); err != nil {
			return
		}
	case thread_type.THREAD_IMAGE, thread_type.THREAD_LINK:
		if b, err = r.ReadByte(); err != nil {
			return
		}
		this.Links = make([]string, b)
		for i := range this.Links {
			if this.Links[i], err = r.ReadString(config.THREAD_LINK_MAX_LENGTH); err != nil {
				return
			}
		}
	default:
		return errors.New("invalid listing type")
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

func (this *Thread) SetIdentityNow() (err error) {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, false, true, true)
	this.Identity, err = addresses.CreateAddr(cryptography.SHA3(w.Bytes()))
	return
}

func (this *Thread) GetMessageForSigningValidator() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, false, false, false)
	return w.Bytes()
}

func (this *Thread) GetMessageForSigningPublisher() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, true, true, true, false)
	return w.Bytes()
}

func (this *Thread) IsDeletable() bool {
	return false
}

func (this *Thread) Validate() error {

	switch this.Version {
	case THREAD_VERSION:
	default:
		return errors.New("invalid thread")
	}

	if this.FederationIdentity == nil {
		return errors.New("federation is not set")
	}

	switch this.Type {
	case thread_type.THREAD_TEXT:
		if len(this.Content) > config.THREAD_CONTENT_MAX_LENGTH {
			return errors.New("thread content is invalid")
		}
	case thread_type.THREAD_IMAGE, thread_type.THREAD_LINK:
		if len(this.Links) > config.THREAD_LINKS_MAX_COUNT {
			return errors.New("thread links invalid count")
		}
		for i := range this.Links {
			if len(this.Links[i]) > config.THREAD_LINK_MAX_LENGTH {
				return errors.New("thread link is invalid")
			}
			if _, err := url.ParseRequestURI(this.Links[i]); err != nil {
				return errors.New("invalid image url")
			}
		}
	default:
		return errors.New("invalid type")
	}

	if this.Type != thread_type.THREAD_TEXT && len(this.Content) > 0 {
		return errors.New("thread content should be empty")
	}

	if this.Type != thread_type.THREAD_LINK && this.Type != thread_type.THREAD_IMAGE && len(this.Links) > 0 {
		return errors.New("thread links content should be empty")
	}

	if len(this.Title) > config.THREAD_TITLE_MAX_LENGTH {
		return errors.New("thread title is too large")
	}
	if len(this.Title) < config.THREAD_TITLE_MIN_LENGTH {
		return errors.New("thread title is too short")
	}

	if len(this.Keywords) > config.THREAD_KEYWORDS_MAX_COUNT {
		return errors.New("thread has too many keywords")
	}
	dict := make(map[string]bool)
	for _, x := range this.Keywords {
		if len(x) >= config.THREAD_KEYWORD_MAX_LENGTH {
			return errors.New("thread keyword length is invalid")
		}
		dict[x] = true
	}
	if len(dict) != len(this.Keywords) {
		return errors.New("thread keywords repeated itself")
	}
	if len(this.Keywords) == 0 {
		return errors.New("thread should have at least one keyword")
	}

	return nil
}

func (this *Thread) ValidateSignatures() error {
	if !this.Validation.Verify(this.GetMessageForSigningValidator, nil) {
		return errors.New("listing validation signature failed")
	}
	if !this.Publisher.Verify(this.GetMessageForSigningPublisher) {
		return errors.New("listing publisher signature failed")
	}
	return nil
}

func (this *Thread) GetWords() []string {

	words := advanced_strings.SplitterSeparators(this.Title)

	list := []string{}
	for _, word := range words {
		if len(word) < 4 {
			continue
		}
		if len(list) > 10 {
			break
		}
		list = append(list, strings.ToLower(word))
	}
	return list
}

func (this *Thread) GetScore(votes float64) float64 {

	t := float64(this.Validation.Timestamp) / 45000
	f := math.Log2(math.Max(math.Abs(votes), 1))

	s := float64(1)
	if votes == 0 {
		s = 0
	} else if votes < 0 {
		s = -1
	}

	return helpers.RoundFloat(s*f+t/45000, 7)
}
