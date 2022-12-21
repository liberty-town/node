package api_types

import "liberty-town/node/validator/validation/validation_type"

type ValidatorCheckRequest struct {
	Version   uint64 `json:"version" msgpack:"version"`
	Message   []byte `json:"message" msgpack:"message"`
	Size      uint64 `json:"size" msgpack:"size"`
	Signature []byte `json:"signature" msgpack:"signature"`
}

type ValidatorCheckResult struct {
	Challenge    validation_type.ValidatorChallengeType `json:"challenge" msgpack:"challenge"`
	Required     bool                                   `json:"required" msgpack:"required"`
	ChallengeUri string                                 `json:"challengeUri" msgpack:"challengeUri"`
	Data         []byte                                 `json:"data" msgpack:"data"`
}

type ValidatorSolutionRequest struct {
	Version   uint64 `json:"version" msgpack:"version"`
	Message   []byte `json:"message" msgpack:"message"`
	Size      uint64 `json:"size" msgpack:"size"`
	Signature []byte `json:"signature" msgpack:"signature"`
	Solution  []byte `json:"solution" msgpack:"solution"`
}

type ValidatorSolutionResult struct {
	Nonce     []byte `json:"nonce" msgpack:"nonce"`
	Timestamp uint64 `json:"timestamp" msgpack:"timestamp"`
	Signature []byte `json:"signature" msgpack:"signature"`
}
