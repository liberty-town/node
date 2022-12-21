package federation_serve

import (
	"liberty-town/node/addresses"
	"liberty-town/node/federations/federation"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/store"
	"liberty-town/node/store/store_db/store_db_interface"
)

func saveFederation(f *federation.Federation) (err error) {
	return store.StoreFederations.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {
		tx.Put("feds:"+string(helpers.SerializeToBytes(f.Ownership.Address)), helpers.SerializeToBytes(f))
		return nil
	})
}

func LoadFederation(address *addresses.Address) (fed *federation.Federation, err error) {

	err = store.StoreFederations.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		data := tx.Get("feds:" + string(helpers.SerializeToBytes(address)))
		if data == nil {
			return
		}

		fed = &federation.Federation{}
		if err = fed.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			return err
		}

		return
	})
	return
}
