//go:build !wasm
// +build !wasm

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
		return errors.New("validation signature is not accepted")
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
			if err = storeSortedSet("reviews_by_listings:"+string(cryptography.SHA3([]byte(review.ListingIdentity.Encoded))), review.Identity.Encoded, float64(review.Ownership.Timestamp), false, tx); err != nil {
				return err
			}
		}

		if err = storeSortedSet("reviews_by_accounts:"+string(cryptography.SHA3([]byte(review.AccountIdentity.Encoded))), review.Identity.Encoded, float64(review.Ownership.Timestamp), false, tx); err != nil {
			return err
		}

		if err = store_utils.IncreaseCount("reviews", review.Identity.Encoded, review.GetBetterScore(), tx); err != nil {
			return
		}

		return nil
	})
}

func GetReviews(identity *addresses.Address, identityType byte, start int) (list []*api_types.APIMethodFindListItem, err error) {
	switch identityType {
	case 0:
		return FindData("reviews_by_accounts:"+string(cryptography.SHA3([]byte(identity.Encoded))), start, config.REVIEWS_LIST_COUNT)
	case 1:
		return FindData("reviews_by_listings:"+string(cryptography.SHA3([]byte(identity.Encoded))), start, config.REVIEWS_LIST_COUNT)
	default:
		return nil, errors.New("invalid identity type")
	}
}
