package store

import (
	"liberty-town/node/store/store_db/store_db_interface"
)

type Store struct {
	Name   string
	Opened bool
	DB     store_db_interface.StoreDBInterface
}

var StoreSettings *Store
var StoreFederations *Store

func (store *Store) close() error {
	return store.DB.Close()
}

func createStore(name string, db store_db_interface.StoreDBInterface) (*Store, error) {

	store := &Store{
		Name:   name,
		Opened: false,
		DB:     db,
	}

	store.Opened = true

	return store, nil
}

func InitDB() (err error) {
	return create_db()
}

func DBClose() (err error) {
	if err = StoreSettings.close(); err != nil {
		return
	}
	if err = StoreFederations.close(); err != nil {
		return
	}
	return
}

func GetStoreType(value string, allowed map[string]bool) string {

	if allowed[value] {
		return value
	}

	return ""
}
