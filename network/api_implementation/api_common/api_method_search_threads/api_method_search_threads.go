package api_method_search_threads

type APIMethodSearchThreadsRequest struct {
	Query []string `json:"query" msgpack:"query"`
	Type  byte     `json:"type" msgpack:"type"`
	Start int      `json:"start" msgpack:"start"`
}
