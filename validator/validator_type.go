package validator

type ValidatorVersion uint64

const (
	VALIDATOR_VERSION ValidatorVersion = iota
)

type validatorProof struct {
	Message   []byte `json:"message" msgpack:"message"`
	Size      uint64 `json:"size" msgpack:"size"`
	Signature []byte `json:"signature" msgpack:"signature"`
}
