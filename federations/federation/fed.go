package federation

import (
	"errors"
	"golang.org/x/exp/rand"
	"liberty-town/node/config"
	"liberty-town/node/contact"
	"liberty-town/node/federations/blockchain-nodes"
	"liberty-town/node/federations/category"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/moderator"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/validator"
	"liberty-town/node/validator/validation"
)

type Federation struct {
	Version        FederationVersion                  `json:"version"`
	Name           string                             `json:"name"`
	Description    string                             `json:"description"`
	Categories     []*category.Category               `json:"categories"`
	Validators     []*validator.Validator             `json:"validators"`
	Seeds          []*contact.Contact                 `json:"seeds"`
	Moderators     []*moderator.Moderator             `json:"moderators"`
	Ownership      *ownership.Ownership               `json:"ownership"`
	Nodes          []*blockchain_nodes.BlockchainNode `json:"nodes"`
	AcceptedAssets []string                           `json:"acceptedAssets"`
}

func (this *Federation) FindModerator(addr string) *moderator.Moderator {
	for i := range this.Moderators {
		if this.Moderators[i].Ownership.Address.Encoded == addr {
			return this.Moderators[i]
			break
		}
	}
	return nil
}

func (this *Federation) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeOwnershipSignature bool) {
	w.WriteUvarint(uint64(this.Version))
	w.WriteString(this.Name)
	w.WriteString(this.Description)
	w.WriteByte(byte(len(this.Categories)))
	for i := range this.Categories {
		this.Categories[i].Serialize(w)
	}
	w.WriteUvarint(uint64(len(this.Validators)))
	for i := range this.Validators {
		this.Validators[i].Serialize(w)
	}
	w.WriteUvarint(uint64(len(this.Seeds)))
	for i := range this.Seeds {
		this.Seeds[i].Serialize(w)
	}
	w.WriteUvarint(uint64(len(this.Nodes)))
	for i := range this.Nodes {
		this.Nodes[i].Serialize(w)
	}

	w.WriteUvarint(uint64(len(this.Moderators)))
	for i := range this.Moderators {
		this.Moderators[i].Serialize(w)
	}
	w.WriteUvarint(uint64(len(this.AcceptedAssets)))
	for i := range this.AcceptedAssets {
		w.WriteString(this.AcceptedAssets[i])
	}
	this.Ownership.AdvancedSerialize(w, includeOwnershipSignature)
}

func (this *Federation) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true)
}

