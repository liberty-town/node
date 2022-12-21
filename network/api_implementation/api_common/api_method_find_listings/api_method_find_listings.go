package api_method_find_listings

import (
	"liberty-town/node/addresses"
	"liberty-town/node/federations/federation_store/store_data/listings/listing_type"
)

type APIMethodFindListingsRequest struct {
	Account *addresses.Address       `json:"account" msgpack:"account"`
	Type    listing_type.ListingType `json:"type" msgpack:"type"`
	Start   int                      `json:"start" msgpack:"start"`
}
