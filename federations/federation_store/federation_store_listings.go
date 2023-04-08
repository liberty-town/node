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

func storeListingCountry(prefix string, listing *listings.Listing, country uint64, score float64, storeAll bool, remove bool, tx store_db_interface.StoreDBTransactionInterface) (err error) {

	if err = storeSortedSet(prefix+strconv.FormatUint(country, 10), listing.Identity.Encoded, score, remove, tx); err != nil {
		return err
	}

	if storeAll {

		if country != 244 {
			if err = storeSortedSet(prefix+strconv.FormatUint(244, 10), listing.Identity.Encoded, score, remove, tx); err != nil {
				return err
			}
		}

		for code, dict := range AllCountries {
			if dict[country] {
				if err = storeSortedSet(prefix+strconv.FormatUint(code, 10), listing.Identity.Encoded, score, remove, tx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func storeListingScore(listingIdentity string, listing *listings.Listing, remove bool, accountSummary *accounts_summaries.AccountSummary, listingSummary *listings_summaries.ListingSummary, tx store_db_interface.StoreDBTransactionInterface) (err error) {

	if listing == nil {
		data := tx.Get("listings:" + listingIdentity)
		if len(data) == 0 {
			return nil
		}
		listing = &listings.Listing{}
		if err = listing.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			return
		}
	}

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

	if err = storeSortedSet("listings_by_publisher:"+string(listing.Type)+":"+string(cryptography.SHA3([]byte(listing.Publisher.Address.Encoded))), listing.Identity.Encoded, score, remove, tx); err != nil {
		return
	}

	for _, category := range listing.Categories {
		if err = storeSortedSet("listings_by_categories:"+string(listing.Type)+":"+string(cryptography.SHA3([]byte(strconv.FormatUint(category, 10)))), listing.Identity.Encoded, score, remove, tx); err != nil {
			return
		}

		if err = storeListingCountry("listings_by_categories:"+string(listing.Type)+":"+string(cryptography.SHA3([]byte(strconv.FormatUint(category, 10))))+":from:", listing, listing.ShipsFrom, score, true, remove, tx); err != nil {
			return
		}

		for _, countryTo := range listing.ShipsTo {
			if err = storeListingCountry("listings_by_categories:"+string(listing.Type)+":"+string(cryptography.SHA3([]byte(strconv.FormatUint(category, 10))))+":to:", listing, countryTo, score, false, remove, tx); err != nil {
				return err
			}
		}
	}

	words := append(listing.GetWords(), "")
	for _, word := range words {

		if err = storeSortedSet("listings_by_name:"+string(listing.Type)+":"+string(cryptography.SHA3([]byte(word))), listing.Identity.Encoded, score, remove, tx); err != nil {
			return
		}

		if err = storeListingCountry("listings_by_name:"+string(listing.Type)+":"+string(cryptography.SHA3([]byte(word)))+":from:", listing, listing.ShipsFrom, score, true, remove, tx); err != nil {
			return
		}

		for _, countryTo := range listing.ShipsTo {
			if err = storeListingCountry("listings_by_name:"+string(listing.Type)+":"+string(cryptography.SHA3([]byte(word)))+":to:", listing, countryTo, score, false, remove, tx); err != nil {
				return err
			}
		}
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
			if err = storeListingScore(listing.Identity.Encoded, nil, true, nil, nil, tx); err != nil {
				return
			}
		}

		tx.Put("listings:"+listing.Identity.Encoded, helpers.SerializeToBytes(listing))
		tx.Put("listings_publishers:"+listing.Identity.Encoded, []byte(listing.Publisher.Address.Encoded))

		if err = store_utils.IncreaseCount("listings", listing.Identity.Encoded, listing.GetBetterScore(), tx); err != nil {
			return
		}

		if err = storeListingScore(listing.Identity.Encoded, listing, false, nil, nil, tx); err != nil {
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

		if err = storeListingScore(listing.Identity.Encoded, listing, true, nil, nil, tx); err != nil {
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

func SearchListings(queries []string, listingType listing_type.ListingType, queryType byte, shipping uint64, shippingType byte, start int) (list []*api_types.APIMethodFindListItem, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	if len(queries) > 3 {
		return nil, errors.New("too many words in the query")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		var strQuery string
		switch queryType {
		case 0:
			strQuery = "listings_by_name:"
		case 1:
			strQuery = "listings_by_categories:"
		}

		var strShippingQueries []string

		var ss *small_sorted_set.SmallSortedSet

		switch shippingType {
		case 0:
			strShippingQueries = append(strShippingQueries, "")
		case 1:
			strShippingQueries = append(strShippingQueries, ":from:"+strconv.FormatUint(shipping, 10))
		case 2:
			strShippingQueries = append(strShippingQueries, ":to:"+strconv.FormatUint(shipping, 10))
			if shipping != 244 {
				strShippingQueries = append(strShippingQueries, ":to:"+strconv.FormatUint(244, 10))
			}
			for code, dict := range AllCountries {
				if dict[shipping] && code != shipping {
					strShippingQueries = append(strShippingQueries, ":to:"+strconv.FormatUint(code, 10))
				}
			}
		}

		if len(queries) == 1 && queries[0] == "*" {
			queries[0] = ""
		}

		reunion := make(map[string]*small_sorted_set.SmallSortedSetNode)

		for _, strShipping := range strShippingQueries {

			intersection := make(map[string]int)

			ss = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, strQuery+string(listingType)+":"+string(cryptography.SHA3([]byte(queries[0])))+strShipping, tx)
			if err = ss.Read(); err != nil {
				return err
			}
			for _, d := range ss.Data {
				intersection[d.Key] = 1
			}

			for i := 1; i < len(queries); i++ {
				ss = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, strQuery+string(listingType)+":"+string(cryptography.SHA3([]byte(queries[i])))+strShipping, tx)
				if err = ss.Read(); err != nil {
					return err
				}
				for _, d := range ss.Data {
					if intersection[d.Key] > 0 {
						intersection[d.Key]++
					}
				}
			}

			for k, v := range intersection {
				if v == len(queries) {
					reunion[k] = ss.Dict[k]
				}
			}

		}

		var finals []*small_sorted_set.SmallSortedSetNode
		for _, v := range reunion {
			finals = append(finals, v)
		}

		slices.SortFunc(finals, func(a, b *small_sorted_set.SmallSortedSetNode) bool {
			return a.Score > b.Score
		})

		for i := start; i < len(finals) && len(list) < config.LISTINGS_LIST_COUNT; i++ {

			result := finals[i]

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
	return FindData("listings_by_publisher:"+string(listingType)+":"+string(cryptography.SHA3([]byte(account.Encoded))), start, config.LISTINGS_LIST_COUNT)
}
