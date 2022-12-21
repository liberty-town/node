//go:build wasm
// +build wasm

package store

import (
	"errors"
	"liberty-town/node/config/arguments"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_db/store_db_js"
	"liberty-town/node/store/store_db/store_db_memory"
)

func CreateStoreNow(name string, storeType string) (*Store, error) {

	var db store_db_interface.StoreDBInterface
	var err error

	switch storeType {
	case "js":
		db, err = store_db_js.CreateStoreDBJS(name)
	case "memory":
		db, err = store_db_memory.CreateStoreDBMemory(name)
	default:
		err = errors.New("Invalid --store-type argument: " + storeType)
	}

	if err != nil {
		return nil, err
	}

	return createStore("/"+name, db)
}

var AllowedStores = map[string]bool{"memory": true, "js": true}

func create_db() (err error) {

	if StoreSettings, err = CreateStoreNow("settings", GetStoreType(arguments.Arguments["--store-settings-type"].(string), AllowedStores)); err != nil {
		return
	}
	if StoreFederations, err = CreateStoreNow("federations", GetStoreType(arguments.Arguments["--store-settings-type"].(string), AllowedStores)); err != nil {
		return
	}

	return
}
