package api_method_sync_item

import "liberty-town/node/federations/federation_network/sync_type"

type APIMethodSyncItemRequest struct {
	Type sync_type.SyncVersion `json:"type" msgpack:"type"`
	Key  string                `json:"key" msgpack:"key"`
}

type APIMethodSyncItemResult struct {
	BetterScore uint64 `json:"betterScore" msgpack:"betterScore"`
}
