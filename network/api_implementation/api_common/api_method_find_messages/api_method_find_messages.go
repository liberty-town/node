package api_method_find_messages

import "liberty-town/node/addresses"

type APIMethodFindMessagesRequest struct {
	First  *addresses.Address `json:"first" msgpack:"first"`
	Second *addresses.Address `json:"second" msgpack:"second"`
	Start  int                `json:"start" msgpack:"start"`
}
