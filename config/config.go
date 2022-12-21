package config

import (
	"errors"
	semver "github.com/blang/semver/v4"
	"liberty-town/node/config/arguments"
	"math/rand"
	"runtime"
	"time"
)

var (
	DEBUG            = false
	CPU_THREADS      = 1
	ARCHITECTURE     = ""
	OS               = ""
	PROTOCOL_VERSION = uint64(0)
	NAME             = "LIBERTYTOWN"
	VERSION          = semver.MustParse("0.0.1-test.0")
	VERSION_STRING   = VERSION.String()
	ORIGINAL_PATH    = "" //the original path where the software is located
	INSTANCE         = ""
	INSTANCE_ID      = 0
)

const (
	MAIN_NET_NETWORK_BYTE        uint64 = 1
	MAIN_NET_NETWORK_BYTE_PREFIX        = "LIBERTY" // must have 7 characters
	MAIN_NET_NETWORK_NAME               = "MAIN"    // must have 7 characters
	TEST_NET_NETWORK_BYTE        uint64 = 1034
	TEST_NET_NETWORK_BYTE_PREFIX        = "TIBERTY" // must have 7 characters
	TEST_NET_NETWORK_NAME               = "TEST"    // must have 7 characters
	DEV_NET_NETWORK_BYTE         uint64 = 4256
	DEV_NET_NETWORK_BYTE_PREFIX         = "DIBERTY" // must have 7 characters
	DEV_NET_NETWORK_NAME                = "DEV"     // must have 7 characters
	NETWORK_BYTE_PREFIX_LENGTH          = 7
)

var (
	NETWORK_SELECTED             uint64 = MAIN_NET_NETWORK_BYTE
	NETWORK_SELECTED_BYTE_PREFIX        = MAIN_NET_NETWORK_BYTE_PREFIX
	NETWORK_SELECTED_NAME               = MAIN_NET_NETWORK_NAME
	BUILD_VERSION                       = ""
)

var (
	NODE_CONSENSUS NodeConsensusType = NODE_CONSENSUS_TYPE_FULL
)

const (
	LIST_SIZE       = 500               //列表长度
	ITEM_EXPIRATION = 60 * 24 * 60 * 60 //日
	CONCURENCY      = 10
)

const (
	ACCEPTED_ASSET_LENGTH = 10
)

func InitConfig() (err error) {

	if arguments.Arguments["--network"] == "mainnet" {

	} else if arguments.Arguments["--network"] == "testnet" {
		NETWORK_SELECTED = TEST_NET_NETWORK_BYTE
		NETWORK_SELECTED_NAME = TEST_NET_NETWORK_NAME
		NETWORK_SELECTED_BYTE_PREFIX = TEST_NET_NETWORK_BYTE_PREFIX
	} else if arguments.Arguments["--network"] == "devnet" {
		NETWORK_SELECTED = DEV_NET_NETWORK_BYTE
		NETWORK_SELECTED_NAME = DEV_NET_NETWORK_NAME
		NETWORK_SELECTED_BYTE_PREFIX = DEV_NET_NETWORK_BYTE_PREFIX
	} else {
		return errors.New("selected --network is invalid. Accepted only: mainnet, testnet, devnet")
	}

	if arguments.Arguments["--debug"] == true {
		DEBUG = true
	}

	switch arguments.Arguments["--node-consensus"] {
	case "full":
		NODE_CONSENSUS = NODE_CONSENSUS_TYPE_FULL
	case "app":
		NODE_CONSENSUS = NODE_CONSENSUS_TYPE_APP
	case "none":
		NODE_CONSENSUS = NODE_CONSENSUS_TYPE_NONE
	default:
		return errors.New("invalid consensus argument")
	}

	if err = config_init(); err != nil {
		return
	}

	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
	CPU_THREADS = runtime.GOMAXPROCS(0)
	ARCHITECTURE = runtime.GOARCH
	OS = runtime.GOOS
}
