package main

import (
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/exp/slices"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/config"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_network/sync_type"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/ownership"
	"liberty-town/node/federations/federation_store/store_data/polls"
	"liberty-town/node/federations/federation_store/store_data/polls/vote"
	"liberty-town/node/federations/federation_store/store_data/threads"
	"liberty-town/node/federations/federation_store/store_data/threads/thread_type"
	"liberty-town/node/network/api_implementation/api_common/api_method_search_threads"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_item"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/pandora-pay/helpers/msgpack"
	"liberty-town/node/settings"
	"sync/atomic"
	"syscall/js"
)

func threadStore(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		var err error

		req := &struct {
			Type     thread_type.ThreadType `json:"type"`
			Title    string                 `json:"title"`
			Keywords []string               `json:"keywords"`
			Content  string                 `json:"content"`
			Links    []string               `json:"links"`
		}{}

		if err = json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		account := settings.Settings.Load().Account

		it := &threads.Thread{
			Version:            threads.THREAD_VERSION,
			FederationIdentity: f.Federation.Ownership.Address,
			Type:               req.Type,
			Title:              req.Title,
			Keywords:           req.Keywords,
			Content:            req.Content,
			Links:              req.Links,
			Publisher:          &ownership.Ownership{},
		}

		if err = it.Validate(); err != nil && err.Error() != "listing ownership identity does not match" {
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
		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-thread", &api_types.APIMethodStoreRequest{helpers.SerializeToBytes(it)}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
			if a != nil && a.Result {
				results.Add(1)
			}
			return true
		}, &generics.Map[string, bool]{}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			Thread  *threads.Thread `json:"thread"`
			Results int32           `json:"results"`
		}{it, results.Load()})

	})
}

func threadGet(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Thread string `json:"thread,omitempty"`
		}{}

		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		var thread *threads.Thread

		if err := federation_network.FetchData[api_types.APIMethodGetResult]("get-thread", &api_types.APIMethodGetRequest{
			req.Thread,
		}, func(data *api_types.APIMethodGetResult, contact *connection.AdvancedConnection) bool {
			if data == nil || data.Result == nil {
				return true
			}
			temp := &threads.Thread{
				FederationIdentity: f.Federation.Ownership.Address,
			}
			if err := temp.Deserialize(advanced_buffers.NewBufferReader(data.Result)); err != nil {
				return true
			}
			if !temp.FederationIdentity.Equals(f.Federation.Ownership.Address) {
				return true
			}
			if temp.Validate() != nil || temp.ValidateSignatures() != nil {
				return true
			}
			thread = temp
			return false
		}, &generics.Map[string, bool]{}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(thread)

	})
}

func threadsSearch(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Query     []string `json:"query,omitempty"`
			QueryType byte     `json:"queryType,omitempty"`
			Start     int      `json:"start"`
		}{}

		if err := json.Unmarshal([]byte(args[0].String()), req); err != nil {
			return nil, err
		}

		type SearchResult struct {
			Key    string          `json:"key"`
			Score  float64         `json:"score"`
			Thread *threads.Thread `json:"thread"`
			Poll   *polls.Poll     `json:"poll"`
		}

		count := &atomic.Int32{}

		banned := &generics.Map[string, bool]{}

		if err := federation_network.AggregateListAndCustomProcess("search-threads", api_method_search_threads.APIMethodSearchThreadsRequest{
			req.Query,
			req.QueryType,
			req.Start,
		}, func(searchResult *federation_network.AggregationListResult) (err error) {

			b, err := msgpack.Marshal(&api_types.APIMethodGetRequest{searchResult.Key})
			if err != nil {
				return err
			}

			out := searchResult.Conn.SendAwaitAnswer([]byte("get-thread"), b, context.Background(), 0)
			if out.Err != nil {
				return out.Err
			}

			answer := new(api_types.APIMethodGetResult)
			if err = msgpack.Unmarshal(out.Out, answer); err != nil {
				return err
			}

			if len(answer.Result) == 0 {
				return errors.New("no result")
			}

			var poll *polls.Poll

			thread := &threads.Thread{
				FederationIdentity: f.Federation.Ownership.Address,
			}
			if err = thread.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
				return err
			}

			if !thread.FederationIdentity.Equals(f.Federation.Ownership.Address) ||
				thread.Validate() != nil || thread.ValidateSignatures() != nil ||
				!f.Federation.IsValidationAccepted(thread.Validation) {
				return errors.New("thread was not accepted")
			}

			if thread.Identity.Encoded != searchResult.Key {
				return errors.New("listing identity mismatch")
			}

			//pass banned
			if err = federation_network.AggregateBestResult[api_types.APIMethodGetResult]("sync-item", &api_method_sync_item.APIMethodSyncItemRequest{
				sync_type.SYNC_POLLS,
				thread.Identity.Encoded,
			}, "get-poll", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {
				if len(answer.Result) > 0 {
					addr, err := addresses.DecodeAddr(key)
					if err != nil {
						return err
					}

					poll = &polls.Poll{
						FederationIdentity: f.Federation.Ownership.Address,
						Identity:           addr,
					}
					if err = poll.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
						return err
					}
					if poll.Validate() != nil || poll.ValidateSignatures() != nil {
						return errors.New("poll is invalid")
					}
					for i := range poll.List {
						if !f.Federation.IsValidationAccepted(poll.List[i].Validation) {
							return errors.New("poll validation is not accepted")
						}
					}
					if !poll.Identity.Equals(thread.Identity) {
						return errors.New("poll identity mismatch")
					}
				}
				return nil
			}, banned); err != nil {
				return err
			}

			pollScore := poll.GetScore()
			foundScore := thread.GetScore(pollScore)

			if searchResult.Score > foundScore {
				return errors.New("score is less than it should be")
			}

			var searchData []string
			switch req.QueryType {
			case 0:
				searchData = thread.GetWords()
			case 1:
				searchData = slices.Clone(thread.Keywords)
			}

			for _, query := range req.Query {
				if query != "" {
					found := false
					for _, c := range searchData {
						if c == query {
							found = true
							break
						}
					}
					if !found {
						return errors.New("query not found")
					}
				}
			}

			if poll == nil {
				poll = &polls.Poll{
					polls.POLL_VERSION,
					thread.FederationIdentity,
					thread.Identity,
					[]*vote.Vote{
						{
							thread.FederationIdentity,
							thread.Identity,
							0, 0, nil,
						},
					},
				}
			}

			b2, err := webassembly_utils.ConvertJSONBytes(&SearchResult{searchResult.Key, foundScore, thread, poll})
			if err != nil {
				return err
			}

			args[1].Invoke(b2)
			count.Add(1)

			return
		}, config.THREADS_LIST_COUNT, banned); err != nil {
			return nil, err
		}

		return count.Load(), nil
	})
}
