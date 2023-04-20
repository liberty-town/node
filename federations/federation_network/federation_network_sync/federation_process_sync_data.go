//go:build !wasm
// +build !wasm

package federation_network_sync

import (
	"liberty-town/node/addresses"
	"liberty-town/node/federations/chat/chat_message"
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/federations/federation_store/store_data/accounts"
	"liberty-town/node/federations/federation_store/store_data/accounts_summaries"
	"liberty-town/node/federations/federation_store/store_data/comments"
	"liberty-town/node/federations/federation_store/store_data/listings"
	"liberty-town/node/federations/federation_store/store_data/listings_summaries"
	"liberty-town/node/federations/federation_store/store_data/polls"
	"liberty-town/node/federations/federation_store/store_data/reviews"
	"liberty-town/node/federations/federation_store/store_data/threads"
	"liberty-town/node/gui"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
)

func ProcessSync(conn *connection.AdvancedConnection, syncType sync_type.SyncVersion, keys []string, betterScores []uint64) error {

	download, err := federation_store.ProcessSyncList(syncType, keys, betterScores)
	if err != nil {
		return err
	}

	if len(download) > 0 {

		for i := range download {

			var addr *addresses.Address
			switch syncType {
			case sync_type.SYNC_COMMENTS, sync_type.SYNC_POLLS:
				if addr, err = addresses.DecodeAddr(download[i]); err != nil {
					return err
				}

			}

			var command string
			var result helpers.SerializableInterface
			switch syncType {
			case sync_type.SYNC_ACCOUNTS:
				command = "get-account"
				result = &accounts.Account{}
			case sync_type.SYNC_LISTINGS:
				command = "get-listing"
				result = &listings.Listing{}
			case sync_type.SYNC_LISTINGS_SUMMARIES:
				command = "get-listing-summary"
				result = &listings_summaries.ListingSummary{}
			case sync_type.SYNC_ACCOUNTS_SUMMARIES:
				command = "get-account-summary"
				result = &accounts_summaries.AccountSummary{}
			case sync_type.SYNC_MESSAGES:
				command = "get-msg"
				result = &chat_message.ChatMessage{}
			case sync_type.SYNC_REVIEWS:
				command = "get-review"
				result = &reviews.Review{}
			case sync_type.SYNC_THREADS:
				command = "get-thread"
				result = &threads.Thread{
					FederationIdentity: conn.Handshake.Federation,
				}
			case sync_type.SYNC_COMMENTS:
				command = "get-comment"
				result = &comments.Comment{
					FederationIdentity: conn.Handshake.Federation,
					ParentIdentity:     addr,
				}
			case sync_type.SYNC_POLLS:
				command = "get-poll"
				result = &polls.Poll{
					FederationIdentity: conn.Handshake.Federation,
					Identity:           addr,
				}
			default:
				return nil
			}

			data, err := connection.SendJSONAwaitAnswer[api_types.APIMethodGetResult](conn, []byte(command), &api_types.APIMethodGetRequest{
				download[i],
			}, nil, 0)

			if err != nil {
				return nil
			}

			if err = result.Deserialize(advanced_buffers.NewBufferReader(data.Result)); err != nil {
				return err
			}

			gui.GUI.Log("sync fetching", command, download[i])

			switch syncType {
			case sync_type.SYNC_ACCOUNTS:
				return federation_store.StoreAccount(result.(*accounts.Account))
			case sync_type.SYNC_LISTINGS:
				return federation_store.StoreListing(result.(*listings.Listing))
			case sync_type.SYNC_LISTINGS_SUMMARIES:
				return federation_store.StoreListingSummary(result.(*listings_summaries.ListingSummary))
			case sync_type.SYNC_ACCOUNTS_SUMMARIES:
				return federation_store.StoreAccountSummary(result.(*accounts_summaries.AccountSummary))
			case sync_type.SYNC_MESSAGES:
				return federation_store.StoreChatMessage(result.(*chat_message.ChatMessage))
			case sync_type.SYNC_REVIEWS:
				return federation_store.StoreReview(result.(*reviews.Review))
			case sync_type.SYNC_THREADS:
				return federation_store.StoreThread(result.(*threads.Thread))
			case sync_type.SYNC_COMMENTS:
				return federation_store.StoreComments(result.(*comments.Comment))
			case sync_type.SYNC_POLLS:
				return federation_store.StorePoll(result.(*polls.Poll))
			default:
				return nil
			}

		}

	}

	return nil
}
