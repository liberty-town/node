package validator

import (
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/contact"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/network/api_implementation/api_common/api_method_ping"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/settings"
	"liberty-town/node/validator/validation"
	"liberty-town/node/validator/validation/validation_type"
	"strings"
)

type Validator struct {
	Version   ValidatorVersion     `json:"version" msgpack:"version"`
	Contact   *contact.Contact     `json:"contact" msgpack:"contact"`
	Ownership *ownership.Ownership `json:"ownership" msgpack:"ownership"`
}

func (this *Validator) AdvancedSerialize(w *advanced_buffers.BufferWriter, includeOwnershipSignature bool) {
	w.WriteUvarint(uint64(this.Version))
	this.Contact.Serialize(w)
	this.Ownership.AdvancedSerialize(w, includeOwnershipSignature)
}

func (this *Validator) Serialize(w *advanced_buffers.BufferWriter) {
	this.AdvancedSerialize(w, true)
}

func (this *Validator) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	this.Version = ValidatorVersion(n)

	this.Contact = &contact.Contact{}
	if err = this.Contact.Deserialize(r); err != nil {
		return
	}

	this.Ownership = &ownership.Ownership{}
	if err = this.Ownership.Deserialize(r, this.GetMessageToSign); err != nil {
		return
	}

	return
}

func (this *Validator) GetMessageToSign() []byte {
	w := advanced_buffers.NewBufferWriter()
	this.AdvancedSerialize(w, false)
	return w.Bytes()
}

// 从服务器获取签名
func (this *Validator) sign(getMessage func() []byte, validate func([]byte) []byte, extra *api_types.ValidatorCheckExtraRequest) ([]byte, uint64, []byte, any, error) {

	pong, err := contact.Send[api_method_ping.APIPingReply](this.Contact, "ping", []byte{})
	if err != nil {
		return nil, 0, nil, nil, errors.New("validator no response")
	}
	if pong == nil || pong.Ping != "pong" {
		return nil, 0, nil, nil, errors.New("invalid pong")
	}

	message := getMessage()
	hash := cryptography.SHA3(message) //消息哈希

	wr := advanced_buffers.NewBufferWriter()
	wr.Write(hash)
	wr.WriteUvarint(uint64(len(message)))

	//签名消息
	mySignature, err := settings.Settings.Load().Validation.PrivateKey.Sign(wr.Bytes())
	if err != nil {
		return nil, 0, nil, nil, err
	}

	checkRequest := &api_types.ValidatorCheckRequest{
		0,
		hash,
		uint64(len(message)),
		mySignature,
	}

	var b []byte
	if b, err = json.Marshal(checkRequest); err != nil {
		return nil, 0, nil, nil, err
	}

	solutionRequest := &api_types.ValidatorSolutionRequest{
		0,
		hash,
		uint64(len(message)),
		mySignature,
		nil,
		extra,
	}

	//有时需要验证码
	var checkResult *api_types.ValidatorCheckResult
	if checkResult, err = contact.Send[api_types.ValidatorCheckResult](this.Contact, "check", b); err != nil {
		return nil, 0, nil, nil, err
	}

	if checkResult == nil {
		return nil, 0, nil, nil, errors.New("no result from validator")
	}

	//验证码的类型
	switch checkResult.Challenge {
	case validation_type.VALIDATOR_CHALLENGE_NO_CAPTCHA:
		solutionRequest.Solution = checkResult.Data
	//hChaptcha, reCaptcha
	//自定义生成验证码
	case validation_type.VALIDATOR_CHALLENGE_HCAPTCHA, validation_type.VALIDATOR_CHALLENGE_CUSTOM:

		if checkResult.Required {

			proof := &validatorProof{
				hash,
				uint64(len(message)),
				mySignature,
			}

			if b, err = json.Marshal(struct {
				Type         validation_type.ValidatorChallengeType `json:"type"`
				Origin       string                                 `json:"origin"`
				ChallengeUri string                                 `json:"challengeUri"`
				Data         string                                 `json:"data"`
				Proof        *validatorProof                        `json:"proof"`
			}{
				checkResult.Challenge,
				this.Contact.GetAddress(contact.CONTACT_ADDRESS_TYPE_HTTP_SERVER),
				checkResult.ChallengeUri,
				string(checkResult.Data),
				proof,
			}); err != nil {
				return nil, 0, nil, nil, err
			}

			//显示验证码
			if solutionRequest.Solution, err = this.processValidate(validate, checkResult.ChallengeUri, proof, b); err != nil {
				return nil, 0, nil, nil, err
			}

			if len(solutionRequest.Solution) == 0 {
				return nil, 0, nil, nil, errors.New("validation canceled")
			}

		}
	default:
		return nil, 0, nil, nil, errors.New("unknown type of challenge")
	}

	if b, err = json.Marshal(solutionRequest); err != nil {
		return nil, 0, nil, nil, err
	}

	for i := 0; i < 10; i++ {
		//请求数字签名
		var result *api_types.ValidatorSolutionResult
		if result, err = contact.Send[api_types.ValidatorSolutionResult](this.Contact, "solution", b); err != nil {
			if strings.Contains(err.Error(), "(Client.Timeout exceeded while awaiting headers)") {
				continue
			}
			return nil, 0, nil, nil, err
		}

		if extra != nil && extra.Version == api_types.VALIDATOR_EXTRA_VOTE {

			b, err := json.Marshal(result.Extra)
			if err != nil {
				return nil, 0, nil, nil, err
			}

			votePayload := &api_types.ValidatorSolutionVoteExtraResult{}
			if err = json.Unmarshal(b, votePayload); err != nil {
				return nil, 0, nil, nil, err
			}
			result.Extra = votePayload
		}

		return result.Nonce, result.Timestamp, result.Signature, result.Extra, nil
	}

	return nil, 0, nil, nil, errors.New("timeout")
}

// 数字签名
func (this *Validator) SignValidation(getMessage func() []byte, validate func([]byte) []byte, extra *api_types.ValidatorCheckExtraRequest) (*validation.Validation, any, error) {

	nonce, timestamp, signature, validationExtra, err := this.sign(getMessage, validate, extra)
	if err != nil {
		return nil, nil, err
	}

	v := &validation.Validation{
		validation_type.VALIDATION_VERSION_V0,
		nonce,
		timestamp,
		signature,
		nil,
	}

	var getExtraInfo func() []byte
	if extra != nil && extra.Version == api_types.VALIDATOR_EXTRA_VOTE {
		getExtraInfo = func() []byte {
			w := advanced_buffers.NewBufferWriter()
			w.WriteUvarint(validationExtra.(*api_types.ValidatorSolutionVoteExtraResult).Upvotes)
			w.WriteUvarint(validationExtra.(*api_types.ValidatorSolutionVoteExtraResult).Downvotes)
			return w.Bytes()
		}
	}

	if v.Address, err = addresses.CreateAddrFromSignature(v.GetMessageToValidator(getMessage, getExtraInfo), signature); err != nil {
		return nil, nil, err
	}

	return v, validationExtra, nil
}

func (this *Validator) Validate() error {
	switch this.Version {
	case VALIDATOR_VERSION:
	default:
		return errors.New("invalid validator type")
	}
	return nil
}
