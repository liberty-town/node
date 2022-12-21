package chat_message

type ChatMessageToSign struct {
	Type uint64 `json:"type"`
	Text []byte `json:"text"`
	Data []byte `json:"data"`
}
