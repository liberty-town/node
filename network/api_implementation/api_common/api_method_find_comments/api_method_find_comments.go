package api_method_find_comments

import "liberty-town/node/addresses"

type APIMethodFindCommentsRequest struct {
	Identity *addresses.Address `json:"address" msgpack:"address"`
	Type     byte               `json:"type" msgpack:"type"`
	Start    int                `json:"start" msgpack:"start"`
}
