package main

import (
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/polls"
	"liberty-town/node/federations/federation_store/store_data/polls/vote"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/validator/validation"
	"syscall/js"
)

func voteNow(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		var err error

		req := &struct {
			Identity *addresses.Address `json:"identity"`
			Vote     int                `json:"vote"`
		}{}

		if err = json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		newVote := &vote.Vote{
			f.Federation.Ownership.Address,
			req.Identity,
			0,
			0,
			&validation.Validation{},
		}

		var validationExtra any
		if newVote.Validation, validationExtra, err = federationValidate(f.Federation, newVote.GetMessageForSigningValidator, args[1], &api_types.ValidatorCheckExtraRequest{
			0,
			&api_types.ValidatorCheckVoteExtraRequest{
				req.Vote,
				req.Identity,
			},
		}); err != nil {
			return nil, err
		}

		if validationExtra == nil {
			return nil, errors.New("validation should not be empty")
		}

		extra := validationExtra.(*api_types.ValidatorSolutionVoteExtraResult)
		newVote.Upvotes = extra.Upvotes
		newVote.Downvotes = extra.Downvotes

		newPoll := &polls.Poll{
			polls.POLL_VERSION,
			f.Federation.Ownership.Address,
			req.Identity,
			[]*vote.Vote{
				newVote,
			},
		}

		results := 0
		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-vote",
			&api_types.APIMethodStoreIdentityRequest{
				newVote.Identity,
				helpers.SerializeToBytes(newVote),
			}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
				if a != nil && a.Result {
					results++
				}
				return true
			}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Poll    *polls.Poll `json:"poll"`
			Results int         `json:"results"`
		}{newPoll, results})

	})
}
