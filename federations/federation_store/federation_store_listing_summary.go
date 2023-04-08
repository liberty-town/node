package federation_store

import (
	"errors"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/listings_summaries"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func StoreListingSummary(listingSummary *listings_summaries.ListingSummary) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(listingSummary.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if f.Federation.FindModerator(listingSummary.Signer.Address.Encoded) == nil {
		return errors.New("signer is not a federation moderator")
	}

	if err := listingSummary.Validate(); err != nil {
		return err
	}
	if err := listingSummary.ValidateSignatures(); err != nil {
		return err
	}
	if !f.Federation.IsValidationAccepted(listingSummary.Validation) {
		return errors.New("validation signature is not accepted")
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		score, err := store_utils.GetBetterScore("accounts_summaries", listingSummary.ListingIdentity.Encoded, tx)
		if err != nil {
			return err
		}
		if listingSummary.GetBetterScore() < score {
			return errors.New("data is older")
		}

		tx.Put("listings_summaries:"+listingSummary.ListingIdentity.Encoded, helpers.SerializeToBytes(listingSummary))

		if err = storeListingScore(listingSummary.ListingIdentity.Encoded, nil, false, nil, listingSummary, tx); err != nil {
			return
		}

		if err = store_utils.IncreaseCount("listings_summaries", listingSummary.ListingIdentity.Encoded, listingSummary.GetBetterScore(), tx); err != nil {
			return
		}

		return nil
	})
}
