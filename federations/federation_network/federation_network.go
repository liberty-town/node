package federation_network

import (
	"context"
	"errors"
	"golang.org/x/exp/slices"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/gui"
	"liberty-town/node/network/api_implementation/api_common/api_method_sync_item"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/connected_nodes"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers/msgpack"
	"sync"
)

type AggregationListResult struct {
	Key   string
	Score float64
	Conn  *connection.AdvancedConnection
}

func JustAggregateData(methodFindList string, req any) ([]*AggregationListResult, error) {
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
	}); err != nil {
		return nil, err
	}

	slices.SortFunc(list, func(a, b *AggregationListResult) bool {
		return a.Score > b.Score
	})

	return list, nil
}

func AggregateData[T any](list []*AggregationListResult, methodFindItem string, getRequest func(data *AggregationListResult) (any, error), validateItem func(*T, string, float64) error, returnFirst bool, banned map[string]bool) error {

	duplicates := make(map[string]bool)

	if banned == nil {
		banned = make(map[string]bool)
	}

	for _, it := range list {

		k := it.Key

		if duplicates[k] || banned[it.Conn.RemoteAddr] { //已经找到 || 已被禁止
			continue
		}

		var request any
		var err error
		if getRequest != nil {
			if request, err = getRequest(it); err != nil {
				banned[it.Conn.RemoteAddr] = true
				continue
			}
		} else {
			request = &api_types.APIMethodGetRequest{
				it.Key,
			}
		}

		answer, err := connection.SendJSONAwaitAnswer[T](it.Conn, []byte(methodFindItem), request, nil, 0)
		if err != nil {
			continue
		}

		if err = validateItem(answer, it.Key, it.Score); err != nil {
			banned[it.Conn.RemoteAddr] = true
			continue
		}

		duplicates[k] = true
		if returnFirst {
			return nil
		}

	}

	return nil
}

func AggregateListData[T any](methodFindList string, req any, methodFindItem string, getRequest func(data *AggregationListResult) (any, error), validateItem func(*T, string, float64) error, banned map[string]bool) error {
	list, err := JustAggregateData(methodFindList, req)
	if err != nil {
		return err
	}

	return AggregateData[T](list, methodFindItem, getRequest, validateItem, false, banned)
}

func AggregateBestData[T any](methodFindList string, req *api_method_sync_item.APIMethodSyncItemRequest, methodFindItem string, getRequest func(data *AggregationListResult) (any, error), validateItem func(*T, string, float64) error, banned map[string]bool) error {

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
	}); err != nil {
		return err
	}

	slices.SortFunc(list, func(a, b *AggregationListResult) bool {
		return a.Score > b.Score
	})

	return AggregateData[T](list, methodFindItem, getRequest, validateItem, true, banned)
}

func FetchData[T any](method string, data any, next func(*T, *connection.AdvancedConnection) bool) error {

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

		for _, conn := range list {

			out := conn.SendAwaitAnswer([]byte(method), b, context.Background(), 0)
			if out.Err != nil {
				gui.GUI.Error("Error sending request", out.Err)
				continue
			}

			final := new(T)
			if err = msgpack.Unmarshal(out.Out, final); err != nil {
				gui.GUI.Error("Error retrieving answer", err)
				continue
			}

			if next != nil {
				if !next(final, conn) {
					break
				}
			}

		}

		break
	}

	return nil
}
