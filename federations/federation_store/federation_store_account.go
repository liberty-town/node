package federation_store

import (
	"errors"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/accounts"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func StoreAccount(account *accounts.Account) error {

	f := federation_serve.ServeFederation.Load()
	if f == nil || !f.Federation.Ownership.Address.Equals(account.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if err := account.Validate(); err != nil {
		return err
	}
	if err := account.ValidateSignatures(); err != nil {
		return err
	}
	if !f.Federation.IsValidationAccepted(account.Validation) {
		return errors.New("validation signature is not accepted")
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		score, err := store_utils.GetBetterScore("accounts", account.Identity.Encoded, tx)
		if err != nil {
			return err
		}
		if account.GetBetterScore() < score {
			return errors.New("data is older")
		}

		tx.Put("accounts:"+account.Identity.Encoded, helpers.SerializeToBytes(account))
		if err = store_utils.IncreaseCount("accounts", account.Identity.Encoded, account.GetBetterScore(), tx); err != nil {
			return
		}

		return
	})
}
