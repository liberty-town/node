package api_types

type APIMethodGetRequest struct {
	Identity string `json:"identity" msgpack:"identity"`
}

type APIMethodGetResult struct {
	Result []byte `json:"result" msgpack:"result"`
}

type APIMethodStoreRequest struct {
	Data []byte `json:"data" msgpack:"data"`
}

type APIMethodStoreResult struct {
	Result bool `json:"result" msgpack:"result"`
}

type APIMethodFindListItem struct {
	Key   string  `json:"key" msgpack:"key"`
	Score float64 `json:"score" msgpack:"score"`
}

type APIMethodFindListResult struct {
	Results []*APIMethodFindListItem `json:"results" msgpack:"results"`
}

type APISubscriptionExtra struct {
	Index uint64 `json:"index" msgpack:"index"`
}

type APISubscriptionNotificationAccountExtra struct {
}
