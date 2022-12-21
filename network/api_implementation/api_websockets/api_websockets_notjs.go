//go:build !wasm
// +build !wasm

package api_websockets

import (
	"liberty-town/node/network/api_code/api_code_websockets"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_conversations"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_listings"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_messages"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_reviews"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_account"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_account_summary"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_last_msg"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_listing"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_listing_data"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_listing_summary"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_message"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_review"
	"liberty-town/node/network/api_implementation/api_common/api_method_search_listings"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_account"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_account_summary"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_listing"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_listing_summary"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_message"
	"liberty-town/node/network/api_implementation/api_common/api_method_store_review"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_list"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/api_implementation/api_websockets/api_method_sync_notification"
)

func (api *APIWebsockets) initApi() {
	api.GetMap["get-listing"] = api_code_websockets.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_listing.MethodGetListing)
	api.GetMap["get-listing-summary"] = api_code_websockets.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_listing_summary.MethodGetListingSummary)
	api.GetMap["get-listing-data"] = api_code_websockets.Handle[api_types.APIMethodGetRequest, api_method_get_listing_data.APIMethodGetListingDataReply](api_method_get_listing_data.MethodGetListingData)
	api.GetMap["get-account"] = api_code_websockets.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_account.MethodGetAccount)
	api.GetMap["get-account-summary"] = api_code_websockets.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_account_summary.MethodGetAccountSummary)
	api.GetMap["get-msg"] = api_code_websockets.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_message.MethodGetMessage)
	api.GetMap["get-review"] = api_code_websockets.Handle[api_types.APIMethodGetRequest, api_types.APIMethodGetResult](api_method_get_review.MethodGetReview)
	api.GetMap["store-account"] = api_code_websockets.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_account.MethodStoreAccount)
	api.GetMap["store-account-summary"] = api_code_websockets.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_account_summary.MethodStoreAccountSummary)
	api.GetMap["store-listing"] = api_code_websockets.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_listing.MethodStoreListing)
	api.GetMap["store-listing-summary"] = api_code_websockets.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_listing_summary.MethodStoreListingSummary)
	api.GetMap["store-review"] = api_code_websockets.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_review.MethodStoreReview)
	api.GetMap["store-msg"] = api_code_websockets.Handle[api_types.APIMethodStoreRequest, api_types.APIMethodStoreResult](api_method_store_message.MethodStoreMessage)
	api.GetMap["get-last-msg"] = api_code_websockets.Handle[api_method_get_last_msg.APIMethodGetLastMessageRequest, api_types.APIMethodGetResult](api_method_get_last_msg.MethodGetLastMessage)
	api.GetMap["search-listings"] = api_code_websockets.Handle[api_method_search_listings.APIMethodSearchListingsRequest, api_types.APIMethodFindListResult](api_method_search_listings.MethodSearchListings)
	api.GetMap["find-listings"] = api_code_websockets.Handle[api_method_find_listings.APIMethodFindListingsRequest, api_types.APIMethodFindListResult](api_method_find_listings.MethodFindListings)
	api.GetMap["find-conversations"] = api_code_websockets.Handle[api_method_find_conversations.APIMethodFindConversationsRequest, api_types.APIMethodFindListResult](api_method_find_conversations.MethodFindConversations)
	api.GetMap["find-reviews"] = api_code_websockets.Handle[api_method_find_reviews.APIMethodFindReviewsRequest, api_types.APIMethodFindListResult](api_method_find_reviews.MethodFindReviews)
	api.GetMap["find-msgs"] = api_code_websockets.Handle[api_method_find_messages.APIMethodFindMessagesRequest, api_types.APIMethodFindListResult](api_method_find_messages.MethodFindMessages)
	api.GetMap["sync-list"] = api_code_websockets.Handle[api_method_sync_list.APIMethodSyncListRequest, api_method_sync_list.APIMethodSyncListResult](api_method_sync_list.MethodSyncList)
	api.GetMap["sync-not"] = api_method_sync_notification.MethodStoreSyncNotification
}
