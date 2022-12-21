package config

const (
	MAIN_NET_NETWORK_BYTE        uint64 = 1
	MAIN_NET_NETWORK_BYTE_PREFIX        = "PCASH" // must have 7 characters
	MAIN_NET_NETWORK_NAME               = "MAIN"  // must have 7 characters
	TEST_NET_NETWORK_BYTE        uint64 = 1034
	TEST_NET_NETWORK_BYTE_PREFIX        = "TCASH" // must have 7 characters
	TEST_NET_NETWORK_NAME               = "TEST"  // must have 7 characters
	DEV_NET_NETWORK_BYTE         uint64 = 4256
	DEV_NET_NETWORK_BYTE_PREFIX         = "DCASH" // must have 7 characters
	DEV_NET_NETWORK_NAME                = "DEV"   // must have 7 characters
	NETWORK_BYTE_PREFIX_LENGTH          = 5
)

var (
	NETWORK_SELECTED             = MAIN_NET_NETWORK_BYTE
	NETWORK_SELECTED_BYTE_PREFIX = MAIN_NET_NETWORK_BYTE_PREFIX
	NETWORK_SELECTED_NAME        = MAIN_NET_NETWORK_NAME
)

const (
	TRANSACTIONS_MAX_DATA_LENGTH = 512
	TRANSACTIONS_ZETHER_RING_MAX = 256
)
