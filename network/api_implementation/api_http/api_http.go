package api_http

import (
	"io"
	"liberty-town/node/network/api_code/api_code_http"
	"liberty-town/node/network/api_implementation/api_common"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_conversations"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_listings"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_messages"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_reviews"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_account"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_account_summary"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_fed"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_last_msg"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_listing"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_listing_data"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_listing_summary"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_message"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_review"
	"liberty-town/node/network/api_implementation/api_common/api_method_ping"
	"liberty-town/node/network/api_implementation/api_common/api_method_search_listings"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_account"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_account_summary"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_listing"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_listing_summary"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_message"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_review"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_fed"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_list"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"net/url"
)

type API struct {
	GetMap    map[string]func(values url.Values) (interface{}, error)
	PostMap   map[string]func(values io.ReadCloser) (interface{}, error)
	apiCommon *api_common.APICommon
}

var ConfigureAPIRoutes func(api *API)

func NewAPI(apiCommon *api_common.APICommon) *API {

	api := &API{
		apiCommon: apiCommon,
	}

	api.GetMap = map[string]func(values url.Values) (interface{}, error){
		"ping":                  api_code_http.Handle[struct{}, api_method_ping.APIPingReply](api_method_ping.GetPing),
		"":                      api_code_http.Handle[struct{}, api_common.APIInfoReply](api.apiCommon.GetInfo),
		"network/nodes":         api_code_http.Handle[struct{}, api_common.APINetworkNodesReply](api.apiCommon.GetNetworkNodes),
		"get-listing":           api_code_http.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_listing.MethodGetListing),
		"get-listing-summary":   api_code_http.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_listing_summary.MethodGetListingSummary),
		"get-listing-data":      api_code_http.Handle[api_types.APIMethodGetRequest, api_method_get_listing_data.APIMethodGetListingDataReply](api_method_get_listing_data.MethodGetListingData),
		"get-account":           api_code_http.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_account.MethodGetAccount),
		"get-account-summary":   api_code_http.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_account_summary.MethodGetAccountSummary),
		"get-msg":               api_code_http.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_message.MethodGetMessage),
		"get-review":            api_code_http.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_review.MethodGetReview),
		"store-account":         api_code_http.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_account.MethodStoreAccount),
		"store-account-summary": api_code_http.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_account_summary.MethodStoreAccountSummary),
		"store-listing":         api_code_http.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_listing.MethodStoreListing),
		"store-listing-summary": api_code_http.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_listing_summary.MethodStoreListingSummary),
		"store-review":          api_code_http.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_review.MethodStoreReview),
		"store-msg":             api_code_http.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_message.MethodStoreMessage),
		"get-last-msg":          api_code_http.Handle[api_method_get_last_msg.APIMethodGetLastMessageRequest, api_types.APIMethodGetResult](api_method_get_last_msg.MethodGetLastMessage),
		"search-listings":       api_code_http.Handle[api_method_search_listings.APIMethodSearchListingsRequest, api_types.APIMethodFindListResult](api_method_search_listings.MethodSearchListings),
		"find-listings":         api_code_http.Handle[api_method_find_listings.APIMethodFindListingsRequest, api_types.APIMethodFindListResult](api_method_find_listings.MethodFindListings),
		"find-conversations":    api_code_http.Handle[api_method_find_conversations.APIMethodFindConversationsRequest, api_types.APIMethodFindListResult](api_method_find_conversations.MethodFindConversations),
		"find-reviews":          api_code_http.Handle[api_method_find_reviews.APIMethodFindReviewsRequest, api_types.APIMethodFindListResult](api_method_find_reviews.MethodFindReviews),
		"find-msgs":             api_code_http.Handle[api_method_find_messages.APIMethodFindMessagesRequest, api_types.APIMethodFindListResult](api_method_find_messages.MethodFindMessages),
		"sync-list":             api_code_http.Handle[api_method_sync_list.APIMethodSyncListRequest, api_method_sync_list.APIMethodSyncListResult](api_method_sync_list.MethodSyncList),
		"sync-fed":              api_code_http.Handle[api_method_sync_fed.APIMethodSyncFedRequest, api_method_sync_fed.APIMethodSyncFedResult](api_method_sync_fed.MethodSyncFed),
		"get-fed":               api_code_http.Handle[api_method_get_fed.APIMethodGetFedRequest, api_method_get_fed.APIMethodGetFedResult](api_method_get_fed.MethodGetFed),
	}

	if ConfigureAPIRoutes != nil {
		ConfigureAPIRoutes(api)
	}

	return api
}
