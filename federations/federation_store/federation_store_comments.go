//go:build !wasm
// +build !wasm

package federation_store

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/comments"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func StoreComments(comment *comments.Comment) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(comment.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if err := comment.Validate(); err != nil {
		return err
	}
	if err := comment.ValidateSignatures(); err != nil {
		return err
	}
	if !f.Federation.IsValidationAccepted(comment.Validation) {
		return errors.New("validation signature is not accepted")
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		if x := tx.Get("comments:" + comment.Identity.Encoded); x != nil {
			return errors.New("comment already exists")
		}

		tx.Put("comments:"+comment.Identity.Encoded, helpers.SerializeToBytes(comment))

		if err := storeSortedSet("comments_by_threads:"+string(cryptography.SHA3([]byte(comment.ParentIdentity.Encoded))), comment.Identity.Encoded, -float64(comment.Validation.Timestamp), false, tx); err != nil {
			return err
		}

		if err := storeSortedSet("comments_by_accounts:"+string(cryptography.SHA3([]byte(comment.Publisher.Address.Encoded))), comment.Identity.Encoded, -float64(comment.Validation.Timestamp), false, tx); err != nil {
			return err
		}

		if err = store_utils.IncreaseCount("comments", comment.Identity.Encoded, 0, tx); err != nil {
			return
		}

		return nil
	})
}

func GetComments(identity *addresses.Address, identityType byte, start int) ([]*api_types.APIMethodFindListItem, error) {
	switch identityType {
	case 0:
		return FindData("comments_by_threads:"+string(cryptography.SHA3([]byte(identity.Encoded))), start, config.COMMENTS_LIST_COUNT)
	case 1:
		return FindData("comments_by_accounts:"+string(cryptography.SHA3([]byte(identity.Encoded))), start, config.COMMENTS_LIST_COUNT)
	default:
		return nil, errors.New("invalid identity type")
	}
}
