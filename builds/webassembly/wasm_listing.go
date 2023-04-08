package main

import (
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/accounts_summaries"
	"liberty-town/node/federations/federation_store/store_data/listings"
	"liberty-town/node/federations/federation_store/store_data/listings/listing_type"
	"liberty-town/node/federations/federation_store/store_data/listings/offer"
	"liberty-town/node/federations/federation_store/store_data/listings/shipping"
	"liberty-town/node/federations/federation_store/store_data/listings_summaries"
	"liberty-town/node/gui"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_listings"
	"liberty-town/node/network/api_implementation/api_common/api_method_search_listings"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_item"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/settings"
	"strconv"
	"syscall/js"
)

func listingStore(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		var err error

		req := &struct {
			Type              listing_type.ListingType `json:"type"`
			Title             string                   `json:"title"`
			Description       string                   `json:"description"`
			Images            []string                 `json:"images"`
			Categories        []uint64                 `json:"categories"`
			QuantityUnlimited bool                     `json:"quantityUnlimited"`
			QuantityAvailable uint64                   `json:"quantityAvailable"`
			ShipsFrom         uint64                   `json:"shipsFrom"`
			ShipsTo           []uint64                 `json:"shipsTo"`
			Offers            []*offer.Offer           `json:"offers"`
			Shipping          []*shipping.Shipping     `json:"shipping"`
			Nonce             []byte                   `json:"nonce,omitempty"`
		}{}

		if err = json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		if len(req.Nonce) == 0 {
			req.Nonce = cryptography.RandomHash()
		}

		account := settings.Settings.Load().Account

		key := cryptography.SHA3(cryptography.SHA3(append(account.PrivateKey.Key[:], req.Nonce...)))

		listingPrivateKey, err := addresses.NewPrivateKey(key)
		if err != nil {
			return nil, err
		}

		listingAddress, err := listingPrivateKey.GenerateAddress()
		if err != nil {
			return nil, err
		}

		it := &listings.Listing{
			listings.LISTING_VERSION,
			f.Federation.Ownership.Address,
			req.Nonce,
			listingAddress,
			req.Type,
			req.Title,
			req.Description,
			req.Categories,
			req.Images,
			req.QuantityUnlimited,
			req.QuantityAvailable,
			req.ShipsFrom,
			req.ShipsTo,
			req.Offers,
			req.Shipping,
			nil,
			&ownership.Ownership{},
			nil,
		}

		if err = it.Validate(); err != nil && err.Error() != "listing ownership identity does not match" {
			return nil, err
		}

		it.Ownership = &ownership.Ownership{}

		if it.Validation, _, err = federationValidate(f.Federation, it.GetMessageForSigningValidator, args[1], nil); err != nil {
			return nil, err
		}

		if err = it.Publisher.Sign(account.PrivateKey, it.GetMessageForSigningPublisher); err != nil {
			return nil, err
		}

		if err = it.Ownership.Sign(listingPrivateKey, it.GetMessageForSigningOwnership); err != nil {
			return nil, err
		}

		if err = it.Validate(); err != nil {
			return nil, err
		}

		results := 0
		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-listing", &api_types.APIMethodStoreRequest{helpers.SerializeToBytes(it)}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
			if a != nil && a.Result {
				results++
			}
			return true
		}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Listing *listings.Listing `json:"listing"`
			Results int               `json:"results"`
		}{it, results})

	})
}

func listingGet(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Listing *addresses.Address `json:"listing,omitempty"`
		}{}

		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		var listing *listings.Listing

		if err := federation_network.FetchData[api_types.APIMethodGetResult]("get-listing", &api_types.APIMethodGetRequest{
			req.Listing.Encoded,
		}, func(data *api_types.APIMethodGetResult, contact *connection.AdvancedConnection) bool {
			if data == nil || data.Result == nil {
				return true
			}
			temp := &listings.Listing{}
			if err := temp.Deserialize(advanced_buffers.NewBufferReader(data.Result)); err != nil {
				return true
			}
			if !temp.FederationIdentity.Equals(f.Federation.Ownership.Address) {
				return true
			}
			if temp.Validate() != nil || temp.ValidateSignatures() != nil {
				return true
			}
			if temp.IsBetter(listing) {
				listing = temp
			}
			return true
		}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(listing)

	})
}

