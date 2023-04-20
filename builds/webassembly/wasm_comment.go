package main

import (
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/config"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/comments"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_comments"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/settings"
	"sync/atomic"
	"syscall/js"
)

func commentStore(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		var err error

		req := &struct {
			Thread  *addresses.Address `json:"thread"`
			Content string             `json:"content"`
		}{}

		if err = json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		account := settings.Settings.Load().Account

		it := &comments.Comment{
			Version:            comments.COMMENT_VERSION,
			FederationIdentity: f.Federation.Ownership.Address,
			ParentIdentity:     req.Thread,
			Content:            req.Content,
			Publisher:          &ownership.Ownership{},
		}

		if err = it.Validate(); err != nil && err.Error() != "comment ownership identity does not match" {
			return nil, err
		}

		if it.Validation, _, err = federationValidate(f.Federation, it.GetMessageForSigningValidator, args[1], nil); err != nil {
			return nil, err
		}

		if err = it.Publisher.Sign(account.PrivateKey, it.GetMessageForSigningPublisher); err != nil {
			return nil, err
		}

		if err = it.SetIdentityNow(); err != nil {
			return nil, err
		}

		if err = it.Validate(); err != nil {
			return nil, err
		}

		results := &atomic.Int32{}
		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-comment",
			&api_types.APIMethodStoreIdentityRequest{
				req.Thread,
				helpers.SerializeToBytes(it),
			}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
				if a != nil && a.Result {
					results.Add(1)
				}
				return true
			}, &generics.Map[string, bool]{}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Comment *comments.Comment `json:"comment"`
			Results int32             `json:"results"`
		}{it, results.Load()})

	})
}

func commentsGetAll(this js.Value, args []js.Value) any {
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
			Key     string            `json:"key"`
			Score   float64           `json:"score"`
			Comment *comments.Comment `json:"comment"`
		}

		count := 0
		err := federation_network.AggregateListData[api_types.APIMethodGetResult]("find-comments", &api_method_find_comments.APIMethodFindCommentsRequest{
			req.Identity,
			req.Type,
			req.Start,
		}, "get-comment", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {

			comment := &comments.Comment{
				FederationIdentity: f.Federation.Ownership.Address,
				ParentIdentity:     req.Identity,
			}
			if err := comment.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
				return err
			}
			if comment.Validate() != nil || comment.ValidateSignatures() != nil || !f.Federation.IsValidationAccepted(comment.Validation) {
				return errors.New("invalid comment")
			}

			if comment.Identity.Encoded != key {
				return errors.New("invalid comment identity")
			}

			if score != -float64(comment.Validation.Timestamp) {
				return errors.New("invalid comment score")
			}

			switch req.Type {
			case 0:
				if !comment.ParentIdentity.Equals(req.Identity) {
					return errors.New("invalid comment thread identity")
				}
			case 1:
				if !comment.Publisher.Address.Equals(req.Identity) {
					return errors.New("invalid comment publisher identity")
				}
			}

			result := &SearchResult{key, score, comment}
			b, err := webassembly_utils.ConvertJSONBytes(result)
			if err != nil {
				return err
			}

			args[1].Invoke(b)

			count++
			return nil
		}, config.COMMENTS_LIST_COUNT, &generics.Map[string, bool]{})

		return count, err
	})
}
