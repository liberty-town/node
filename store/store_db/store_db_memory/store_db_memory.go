package store_db_memory

import (
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/store/store_db/store_db_interface"
	"sync"
)

type StoreDBMemory struct {
	store_db_interface.StoreDBInterface
	Name    []byte
	store   map[string][]byte
	rwmutex *sync.RWMutex
}

func (store *StoreDBMemory) Close() error {
	return nil
}

func (store *StoreDBMemory) View(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	store.rwmutex.RLock()
	defer store.rwmutex.RUnlock()

	tx := &StoreDBMemoryTransaction{
		store: store.store,
		local: &generics.Map[string, *StoreDBMemoryTransactionData]{},
	}
	return callback(tx)
}

func (store *StoreDBMemory) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	store.rwmutex.Lock()
	defer store.rwmutex.Unlock()

	tx := &StoreDBMemoryTransaction{
		store: store.store,
		local: &generics.Map[string, *StoreDBMemoryTransactionData]{},
		write: true,
	}

	err := callback(tx)

	if err == nil {
		if err = tx.writeTx(); err != nil {
			return err
		}
	}

	return nil
}

func CreateStoreDBMemory(name string) (*StoreDBMemory, error) {
	return &StoreDBMemory{
		Name:    []byte(name),
		store:   make(map[string][]byte),
		rwmutex: &sync.RWMutex{},
	}, nil

}
