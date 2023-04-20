//go:build !wasm
// +build !wasm

package federation_store

import (
	"errors"
	"golang.org/x/exp/slices"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/federations/federation_store/store_data/polls"
	"liberty-town/node/federations/federation_store/store_data/threads"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/store/small_sorted_set"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func storeThreadScore(f *federation.Federation, tx store_db_interface.StoreDBTransactionInterface, threadIdentity *addresses.Address, thread *threads.Thread, remove bool, threadPoll *polls.Poll) (err error) {

	if thread == nil {
		data := tx.Get("threads:" + threadIdentity.Encoded)
		if len(data) == 0 {
			return nil
		}
		thread = &threads.Thread{
			FederationIdentity: f.Ownership.Address,
		}
		if err = thread.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
			return err
		}
	}

	score := float64(0)

	if !remove {

		if threadPoll == nil {
			if data := tx.Get("polls:" + threadIdentity.Encoded); data != nil {
				threadPoll = &polls.Poll{
					FederationIdentity: f.Ownership.Address,
					Identity:           threadIdentity,
				}
				if err = threadPoll.Deserialize(advanced_buffers.NewBufferReader(data)); err != nil {
					return
				}
			}
		}

		v := threadPoll.GetScore()
		score = thread.GetScore(v)
	}

	if err = storeSortedSet("threads_by_publisher:"+string(cryptography.SHA3([]byte(thread.Publisher.Address.Encoded))), thread.Identity.Encoded, score, remove, tx); err != nil {
		return
	}

	for _, keyword := range thread.Keywords {
		if err = storeSortedSet("threads_by_keyword:"+string(cryptography.SHA3([]byte(keyword))), thread.Identity.Encoded, score, remove, tx); err != nil {
			return
		}
	}

	words := thread.GetWords()
	for _, word := range words {
		if err = storeSortedSet("threads_by_title:"+string(cryptography.SHA3([]byte(word))), thread.Identity.Encoded, score, remove, tx); err != nil {
			return
		}
	}

	if err = storeSortedSet("threads_all", thread.Identity.Encoded, score, remove, tx); err != nil {
		return
	}

	return nil
}

func StoreThread(thread *threads.Thread) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(thread.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if err := thread.Validate(); err != nil {
		return err
	}
	if err := thread.ValidateSignatures(); err != nil {
		return err
	}
	if !f.Federation.IsValidationAccepted(thread.Validation) {
		return errors.New("validation signature is not accepted")
	}

	return f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		if x := tx.Get("threads:" + thread.Identity.Encoded); x != nil {
			return errors.New("thread already exists")
		}

		tx.Put("threads:"+thread.Identity.Encoded, helpers.SerializeToBytes(thread))
		tx.Put("threads_publishers:"+thread.Identity.Encoded, []byte(thread.Publisher.Address.Encoded))

		if err = store_utils.IncreaseCount("threads", thread.Identity.Encoded, 0, tx); err != nil {
			return
		}

		if err = storeThreadScore(f.Federation, tx, thread.Identity, thread, false, nil); err != nil {
			return
		}

		return nil
	})
}

func SearchThreads(queries []string, queryType byte, start int) (list []*api_types.APIMethodFindListItem, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	if len(queries) > 3 {
		return nil, errors.New("too many words in the query")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		intersection := make(map[string]int)

		var str string
		var ss *small_sorted_set.SmallSortedSet

		if len(queries) == 1 && (queries[0] == "" || queries[0] == "*") {
			ss = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "threads_all", tx)
		} else {
			switch queryType {
			case 0:
				str = "threads_by_title:"
			case 1:
				str = "threads_by_keyword:"
			}
			ss = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, str+string(cryptography.SHA3([]byte(queries[0]))), tx)
		}

		if err = ss.Read(); err != nil {
			return err
		}
		for _, d := range ss.Data {
			intersection[d.Key] = 1
		}

		for i := 1; i < len(queries); i++ {
			ss = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, str+string(cryptography.SHA3([]byte(queries[i]))), tx)
			if err = ss.Read(); err != nil {
				return err
			}
			for _, d := range ss.Data {
				if intersection[d.Key] > 0 {
					intersection[d.Key]++
				}
			}
		}

		var finals []*small_sorted_set.SmallSortedSetNode
		for key, val := range intersection {
			if val == len(queries) {
				finals = append(finals, ss.Dict[key])
			}
		}

		slices.SortFunc(finals, func(a, b *small_sorted_set.SmallSortedSetNode) bool {
			return a.Score > b.Score
		})

		for i := start; i < len(ss.Data) && len(list) < config.THREADS_LIST_COUNT; i++ {

			result := ss.Data[i]

			list = append(list, &api_types.APIMethodFindListItem{
				result.Key,
				result.Score,
			})
		}

		return nil
	})
	return
}
