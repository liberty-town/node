package api_method_find_conversations

type APIMethodFindConversationsRequest struct {
	First string `json:"first" msgpack:"first"`
	Start int    `json:"start" msgpack:"start"`
}