func listingsSearch(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Query        []string                 `json:"query,omitempty"`
			Type         listing_type.ListingType `json:"type,omitempty"`
			QueryType    byte                     `json:"queryType,omitempty"`
			Start        int                      `json:"start"`
			ShippingType byte                     `json:"shippingType"`
			Shipping     uint64                   `json:"shipping"`
		}{}

		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		type SearchResult struct {
			Key            string                             `json:"key"`
			Score          float64                            `json:"score"`
			Listing        *listings.Listing                  `json:"listing"`
			AccountSummary *accounts_summaries.AccountSummary `json:"accountSummary"`
			ListingSummary *listings_summaries.ListingSummary `json:"listingSummary"`
		}

		count := 0

		list, err := federation_network.JustAggregateData("search-listings", &api_method_search_listings.APIMethodSearchListingsRequest{
			req.Type,
			req.Query,
			req.QueryType,
			req.Start,
			req.ShippingType,
			req.Shipping,
		})

		if err != nil {
			return nil, err
		}

		duplicates := make(map[string]bool)
		banned := make(map[string]bool)

		for _, searchResult := range list {

			if duplicates[searchResult.Key] || banned[searchResult.Conn.RemoteAddr] { //已经找到 || 已被禁止
				continue
			}

			if err = func() error {

				listing := &listings.Listing{}
				if err = federation_network.AggregateBestData[api_types.APIMethodGetResult]("sync-item", &api_method_sync_item.APIMethodSyncItemRequest{
					sync_type.SYNC_LISTINGS,
					searchResult.Key,
				}, "get-listing", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {
					if len(answer.Result) > 0 {
						if err := listing.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
							return err
						}

						if !listing.FederationIdentity.Equals(f.Federation.Ownership.Address) ||
							listing.Validate() != nil || listing.ValidateSignatures() != nil ||
							!f.Federation.IsValidationAccepted(listing.Validation) {
							return errors.New("listing was not accepted")
						}

						if listing.Identity.Encoded != searchResult.Key {
							return errors.New("listing identity mismatch")
						}
					}
					return nil
				}, banned); err != nil {
					return err
				}

				if listing == nil {
					return errors.New("listing not found")
				}

				var accountSummary *accounts_summaries.AccountSummary
				var listingSummary *listings_summaries.ListingSummary

				//pass banned
				if err = federation_network.AggregateBestData[api_types.APIMethodGetResult]("sync-item", &api_method_sync_item.APIMethodSyncItemRequest{
					sync_type.SYNC_ACCOUNTS_SUMMARIES,
					listing.Publisher.Address.Encoded,
				}, "get-account-summary", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {
					if len(answer.Result) > 0 {
						accountSummary = &accounts_summaries.AccountSummary{}
						if err := accountSummary.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
							return err
						}
						if accountSummary.Validate() != nil || accountSummary.ValidateSignatures() != nil || !f.Federation.IsValidationAccepted(accountSummary.Validation) {
							return errors.New("account summary is invalid")
						}
						if !accountSummary.AccountIdentity.Equals(listing.Publisher.Address) {
							return errors.New("accountSummary identity mismatch")
						}
					}
					return nil
				}, banned); err != nil {
					return err
				}

				if err = federation_network.AggregateBestData[api_types.APIMethodGetResult]("sync-item", &api_method_sync_item.APIMethodSyncItemRequest{
					sync_type.SYNC_LISTINGS_SUMMARIES,
					listing.Identity.Encoded,
				}, "get-listing-summary", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {
					if len(answer.Result) > 0 {
						listingSummary = &listings_summaries.ListingSummary{}
						if err := listingSummary.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
							return err
						}
						if listingSummary.Validate() != nil || listingSummary.ValidateSignatures() != nil || !f.Federation.IsValidationAccepted(listingSummary.Validation) {
							return errors.New("listing summary is invalid")
						}
						if !listingSummary.ListingIdentity.Equals(listing.Identity) {
							return errors.New("listingSummary identity mismatch")
						}
					}
					return nil
				}, banned); err != nil {
					return err
				}

				accountSummaryScore := accountSummary.GetScore(req.Type)
				listingSummaryScore := listingSummary.GetScore()
				foundScore := listings.GetScore(listingSummaryScore, accountSummaryScore)

				if searchResult.Score > foundScore {
					return errors.New("score is less than it should be")
				}

				var searchData []string
				switch req.QueryType {
				case 0:
					searchData = listing.GetWords()
				case 1:
					searchData = make([]string, len(listing.Categories))
					for i := range searchData {
						searchData[i] = strconv.FormatUint(listing.Categories[i], 10)
					}
				}

				for _, query := range req.Query {
					if query != "" {
						found := false
						for _, c := range searchData {
							if c == query {
								found = true
								break
							}
						}
						if !found {
							return errors.New("query not found")
						}
					}
				}

				duplicates[searchResult.Key] = true

				result := &SearchResult{searchResult.Key, foundScore, listing, accountSummary, listingSummary}
				b2, err := webassembly_utils.ConvertJSONBytes(result)
				if err != nil {
					return err
				}

				args[1].Invoke(b2)
				count++

				return nil
			}(); err != nil {
				gui.GUI.Error("banning connection", searchResult.Conn.RemoteAddr, err)
				banned[searchResult.Conn.RemoteAddr] = true
			}

		}

		return count, err
	})
}

