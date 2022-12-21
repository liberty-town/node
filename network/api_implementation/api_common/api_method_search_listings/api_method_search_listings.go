package api_method_search_listings

import "liberty-town/node/federations/federation_store/store_data/listings/listing_type"

type APIMethodSearchListingsRequest struct {
	Type      listing_type.ListingType `json:"type" msgpack:"type"`
	Query     []string                 `json:"query" msgpack:"query"`
	QueryType byte                     `json:"queryType" msgpack:"queryType"`
	Start     int                      `json:"start" msgpack:"start"`
}
