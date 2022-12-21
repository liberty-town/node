package connection

import (
	"errors"
	semver "github.com/blang/semver/v4"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/federations/federation_serve"
)

type ConnectionHandshake struct {
	Name       string                   `json:"name" msgpack:"name"`
	Version    string                   `json:"version" msgpack:"version"`
	Network    uint64                   `json:"network" msgpack:"network"`
	Consensus  config.NodeConsensusType `json:"consensus" msgpack:"consensus"`
	Federation *addresses.Address       `json:"fed" msgpack:"fed"`
	URL        string                   `json:"url" msgpack:"url"`
}

func (handshake *ConnectionHandshake) ValidateHandshake() (*semver.Version, error) {

	if handshake.Network != config.NETWORK_SELECTED {
		return nil, errors.New("Network is different")
	}

	switch handshake.Consensus {
	case config.NODE_CONSENSUS_TYPE_NONE:
	case config.NODE_CONSENSUS_TYPE_FULL:
	case config.NODE_CONSENSUS_TYPE_APP:
	default:
		return nil, errors.New("Invalid CONSENSUS")
	}

	if handshake.Federation == nil || handshake.Federation.Network != config.NETWORK_SELECTED {
		return nil, errors.New("invalid federation network")
	}

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("federation not init")
	}
	if !handshake.Federation.Equals(f.Federation.Ownership.Address) {
		return nil, errors.New("federation is not served")
	}

	version, err := semver.Parse(handshake.Version)
	if err != nil {
		return nil, errors.New("Invalid VERSION format")
	}

	return &version, nil
}
