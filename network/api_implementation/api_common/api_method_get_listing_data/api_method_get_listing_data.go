package api_method_get_listing_data

type APIMethodGetListingDataReply struct {
	Listing        []byte `json:"listing" msgpack:"listing"`
	AccountSummary []byte `json:"accountSummary" msgpack:"accountSummary"`
	ListingSummary []byte `json:"listingSummary" msgpack:"listingSummary"`
}
