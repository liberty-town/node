package federation_store

import (
	"errors"
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func GetSyncItem(syncType sync_type.SyncVersion, key string) (betterScore uint64, err error) {
	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return 0, errors.New("not serving this federation")
	}

	name, err := syncType.GetStringStoreName()
	if err != nil {
		return
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {
		betterScore, err = store_utils.GetBetterScore(name, key, tx)
		return
	})

	return
}

func GetSyncList(syncType sync_type.SyncVersion, count uint64) (list []string, betterScores []uint64, err error) {
	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, nil, errors.New("not serving this federation")
	}

	name, err := syncType.GetStringStoreName()
	if err != nil {
		return
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {
		list, betterScores, err = store_utils.GetRandomItems(name, tx, count)
		return
	})

	return
}

func ProcessSyncList(syncType sync_type.SyncVersion, keys []string, betterScores []uint64) (list []string, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	name, err := syncType.GetStringStoreName()
	if err != nil {
		return
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		for i := range keys {

			betterScore, err := store_utils.GetBetterScore(name, string(keys[i]), tx)
			if err != nil {
				return err
			}

			if betterScore < betterScores[i] {
				list = append(list, keys[i])
			}

		}

		return
	})

	return
}