func listingsGetAll(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Account *addresses.Address       `json:"account,omitempty"`
			Type    listing_type.ListingType `json:"type,omitempty"`
			Start   int                      `json:"start"`
		}{}

		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		if req.Account == nil {
			req.Account = settings.Settings.Load().Account.Address
		}

		type SearchResult struct {
			Key            string                             `json:"key"`
			Score          float64                            `json:"score"`
			Listing        *listings.Listing                  `json:"listing"`
			ListingSummary *listings_summaries.ListingSummary `json:"listingSummary"`
		}

		list, err := federation_network.JustAggregateData("find-listings", api_method_find_listings.APIMethodFindListingsRequest{
			req.Account,
			req.Type,
			req.Start,
		})
		if err != nil {
			return nil, err
		}

		duplicates := make(map[string]bool)
		banned := make(map[string]bool)

		count := 0

		for _, searchResult := range list {

			if duplicates[searchResult.Key] || banned[searchResult.Conn.RemoteAddr] { //已经找到 || 已被禁止
				continue
			}

			if err = func() error {

				listing := &listings.Listing{}
				if err = federation_network.AggregateBestData[api_types.APIMethodGetResult]("sync-item", &api_method_sync_item.APIMethodSyncItemRequest{
					sync_type.SYNC_LISTINGS,
					searchResult.Key,
				}, "get-listing", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {
					if len(answer.Result) > 0 {
						if err := listing.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
							return err
						}

						if !listing.FederationIdentity.Equals(f.Federation.Ownership.Address) ||
							listing.Validate() != nil || listing.ValidateSignatures() != nil ||
							!f.Federation.IsValidationAccepted(listing.Validation) {
							return errors.New("listing was not accepted")
						}

						if listing.Identity.Encoded != searchResult.Key {
							return errors.New("listing identity mismatch")
						}
						if !listing.Publisher.Address.Equals(req.Account) {
							return errors.New("invalid search")
						}
						if float64(listing.Publisher.Timestamp) < searchResult.Score {
							return errors.New("invalid score")
						}

					}
					return nil
				}, banned); err != nil {
					return err
				}

				if listing == nil {
					return errors.New("listing not found")
				}

				var listingSummary *listings_summaries.ListingSummary
				if err = federation_network.AggregateBestData[api_types.APIMethodGetResult]("sync-item", &api_method_sync_item.APIMethodSyncItemRequest{
					sync_type.SYNC_LISTINGS_SUMMARIES,
					listing.Identity.Encoded,
				}, "get-listing-summary", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {
					if len(answer.Result) > 0 {
						listingSummary = &listings_summaries.ListingSummary{}
						if err := listingSummary.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
							return err
						}
						if listingSummary.Validate() != nil || listingSummary.ValidateSignatures() != nil || !f.Federation.IsValidationAccepted(listingSummary.Validation) {
							return errors.New("listing summary is invalid")
						}
						if !listingSummary.ListingIdentity.Equals(listing.Identity) {
							return errors.New("listingSummary identity mismatch")
						}
					}
					return nil
				}, banned); err != nil {
					return err
				}

				duplicates[searchResult.Key] = true

				b2, err := webassembly_utils.ConvertJSONBytes(&SearchResult{searchResult.Key, searchResult.Score, listing, listingSummary})
				if err != nil {
					return err
				}

				args[1].Invoke(b2)
				count++

				return nil
			}(); err != nil {
				gui.GUI.Error("banning connection", searchResult.Conn.RemoteAddr, err)
				banned[searchResult.Conn.RemoteAddr] = true
			}

		}

		return count, err
	})
}
