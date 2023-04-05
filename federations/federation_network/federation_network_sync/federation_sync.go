package federation_network_sync

import (
	"errors"
	"liberty-town/node/federations"
	"liberty-town/node/federations/federation"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/gui"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_fed"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_fed"
	"liberty-town/node/network/websocks"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/recovery"
	"time"
)

// 检查新更新
func ContinuouslyDownloadFederation() {

	serveFed := federation_serve.ServeFederation.Load().Federation
	serveFedChangedCn := federation_serve.ServeFederationChangedMulticast.AddListener()
	defer federation_serve.ServeFederationChangedMulticast.RemoveChannel(serveFedChangedCn)

	recovery.SafeGo(func() {
		for {

			select {
			case newFed := <-serveFedChangedCn:
				if newFed != nil {
					serveFed = newFed.Federation
				}
			case <-websocks.Websockets.ReadyCn.Load():
			}

			if serveFed == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			list := websocks.Websockets.GetAllSockets()
			for _, conn := range list {

				data, err := connection.SendJSONAwaitAnswer[api_method_sync_fed.APIMethodSyncFedResult](conn, []byte("sync-fed"), &api_method_sync_fed.APIMethodSyncFedRequest{}, nil, 0)
				if err != nil {
					continue
				}

				//获取新版本
				if data.BetterScore > serveFed.GetBetterScore() {

					if err = func() (err error) {
						newFed, err := connection.SendJSONAwaitAnswer[api_method_get_fed.APIMethodGetFedResult](conn, []byte("get-fed"), &api_method_get_fed.APIMethodGetFedRequest{}, nil, 0)
						if err != nil {
							return
						}

						fed2 := &federation.Federation{}
						if err = fed2.Deserialize(advanced_buffers.NewBufferReader(newFed.Federation)); err != nil {
							return
						}

						if !fed2.Ownership.Address.Equals(serveFed.Ownership.Address) {
							return errors.New("fed mismatch")
						}

						if !fed2.IsBetter(serveFed) {
							return errors.New("fed is not better")
						}

						if err = fed2.Validate(); err != nil {
							return
						}

						if err = fed2.ValidateSignatures(); err != nil {
							return
						}

						federations.FederationsDict.Store(fed2.Ownership.Address.Encoded, fed2)
						if err = federation_serve.SetServeFederation(fed2, true); err != nil {
							return
						}

						return
					}(); err != nil {
						gui.GUI.Info("connection was closed during fed syncing", err)
						conn.Close()
						continue
					}

				}

				time.Sleep(1 * time.Millisecond)
			}

			time.Sleep(100 * time.Millisecond)

		}
	})
}
