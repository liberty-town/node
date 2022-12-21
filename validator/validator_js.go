//go:build wasm
// +build wasm

package validator

func (this *Validator) processValidate(validate func([]byte) []byte, challengeUri string, proof *validatorProof, data []byte) ([]byte, error) {
	return validate(data), nil
}
