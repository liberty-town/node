//go:build !wasm
// +build !wasm

package api_method_store_comment

import (
	"errors"
	"liberty-town/node/config"
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/federations/federation_store/store_data/comments"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/api_implementation/api_websockets/api_method_sync_notification"
	"liberty-town/node/network/websocks"
	"liberty-town/node/network/websocks/connection/advanced_connection_types"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"net/http"
)

func MethodStoreComment(r *http.Request, args *api_types.APIMethodStoreIdentityRequest, reply *api_types.APIMethodStoreResult) error {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return errors.New("fed not set")
	}

	item := &comments.Comment{
		FederationIdentity: f.Federation.Ownership.Address,
		ParentIdentity:     args.Identity,
	}
	if err := item.Deserialize(advanced_buffers.NewBufferReader(args.Data)); err != nil {
		return err
	}

	if err := federation_store.StoreComments(item); err != nil {
		return err
	}

	go func() {
		websocks.Websockets.BroadcastJSON([]byte("sync-not"), &api_method_sync_notification.APIMethodSyncNotificationRequest{sync_type.SYNC_COMMENTS, item.Identity.Encoded, 0}, map[config.NodeConsensusType]bool{config.NODE_CONSENSUS_TYPE_FULL: true}, advanced_connection_types.UUID_ALL, 0)
	}()

	reply.Result = true
	return nil
}
