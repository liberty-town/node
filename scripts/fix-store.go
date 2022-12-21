package main

import (
	"liberty-town/node/app"
	"liberty-town/node/config"
	"liberty-town/node/config/arguments"
	"liberty-town/node/federations"
	"liberty-town/node/federations/chat/chat_message"
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/federations/federation_store"
	"liberty-town/node/federations/federation_store/store_data/accounts"
	"liberty-town/node/federations/federation_store/store_data/accounts_summaries"
	"liberty-town/node/federations/federation_store/store_data/listings"
	"liberty-town/node/federations/federation_store/store_data/listings_summaries"
	"liberty-town/node/federations/federation_store/store_data/reviews"
	"liberty-town/node/gui"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/settings"
	"os"
)

func main() {

	var err error
	if err = arguments.InitArguments(os.Args[1:]); err != nil {
		panic(err)
	}
	if err = gui.InitGUI(); err != nil {
		panic(err)
	}
	if err = config.InitConfig(); err != nil {
		panic(err)
	}
	if err = federations.InitializeFederations(); err != nil {
		return
	}
	if err = settings.ImportMnemonic("", false); err != nil {
		return
	}
	if err = app.Init(); err != nil {
		return
	}

	list, _, err := federation_store.GetSyncList(sync_type.SYNC_ACCOUNTS, 1000)
	if err != nil {
		panic(err)
	}

	for _, it := range list {
		data, err := federation_store.GetAccount(it)
		if err != nil {
			panic(err)
		}
		obj := &accounts.Account{}
		if err := obj.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			panic(err)
		}
		if err := federation_store.StoreAccount(obj); err != nil {
			panic(err)
		}
	}

	list, _, err = federation_store.GetSyncList(sync_type.SYNC_LISTINGS, 1000)
	if err != nil {
		panic(err)
	}

	for _, it := range list {
		data, err := federation_store.GetListing(it)
		if err != nil {
			panic(err)
		}
		obj := &listings.Listing{}
		if err := obj.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			panic(err)
		}
		if err := federation_store.StoreListing(obj); err != nil {
			panic(err)
		}
	}

	list, _, err = federation_store.GetSyncList(sync_type.SYNC_LISTINGS_SUMMARIES, 1000)
	if err != nil {
		panic(err)
	}

	for _, it := range list {
		data, err := federation_store.GetListingSummary(it)
		if err != nil {
			panic(err)
		}
		obj := &listings_summaries.ListingSummary{}
		if err := obj.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			panic(err)
		}
		if err := federation_store.StoreListingSummary(obj); err != nil {
			panic(err)
		}
	}

	list, _, err = federation_store.GetSyncList(sync_type.SYNC_ACCOUNTS_SUMMARIES, 1000)
	if err != nil {
		panic(err)
	}

	for _, it := range list {
		data, err := federation_store.GetAccountSummary(it)
		if err != nil {
			panic(err)
		}
		obj := &accounts_summaries.AccountSummary{}
		if err := obj.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			panic(err)
		}
		if err := federation_store.StoreAccountSummary(obj); err != nil {
			panic(err)
		}
	}

	list, _, err = federation_store.GetSyncList(sync_type.SYNC_MESSAGES, 1000)
	if err != nil {
		panic(err)
	}

	for _, it := range list {
		data, err := federation_store.GetChatMessage(it)
		if err != nil {
			panic(err)
		}
		obj := &chat_message.ChatMessage{}
		if err := obj.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			panic(err)
		}
		if err := federation_store.StoreChatMessage(obj); err != nil {
			panic(err)
		}
	}

	list, _, err = federation_store.GetSyncList(sync_type.SYNC_REVIEWS, 1000)
	if err != nil {
		panic(err)
	}

	for _, it := range list {
		data, err := federation_store.GetReview(it)
		if err != nil {
			panic(err)
		}
		obj := &reviews.Review{}
		if err := obj.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			panic(err)
		}
		if err := federation_store.StoreReview(obj); err != nil {
			panic(err)
		}
	}

}
