package federation_store

import (
	"errors"
	"golang.org/x/exp/slices"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/accounts_summaries"
	"liberty-town/node/federations/federation_store/store_data/listings"
	"liberty-town/node/federations/federation_store/store_data/listings/listing_type"
	"liberty-town/node/federations/federation_store/store_data/listings_summaries"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/store/small_sorted_set"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
	"strconv"
)

func storeListingScore(tx store_db_interface.StoreDBTransactionInterface, listing *listings.Listing, remove bool, accountSummary *accounts_summaries.AccountSummary, listingSummary *listings_summaries.ListingSummary) (err error) {

	var score float64

	if !remove {

		if accountSummary == nil {
			if data := tx.Get("accounts_summaries:" + listing.Publisher.Address.Encoded); data != nil {
				accountSummary = &accounts_summaries.AccountSummary{}
				if err = accountSummary.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
					return
				}
			}
		}
		if listingSummary == nil {
			if data := tx.Get("listings_summaries:" + listing.Ownership.Address.Encoded); data != nil {
				listingSummary = &listings_summaries.ListingSummary{}
				if err = listingSummary.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
					return
				}
			}
		}

		accountSummaryScore := accountSummary.GetScore(listing.Type)
		listingSummaryScore := listingSummary.GetScore()
		score = listings.GetScore(listingSummaryScore, accountSummaryScore)
	}

	process := func(name string) (err error) {
		ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, name, tx)
		if err = ss.Read(); err != nil {
			return
		}
		if len(ss.Data) >= config.LIST_SIZE-1 {
			return errors.New("publisher has way to many listings")
		}
		if remove {
			ss.Delete(listing.Identity.Encoded)
		} else {
			ss.Add(listing.Identity.Encoded, score)
		}
		ss.Save()
		return
	}

	if err = process("listings_by_publisher:" + string(listing.Type) + ":" + string(cryptography.SHA3([]byte(listing.Publisher.Address.Encoded)))); err != nil {
		return
	}

	for _, category := range listing.Categories {
		if err = process("listings_by_categories:" + string(listing.Type) + ":" + string(cryptography.SHA3([]byte(strconv.FormatUint(category, 10))))); err != nil {
			return
		}
	}

	words := listing.GetWords()
	for _, word := range words {
		if err = process("listings_by_name:" + string(listing.Type) + ":" + string(cryptography.SHA3([]byte(word)))); err != nil {
			return
		}
	}

	if err = process("listings_all:" + string(listing.Type)); err != nil {
		return
	}

	return nil
}

func StoreListing(listing *listings.Listing) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(listing.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if err := listing.Validate(); err != nil {
		return err
	}
	if err := listing.ValidateSignatures(); err != nil {
		return err
	}
	if !f.Federation.IsValidationAccepted(listing.Validation) {
		return errors.New("validation signature is not accepted")
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		score, err := store_utils.GetBetterScore("listings", listing.Identity.Encoded, tx)
		if err != nil {
			return err
		}
		if listing.GetBetterScore() < score {
			return errors.New("data is older")
		}
		if score > 0 {
			data := tx.Get("listings:" + listing.Ownership.Address.Encoded)
			old := &listings.Listing{}
			if err = old.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
				return
			}
			if err = storeListingScore(tx, old, true, nil, nil); err != nil {
				return
			}
		}

		tx.Put("listings:"+listing.Identity.Encoded, helpers.SerializeToBytes(listing))
		tx.Put("listings_publishers:"+listing.Identity.Encoded, []byte(listing.Publisher.Address.Encoded))

		if err = store_utils.IncreaseCount("listings", listing.Identity.Encoded, listing.GetBetterScore(), tx); err != nil {
			return
		}

		if err = storeListingScore(tx, listing, false, nil, nil); err != nil {
			return
		}

		return nil
	})
}

func RemoveListing(federationIdentity, listingIdentity string) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || f.Federation.Ownership.Address.Encoded != federationIdentity {
		return errors.New("not serving this federation")
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {
		data := tx.Get("listings:" + listingIdentity)
		if data == nil {
			return
		}
		listing := &listings.Listing{}
		if err = listing.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			return
		}

		tx.Delete("listings:" + listingIdentity)
		tx.Delete("listings_publishers:" + listingIdentity)
		tx.Delete("listings_summaries:" + listingIdentity)

		if err = storeListingScore(tx, listing, true, nil, nil); err != nil {
			return
		}

		if err = store_utils.DecreaseCount("listings", listingIdentity, tx); err != nil {
			return
		}
		if err = store_utils.DecreaseCount("listings_summaries", listingIdentity, tx); err != nil {
			return
		}

		return
	})
}

func GetListing(listingIdentity string) (listing []byte, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {
		listing = tx.Get("listings:" + listingIdentity)
		return nil
	})
	return
}

func GetListingData(listingIdentity string) (listing, accountSummary, listingSummary []byte, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, nil, nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		listing = tx.Get("listings:" + listingIdentity)
		if len(listing) > 0 {
			listingSummary = tx.Get("listings_summaries:" + listingIdentity)

			listingPublisher := tx.Get("listings_publishers:" + listingIdentity)
			accountSummary = tx.Get("accounts_summaries:" + string(listingPublisher))
		}
		return nil
	})
	return
}

func SearchListings(queries []string, listingType listing_type.ListingType, queryType byte, start int) (list []*api_types.APIMethodFindListItem, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	if len(queries) > 3 {
		return nil, errors.New("too many words in the query")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		intersection := make(map[string]int)

		var str string
		var ss *small_sorted_set.SmallSortedSet

		if len(queries) == 1 && (queries[0] == "" || queries[0] == "*") {
			ss = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "listings_all:"+string(listingType), tx)
		} else {
			switch queryType {
			case 0:
				str = "listings_by_name:"
			case 1:
				str = "listings_by_categories:"
			}
			ss = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, str+string(listingType)+":"+string(cryptography.SHA3([]byte(queries[0]))), tx)
		}

		if err = ss.Read(); err != nil {
			return err
		}
		for _, d := range ss.Data {
			intersection[d.Key] = 1
		}

		for i := 1; i < len(queries); i++ {
			ss = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, str+string(cryptography.SHA3([]byte(queries[i]))), tx)
			if err = ss.Read(); err != nil {
				return err
			}
			for _, d := range ss.Data {
				if intersection[d.Key] > 0 {
					intersection[d.Key]++
				}
			}
		}

		var finals []*small_sorted_set.SmallSortedSetNode
		for key, val := range intersection {
			if val == len(queries) {
				finals = append(finals, ss.Dict[key])
			}
		}

		slices.SortFunc(finals, func(a, b *small_sorted_set.SmallSortedSetNode) bool {
			return a.Score > b.Score
		})

		for i := start; i < len(ss.Data) && len(list) < config.LISTINGS_LIST_COUNT; i++ {

			result := ss.Data[i]

			list = append(list, &api_types.APIMethodFindListItem{
				result.Key,
				result.Score,
			})
		}

		return nil
	})
	return
}

func GetListings(account *addresses.Address, listingType listing_type.ListingType, start int) (list []*api_types.APIMethodFindListItem, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "listings_by_publisher:"+string(listingType)+":"+string(cryptography.SHA3([]byte(account.Encoded))), tx)
		if err = ss.Read(); err != nil {
			return err
		}

		for i := start; i < len(ss.Data) && len(list) < config.LISTINGS_LIST_COUNT; i++ {
			result := ss.Data[i]
			list = append(list, &api_types.APIMethodFindListItem{
				result.Key,
				result.Score,
			})
		}

		return nil
	})
	return
}
