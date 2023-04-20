//go:build !wasm
// +build !wasm

package federation_store

import (
	"errors"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/accounts_summaries"
	"liberty-town/node/federations/federation_store/store_data/listings/listing_type"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/store/small_sorted_set"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func StoreAccountSummary(accountSummary *accounts_summaries.AccountSummary) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(accountSummary.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if f.Federation.FindModerator(accountSummary.Signer.Address.Encoded) == nil {
		return errors.New("signer is not a federation moderator")
	}

	if err := accountSummary.Validate(); err != nil {
		return err
	}
	if err := accountSummary.ValidateSignatures(); err != nil {
		return err
	}
	if !f.Federation.IsValidationAccepted(accountSummary.Validation) {
		return errors.New("validation signature is not accepted")
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		score, err := store_utils.GetBetterScore("accounts_summaries", accountSummary.AccountIdentity.Encoded, tx)
		if err != nil {
			return err
		}
		if accountSummary.GetBetterScore() < score {
			return errors.New("data is older")
		}

		tx.Put("accounts_summaries:"+accountSummary.AccountIdentity.Encoded, helpers.SerializeToBytes(accountSummary))

		process := func(name string) error {
			ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, name, tx)
			if err = ss.Read(); err != nil {
				return err
			}

			for _, entry := range ss.Data {
				if err = storeListingScore(entry.Key, nil, false, accountSummary, nil, tx); err != nil {
					return err
				}
			}
			return nil
		}

		if err := process("listings_by_publisher:" + string(listing_type.LISTING_BUY) + ":" + string(cryptography.SHA3([]byte(accountSummary.AccountIdentity.Encoded)))); err != nil {
			return err
		}

		if err := process("listings_by_publisher:" + string(listing_type.LISTING_SELL) + ":" + string(cryptography.SHA3([]byte(accountSummary.AccountIdentity.Encoded)))); err != nil {
			return err
		}

		if err = store_utils.IncreaseCount("accounts_summaries", accountSummary.AccountIdentity.Encoded, accountSummary.GetBetterScore(), tx); err != nil {
			return
		}

		return nil
	})
}
