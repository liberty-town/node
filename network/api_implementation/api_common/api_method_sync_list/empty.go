package api_method_sync_list

import "liberty-town/node/federations/federation_network/sync_type"

type APIMethodSyncListRequest struct {
	Type sync_type.SyncVersion `json:"type" msgpack:"type"`
}

type APIMethodSyncListResult struct {
	Keys         []string `json:"keys" msgpack:"keys"`
	BetterScores []uint64 `json:"betterScores" msgpack:"betterScores"`
}
