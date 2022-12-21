package config_coins

import (
	"encoding/base64"
	"liberty-town/node/pandora-pay/cryptography"
)

const (
	ASSET_LENGTH = cryptography.PublicKeyHashSize
)

var (
	NATIVE_ASSET_FULL               = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	NATIVE_ASSET_FULL_STRING        = string(NATIVE_ASSET_FULL)
	NATIVE_ASSET_FULL_STRING_BASE64 = base64.StdEncoding.EncodeToString(NATIVE_ASSET_FULL)
)