func (this *Federation) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	var b byte
	var n uint64

	if b, err = r.ReadByte(); err != nil {
		return
	}

	switch FederationVersion(b) {
	case FEDERATION_VERSION:
	default:
		return errors.New("invalid federation version")
	}

	this.Version = FederationVersion(b)

	if this.Name, err = r.ReadString(100); err != nil {
		return
	}
	if this.Description, err = r.ReadString(5 * 1024); err != nil {
		return
	}

	if b, err = r.ReadByte(); err != nil {
		return err
	}
	this.Categories = make([]*category.Category, b)
	for i := range this.Categories {
		this.Categories[i] = &category.Category{}
		if err = this.Categories[i].Deserialize(r); err != nil {
			return
		}
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if n > 1000 {
		return errors.New("invalid number of validators")
	}
	this.Validators = make([]*validator.Validator, n)
	for i := range this.Validators {
		this.Validators[i] = &validator.Validator{}
		if err = this.Validators[i].Deserialize(r); err != nil {
			return
		}
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if n > 1000 {
		return errors.New("invalid number of seeds")
	}
	this.Seeds = make([]*contact.Contact, n)
	for i := range this.Seeds {
		this.Seeds[i] = &contact.Contact{}
		if err = this.Seeds[i].Deserialize(r); err != nil {
			return
		}
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if n > 1000 {
		return errors.New("invalid number of nodes")
	}
	this.Nodes = make([]*blockchain_nodes.BlockchainNode, n)
	for i := range this.Nodes {
		this.Nodes[i] = &blockchain_nodes.BlockchainNode{}
		if err = this.Nodes[i].Deserialize(r); err != nil {
			return
		}
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if n > 1000 {
		return errors.New("invalid number of moderators")
	}
	this.Moderators = make([]*moderator.Moderator, n)
	for i := range this.Moderators {
		this.Moderators[i] = &moderator.Moderator{}
		if err = this.Moderators[i].Deserialize(r); err != nil {
			return
		}
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if n > 1000 {
		return errors.New("invalid number of assets")
	}
	this.AcceptedAssets = make([]string, n)
	for i := range this.AcceptedAssets {
		if this.AcceptedAssets[i], err = r.ReadString(config.ACCEPTED_ASSET_LENGTH); err != nil {
			return
		}
	}

	this.Ownership = &ownership.Ownership{}
	if err = this.Ownership.Deserialize(r, this.GetMessageToSign); err != nil {
		return
	}

	return
}

func (this *Federation) GetMessageToSign() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false)
	return w.Bytes()
}

func (this *Federation) IsBetter(old *Federation) bool {
	return old == nil || old.GetBetterScore() < this.GetBetterScore()
}

func (this *Federation) GetBetterScore() uint64 {
	return this.Ownership.Timestamp / 15 / 60
}

func (this *Federation) IsValidationAccepted(validation *validation.Validation) bool {
	for _, v := range this.Validators {
		if validation.Address.Equals(v.Ownership.Address) {
			return true
		}
	}
	return false
}

func (this *Federation) Validate() error {

	switch this.Version {
	case FEDERATION_VERSION:
	default:
		return errors.New("invalid federation type")
	}

	if len(this.Name) < 5 || len(this.Name) > 100 {
		return errors.New("federation name is invalid")
	}
	if len(this.Description) > 5*1024 {
		return errors.New("federation description is invalid")
	}

	if len(this.Categories) < 2 {
		return errors.New("too fee categories")
	}

	uniqueCategories := make(map[uint64]bool)
	for i := range this.Categories {
		if err := this.Categories[i].Process(uniqueCategories); err != nil {
			return err
		}
	}

	if len(this.Validators) == 0 {
		return errors.New("validators list is empty")
	}
	for i := range this.Validators {
		if err := this.Validators[i].Validate(); err != nil {
			return err
		}
	}
	if len(this.Seeds) == 0 {
		return errors.New("seeds list is empty")
	}
	for i := range this.Seeds {
		if err := this.Seeds[i].Validate(); err != nil {
			return err
		}
	}

	if len(this.Nodes) == 0 {
		return errors.New("nodes list is empty")
	}
	for i := range this.Nodes {
		if err := this.Nodes[i].Validate(); err != nil {
			return err
		}
	}

	if len(this.AcceptedAssets) == 0 {
		return errors.New("accepted assets list is empty")
	}

	return nil
}

func (this *Federation) SignValidation(getMessage func() []byte, validate func([]byte) []byte) (*validation.Validation, error) {
	all := make(map[int]bool)
	for len(all) < len(this.Validators) {
		index := rand.Intn(len(this.Validators))
		if all[index] {
			continue
		}
		all[index] = true
		v := this.Validators[index]
		validation, err := v.SignValidation(getMessage, validate)
		if err == nil {
			return validation, nil
		}

		if err != nil && err.Error() != "validator no response" {
			return nil, err
		}
	}
	return nil, errors.New("no validator online")
}

func (this *Federation) ValidateSignatures() error {
	if !this.Ownership.Verify(this.GetMessageToSign) {
		return errors.New("listing ownership signature failed")
	}
	return nil
}

func (this *Federation) GetSeeds() []string {
	v := make([]string, len(this.Seeds))
	for i, c := range this.Seeds {
		v[i] = c.GetAddress(contact.CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER)
	}
	return v
}
