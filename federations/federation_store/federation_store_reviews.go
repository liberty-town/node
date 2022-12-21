package federation_store

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/reviews"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/store/small_sorted_set"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func StoreReview(review *reviews.Review) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(review.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if f.Federation.FindModerator(review.Signer.Address.Encoded) == nil {
		return errors.New("signer is not a federation moderator")
	}

	if err := review.Validate(); err != nil {
		return err
	}
	if err := review.ValidateSignatures(); err != nil {
		return err
	}
	if !f.Federation.IsValidationAccepted(review.Validation) {
		return errors.New("validation singuatre is not accepted")
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		score, err := store_utils.GetBetterScore("reviews", review.Identity.Encoded, tx)
		if err != nil {
			return err
		}
		if review.GetBetterScore() < score {
			return errors.New("data is older")
		}

		tx.Put("reviews:"+review.Identity.Encoded, helpers.SerializeToBytes(review))

		if len(review.ListingIdentity.Encoded) > 0 {
			ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "reviews_by_listings:"+string(cryptography.SHA3([]byte(review.ListingIdentity.Encoded))), tx)
			if err = ss.Read(); err != nil {
				return
			}
			ss.Add(review.Identity.Encoded, float64(review.Ownership.Timestamp))
			ss.Save()
		}

		ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "reviews_by_accounts:"+string(cryptography.SHA3([]byte(review.AccountIdentity.Encoded))), tx)
		if err = ss.Read(); err != nil {
			return
		}
		ss.Add(review.Identity.Encoded, float64(review.Ownership.Timestamp))
		ss.Save()

		if err = store_utils.IncreaseCount("reviews", review.Identity.Encoded, review.GetBetterScore(), tx); err != nil {
			return
		}

		return nil
	})
}

func GetReviews(identity *addresses.Address, identityType byte, start int) (list []*api_types.APIMethodFindListItem, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		var str string
		switch identityType {
		case 0:
			str = "reviews_by_accounts:"
		case 1:
			str = "reviews_by_listings:"
		default:
			return errors.New("invalid identity type")
		}

		ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, str+string(cryptography.SHA3([]byte(identity.Encoded))), tx)
		if err := ss.Read(); err != nil {
			return err
		}

		for i := start; i < len(ss.Data) && len(list) < config.REVIEWS_LIST_COUNT; i++ {
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

func GetReview(reviewIdentity string) (review []byte, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {
		review = tx.Get("reviews:" + reviewIdentity)
		return nil
	})
	return
}
