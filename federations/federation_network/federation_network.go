package federation_network

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"liberty-town/node/config"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_item"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/connected_nodes"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers/generics"
	"liberty-town/node/pandora-pay/helpers/msgpack"
	"sync"
	"sync/atomic"
)

type AggregationListResult struct {
	Key   string
	Score float64
	Conn  *connection.AdvancedConnection
}

func JustAggregateData(methodFindList string, req any, banned *generics.Map[string, bool]) ([]*AggregationListResult, error) {

	list := make([]*AggregationListResult, 0)
	lock := &sync.Mutex{}

	if err := FetchData[api_types.APIMethodFindListResult](methodFindList, req, func(data *api_types.APIMethodFindListResult, conn *connection.AdvancedConnection) bool {

		if data == nil || len(data.Results) == 0 {
			return true
		}

		for _, v := range data.Results {
			aggregated := &AggregationListResult{
				v.Key,
				v.Score,
				conn,
			}

			lock.Lock()
			list = append(list, aggregated)
			lock.Unlock()
		}

		return true
	}, banned); err != nil {
		return nil, err
	}

	slices.SortFunc(list, func(a, b *AggregationListResult) bool {
		return a.Score > b.Score
	})

	return list, nil
}

func AggregateDataFromList[T any](list []*AggregationListResult, methodFindItem string, getRequest func(data *AggregationListResult) (any, error), validateItem func(*T, string, float64) error, count int32, banned *generics.Map[string, bool]) error {

	return ProcessDataFromList(list, count, func(searchItem *AggregationListResult) error {

		var request any
		var err error
		if getRequest != nil {
			if request, err = getRequest(searchItem); err != nil {
				return err
			}
		} else {
			request = &api_types.APIMethodGetRequest{
				searchItem.Key,
			}
		}

		answer, err := connection.SendJSONAwaitAnswer[T](searchItem.Conn, []byte(methodFindItem), request, nil, 0)
		if err != nil {
			return nil
		}

		if err = validateItem(answer, searchItem.Key, searchItem.Score); err != nil {
			return err
		}

		return nil
	}, banned)

}

func ProcessDataFromList(list []*AggregationListResult, count int32, process func(searchItem *AggregationListResult) error, banned *generics.Map[string, bool]) error {

	duplicates := &generics.Map[string, bool]{}

	jobs := make(chan *AggregationListResult, len(list))
	results := make(chan bool, len(list))
	counter := &atomic.Int32{}
	counter.Store(0)

	worker := func() {

		for listItem := range jobs {

			func() {

				if counter.Load() >= count {
					return
				}
				if exists, _ := duplicates.Load(listItem.Key); exists {
					return
				}
				if exists, _ := banned.Load(listItem.Conn.RemoteAddr); exists {
					return
				}

				if err := process(listItem); err == nil {
					counter.Add(1)
					duplicates.Store(listItem.Key, true)
				}

			}()

			results <- true
		}

	}

	for w := 0; w < config.CONCURENCY; w++ {
		go worker()
	}

	for _, conn := range list {
		jobs <- conn
	}

	close(jobs)

	for range list {
		<-results
	}

	return nil
}

func AggregateListData[T any](methodFindList string, req any, methodFindItem string, getRequest func(data *AggregationListResult) (any, error), validateItem func(*T, string, float64) error, count int32, banned *generics.Map[string, bool]) error {
	list, err := JustAggregateData(methodFindList, req, banned)
	if err != nil {
		return err
	}

	return AggregateDataFromList[T](list, methodFindItem, getRequest, validateItem, count, banned)
}

func AggregateListAndCustomProcess(methodFindList string, req any, validateItem func(searchItem *AggregationListResult) error, count int32, banned *generics.Map[string, bool]) error {
	list, err := JustAggregateData(methodFindList, req, banned)
	if err != nil {
		return err
	}

	return ProcessDataFromList(list, count, validateItem, banned)
}

func AggregateBestResult[T any](methodFindList string, req *api_method_sync_item.APIMethodSyncItemRequest, methodFindItem string, getRequest func(data *AggregationListResult) (any, error), validateItem func(*T, string, float64) error, banned *generics.Map[string, bool]) error {

	list := make([]*AggregationListResult, 0)
	lock := &sync.Mutex{}

	if err := FetchData[api_method_sync_item.APIMethodSyncItemResult](methodFindList, req, func(data *api_method_sync_item.APIMethodSyncItemResult, conn *connection.AdvancedConnection) bool {

		if data == nil {
			return true
		}

		aggregated := &AggregationListResult{
			req.Key,
			float64(data.BetterScore),
			conn,
		}

		lock.Lock()
		list = append(list, aggregated)
		lock.Unlock()

		return true
	}, banned); err != nil {
		return err
	}

	slices.SortFunc(list, func(a, b *AggregationListResult) bool {
		return a.Score > b.Score
	})

	return AggregateDataFromList[T](list, methodFindItem, getRequest, validateItem, 1, banned)
}

func FetchData[T any](method string, data any, next func(*T, *connection.AdvancedConnection) bool, banned *generics.Map[string, bool]) error {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return errors.New("no federation")
	}

	b, err := msgpack.Marshal(data)
	if err != nil {
		return nil
	}

	for {

		<-websocks.Websockets.ReadyCn.Load()

		list := connected_nodes.ConnectedNodes.AllList.Get()
		if int64(len(list)) < network_config.NETWORK_CONNECTIONS_READY_THRESHOLD {
			continue
		}

		jobs := make(chan *connection.AdvancedConnection, len(list))
		results := make(chan bool, len(list))
		done := &generics.Value[bool]{}
		done.Store(false)

		worker := func() {
			for conn := range jobs {
				func() {

					if done.Load() {
						return
					}

					if exists, _ := banned.Load(conn.RemoteAddr); exists {
						return
					}

					out := conn.SendAwaitAnswer([]byte(method), b, context.Background(), 0)
					if out.Err != nil {
						fmt.Println("error", out.Err)
						banned.Store(conn.RemoteAddr, true)
						return
					}

					final := new(T)
					if err = msgpack.Unmarshal(out.Out, final); err != nil {
						fmt.Println("error2", out.Err)
						banned.Store(conn.RemoteAddr, true)
						return
					}

					if next != nil && !next(final, conn) {
						done.Store(true)
					}
				}()
				results <- true
			}
		}

		for w := 0; w < config.CONCURENCY; w++ {
			go worker()
		}

		for _, conn := range list {
			jobs <- conn
		}

		close(jobs)

		for range list {
			<-results
		}

		break
	}

	return nil
}
