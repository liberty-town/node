package federation_network

import (
	"liberty-town/node/federations/federation"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/gui"
	"liberty-town/node/network"
	"liberty-town/node/pandora-pay/helpers/recovery"
)

func setSeeds(fed *federation.Federation) {
	if err := network.Network.ImportSeeds(fed.GetSeeds()); err != nil {
		gui.GUI.Error("error importing seeds", err)
	}
}

func ConnectFederationSeeds() {

	recovery.SafeGo(func() {

		changedCn := federation_serve.ServeFederationChangedMulticast.AddListener()
		federation_serve.ServeFederationChangedMulticast.RemoveChannel(changedCn)

		for {
			fed := <-federation_serve.ServeFederationChangedMulticast.AddListener()
			setSeeds(fed.Federation)
		}
	})

	fed := federation_serve.ServeFederation.Load()
	if fed != nil {
		setSeeds(fed.Federation)
	}

}
