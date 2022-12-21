//go:build !wasm
// +build !wasm

package store

import (
	"errors"
	"liberty-town/node/config/arguments"
	"liberty-town/node/store/store_db/store_db_bolt"
	"liberty-town/node/store/store_db/store_db_bunt"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_db/store_db_memory"
)

func CreateStoreNow(name, storeType string) (*Store, error) {

	name = "/" + name

	var db store_db_interface.StoreDBInterface
	var err error

	switch storeType {
	case "bolt":
		db, err = store_db_bolt.CreateStoreDBBolt(name)
	case "bunt":
		db, err = store_db_bunt.CreateStoreDBBunt(name, false)
	case "bunt-memory":
		db, err = store_db_bunt.CreateStoreDBBunt(name, true)
	case "memory":
		db, err = store_db_memory.CreateStoreDBMemory(name)
	default:
		err = errors.New("invalid --store-type argument")
	}

	if err != nil {
		return nil, err
	}

	store, err := createStore(name, db)
	if err != nil {
		return nil, err
	}

	return store, nil
}

var AllowedStores = map[string]bool{"bolt": true, "bunt": true, "bunt-memory": true, "memory": true}

func create_db() (err error) {

	if StoreSettings, err = CreateStoreNow("settings", GetStoreType(arguments.Arguments["--store-settings-type"].(string), AllowedStores)); err != nil {
		return
	}
	if StoreFederations, err = CreateStoreNow("federations", GetStoreType(arguments.Arguments["--store-settings-type"].(string), AllowedStores)); err != nil {
		return
	}

	return
}
