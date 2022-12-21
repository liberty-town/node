//go:build !wasm
// +build !wasm

package federation_network_sync

import (
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/gui"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_list"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers/recovery"
	"math/rand"
	"time"
)

//检查更新
func ContinuouslyDownloadFederationData() {

	for i := 0; i < network_config.WEBSOCKETS_CONCURRENT_SYNC_CONNECTIONS; i++ {
		recovery.SafeGo(func() {

			for {

				<-websocks.Websockets.ReadyCn.Load()

				if conn := websocks.Websockets.GetRandomSocket(); conn != nil {

					syncType := sync_type.SyncVersion(rand.Intn(int(sync_type.SYNC_REVIEWS) + 1))
					data, err := connection.SendJSONAwaitAnswer[api_method_sync_list.APIMethodSyncListResult](conn, []byte("sync-list"), &api_method_sync_list.APIMethodSyncListRequest{
						Type: syncType,
					}, nil, 0)

					if err != nil {
						continue
					}

					if err = ProcessSync(conn, syncType, data.Keys, data.BetterScores); err != nil {
						gui.GUI.Info("connection was closed during syncing", err)
						conn.Close()
					}

				}

				time.Sleep(100 * time.Millisecond)
			}

		})
	}
}
