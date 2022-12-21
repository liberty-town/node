package api_code_types

type SubscriptionType uint8

const (
	SUBSCRIPTION_CHAT_ACCOUNT SubscriptionType = iota
)

type APISubscriptionNotification struct {
	SubscriptionType SubscriptionType `json:"type,omitempty" msgpack:"type,omitempty"`
	Key              []byte           `json:"key,omitempty" msgpack:"key,omitempty"`
	Data             []byte           `json:"data,omitempty" msgpack:"data,omitempty"`
	Extra            []byte           `json:"extra,omitempty" msgpack:"extra,omitempty"`
}

type APISubscriptionRequest struct {
	Key        []byte           `json:"key,omitempty" msgpack:"key,omitempty"`
	Type       SubscriptionType `json:"type,omitempty"  msgpack:"type,omitempty"`
	ReturnType APIReturnType    `json:"returnType,omitempty"  msgpack:"returnType,omitempty"`
}

type APIUnsubscriptionRequest struct {
	Key  []byte           `json:"key,omitempty" msgpack:"key,omitempty"`
	Type SubscriptionType `json:"type,omitempty" msgpack:"type,omitempty"`
}
