package federation_serve

import (
	"liberty-town/node/config/arguments"
	"liberty-town/node/federations/federation"
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/pandora-pay/helpers/multicast"
	"liberty-town/node/store"
)

type serveFederationType struct {
	Federation *federation.Federation
	Store      *store.Store
}

var ServeFederation generics.Value[*serveFederationType]
var ServeFederationChangedMulticast *multicast.MulticastChannel[*serveFederationType]

func SetServeFederation(f *federation.Federation, saveStore bool) (err error) {

	old := ServeFederation.Load()
	if old != nil {
		if old.Federation.Ownership.Address.Equals(f.Ownership.Address) {
			ServeFederation.Store(&serveFederationType{
				f,
				old.Store,
			})
			if saveStore {
				if err = saveFederation(f); err != nil {
					return
				}
			}
			return
		}
	}

	serve := &serveFederationType{
		f,
		nil,
	}

	if serve.Store, err = store.CreateStoreNow(f.Ownership.Address.Encoded, store.GetStoreType(arguments.Arguments["--store-data-type"].(string), store.AllowedStores)); err != nil {
		return
	}

	ServeFederation.Store(serve)
	ServeFederationChangedMulticast.Broadcast(serve)

	if saveStore {
		if err = saveFederation(f); err != nil {
			return
		}
	}

	if old != nil {
		if err = old.Store.DB.Close(); err != nil {
			return
		}
	}

	return
}

func init() {
	ServeFederation.Store(nil)
	ServeFederationChangedMulticast = multicast.NewMulticastChannel[*serveFederationType]()
}
