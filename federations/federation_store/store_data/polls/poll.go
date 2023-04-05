package polls

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/federations/federation_store/store_data/polls/vote"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

type Poll struct {
	Version            PollVersion        `json:"version" msgpack:"version"`
	FederationIdentity *addresses.Address `json:"federation" msgpack:"federation"` //not serialized
	Identity           *addresses.Address `json:"identity" msgpack:"identity"`     //not serialized
	List               []*vote.Vote       `json:"list" msgpack:"list"`
}

func (this *Poll) Serialize(w *advanced_buffers.BufferWriter) {

	w.WriteUvarint(uint64(this.Version))
	w.WriteByte(byte(len(this.List)))

	for i := range this.List {
		this.List[i].Serialize(w)
	}

}

func (this *Poll) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	switch PollVersion(n) {
	case POLL_VERSION:
		this.Version = PollVersion(n)
	default:
		return errors.New("invalid version")
	}

	var b byte
	if b, err = r.ReadByte(); err != nil {
		return
	}

	this.List = make([]*vote.Vote, b)
	for i := range this.List {
		this.List[i] = &vote.Vote{}
		this.List[i].Identity = this.Identity
		this.List[i].FederationIdentity = this.FederationIdentity
		if err = this.List[i].Deserialize(r); err != nil {
			return
		}
	}

	return
}

func (this *Poll) IsDeletable() bool {
	return false
}

func (this *Poll) Validate() error {
	if len(this.List) == 0 {
		return errors.New("polls list is empty")
	}
	dict := make(map[string]bool)
	for i := range this.List {
		dict[this.List[i].Validation.Address.Encoded] = true
		if err := this.List[i].Validate(); err != nil {
			return err
		}
	}
	if len(dict) != len(this.List) {
		return errors.New("duplicates detected")
	}

	return nil
}

func (this *Poll) ValidateSignatures() error {
	for i := range this.List {
		if err := this.List[i].ValidateSignatures(); err != nil {
			return err
		}
	}
	return nil
}

// make sure it is valid
func (this *Poll) MergeVote(vote *vote.Vote) bool {
	for i := range this.List {
		if this.List[i].Validation.Address.Equals(vote.Validation.Address) {
			if this.List[i].GetBetterScore() >= vote.GetBetterScore() {
				return false
			}
			this.List[i] = vote
			return true
		}
	}
	this.List = append(this.List, vote)
	return true
}

func (this *Poll) GetBetterScore() (score uint64) {
	for i := range this.List {
		score += this.List[i].Upvotes
		score += this.List[i].Downvotes
	}
	return
}

func (this *Poll) GetScore() (score float64) {
	if this == nil {
		return 0
	}
	for i := range this.List {
		score += float64(this.List[i].Upvotes)
		score -= float64(this.List[i].Downvotes)
	}
	return
}
