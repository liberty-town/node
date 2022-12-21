package api_method_sync_notification

import (
	msgpack "github.com/vmihailenco/msgpack/v5"
	"liberty-town/node/federations/federation_network/federation_network_sync"
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/network/websocks/connection"
)

type APIMethodSyncNotificationRequest struct {
	Type        sync_type.SyncVersion `json:"type" msgpack:"type"`
	Key         string                `json:"key" msgpack:"key"`
	BetterScore uint64                `json:"betterScore" msgpack:"betterScore"`
}

type APIMethodSyncNotificationResult struct {
	Result bool `json:"result"`
}

func MethodStoreSyncNotification(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	args := &APIMethodSyncNotificationRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIMethodSyncNotificationResult{}

	if err := federation_network_sync.ProcessSync(conn, args.Type, []string{args.Key}, []uint64{args.BetterScore}); err != nil {
		conn.Close()
		return nil, err
	}

	reply.Result = true
	return reply, nil
}
