package federation_store

import (
	"errors"
	"liberty-town/node/config"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/store/small_sorted_set"
	"liberty-town/node/store/store_db/store_db_interface"
)

func storeSortedSet(name, identity string, score float64, remove bool, tx store_db_interface.StoreDBTransactionInterface) (err error) {
	ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, name, tx)
	if err = ss.Read(); err != nil {
		return
	}
	if remove {
		ss.Delete(identity)
	} else {
		ss.PopLast()
		ss.Add(identity, score)
	}
	ss.Save()
	return
}

func GetData(table, identity string) (data []byte, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {
		data = tx.Get(table + identity)
		return nil
	})
	return
}

func FindData(table string, start, count int) (list []*api_types.APIMethodFindListItem, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, table, tx)
		if err := ss.Read(); err != nil {
			return err
		}

		for i := start; i < len(ss.Data) && len(list) < count; i++ {
			result := ss.Data[i]

			list = append(list, &api_types.APIMethodFindListItem{
				result.Key,
				float64(result.Score),
			})
		}

		return nil
	})
	return
}
