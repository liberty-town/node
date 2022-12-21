package api_method_find_reviews

import "liberty-town/node/addresses"

type APIMethodFindReviewsRequest struct {
	Identity *addresses.Address `json:"identity" msgpack:"identity"`
	Type     byte               `json:"type" msgpack:"type"`
	Start    int                `json:"start" msgpack:"start"`
}
