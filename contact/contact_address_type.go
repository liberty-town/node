package contact

type ContactAddressType byte

const (
	CONTACT_ADDRESS_TYPE_WEBSOCKET_SERVER ContactAddressType = iota
	CONTACT_ADDRESS_TYPE_HTTP_SERVER
	CONTACT_ADDRESS_TYPE_ONION_SERVER
)
