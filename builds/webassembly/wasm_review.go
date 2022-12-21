package main

import (
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/reviews"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_reviews"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/settings"
	"syscall/js"
)

func reviewStore(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		var err error

		req := &struct {
			ListingIdentity *addresses.Address `json:"listingIdentity,omitempty"`
			AccountIdentity *addresses.Address `json:"accountIdentity,omitempty"`
			Text            string             `json:"text"`
			Score           byte               `json:"score"`
			Amount          uint64             `json:"amount"`
			Nonce           []byte             `json:"nonce,omitempty"`
		}{}

		if err = json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		if len(req.Nonce) == 0 {
			req.Nonce = cryptography.RandomHash()
		}

		account := settings.Settings.Load().Account

		key := cryptography.SHA3(cryptography.SHA3(append(account.PrivateKey.Key[:], req.Nonce...)))

		reviewPrivateKey, err := addresses.NewPrivateKey(key)
		if err != nil {
			return nil, err
		}

		reviewAddress, err := reviewPrivateKey.GenerateAddress()
		if err != nil {
			return nil, err
		}

		it := &reviews.Review{
			reviews.REVIEW_VERSION,
			f.Federation.Ownership.Address,
			req.Nonce,
			reviewAddress,
			req.ListingIdentity,
			req.AccountIdentity,
			req.Text,
			req.Score,
			req.Amount,
			nil,
			&ownership.Ownership{},
			&ownership.Ownership{},
		}

		if it.Validation, err = federationValidate(f.Federation, it.GetMessageForSigningValidator, args[1]); err != nil {
			return nil, err
		}

		if err = it.Ownership.Sign(reviewPrivateKey, it.GetMessageForSigningOwnership); err != nil {
			return nil, err
		}

		if err = it.Signer.Sign(settings.Settings.Load().Account.PrivateKey, it.GetMessageForSigningSigner); err != nil {
			return nil, err
		}

		if err = it.Validate(); err != nil {
			return nil, err
		}

		results := 0
		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-review", api_types.APIMethodStoreRequest{helpers.SerializeToBytes(it)}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
			if a != nil && a.Result {
				results++
			}
			return true
		}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Review  *reviews.Review `json:"review"`
			Results int             `json:"results"`
		}{it, results})

	})
}

func reviewsGetAll(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Identity *addresses.Address `json:"identity,omitempty"`
			Type     byte               `json:"type,omitempty"`
			Start    int                `json:"start"`
		}{}
		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		if req.Identity == nil && req.Type == 0 {
			req.Identity = settings.Settings.Load().Account.Address
		}

		type SearchResult struct {
			Key    string          `json:"key"`
			Score  float64         `json:"score"`
			Review *reviews.Review `json:"review"`
		}

		count := 0
		err := federation_network.AggregateData[api_types.APIMethodGetResult]("find-reviews", &api_method_find_reviews.APIMethodFindReviewsRequest{
			req.Identity,
			req.Type,
			req.Start,
		}, "get-review", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {

			review := &reviews.Review{}
			if err := review.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
				return err
			}
			if review.Validate() != nil || review.ValidateSignatures() != nil || !f.Federation.IsValidationAccepted(review.Validation) {
				return errors.New("invalid review")
			}

			if review.Identity.Encoded != key {
				return errors.New("invalid review identity")
			}

			if score > float64(review.Signer.Timestamp) {
				return errors.New("invalid review score")
			}

			switch req.Type {
			case 0:
				if !review.AccountIdentity.Equals(req.Identity) {
					return errors.New("invalid review account identity")
				}
			case 1:
				if !review.ListingIdentity.Equals(req.Identity) {
					return errors.New("invalid review listing identity")
				}
			}

			result := &SearchResult{key, score, review}
			b, err := webassembly_utils.ConvertJSONBytes(result)
			if err != nil {
				return err
			}

			args[1].Invoke(b)

			count++
			return nil
		})

		return count, err
	})
}
