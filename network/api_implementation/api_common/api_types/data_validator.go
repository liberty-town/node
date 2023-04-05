package api_types

import (
	"liberty-town/node/addresses"
	"liberty-town/node/validator/validation/validation_type"
)

type ValidatorCheckVoteExtraRequest struct {
	Vote     int                `json:"vote" msgpack:"vote"`
	Identity *addresses.Address `json:"identity" msgpack:"identity"`
}

type ValidatorCheckExtraVersionRequest uint64

const (
	VALIDATOR_EXTRA_VOTE ValidatorCheckExtraVersionRequest = iota
)

type ValidatorCheckExtraRequest struct {
	Version ValidatorCheckExtraVersionRequest `json:"version" msgpack:"version"`
	Data    any                               `json:"data" msgpack:"data"`
}

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
	Version   uint64                      `json:"version" msgpack:"version"`
	Message   []byte                      `json:"message" msgpack:"message"`
	Size      uint64                      `json:"size" msgpack:"size"`
	Signature []byte                      `json:"signature" msgpack:"signature"`
	Solution  []byte                      `json:"solution" msgpack:"solution"`
	Extra     *ValidatorCheckExtraRequest `json:"extra" msgpack:"extra"`
}

type ValidatorSolutionVoteExtraResult struct {
	Upvotes   uint64 `json:"up" msgpack:"up"`
	Downvotes uint64 `json:"down" msgpack:"down"`
}

type ValidatorSolutionResult struct {
	Nonce     []byte `json:"nonce" msgpack:"nonce"`
	Timestamp uint64 `json:"timestamp" msgpack:"timestamp"`
	Signature []byte `json:"signature" msgpack:"signature"`
	Extra     any    `json:"extra" msgpack:"extra"`
}
