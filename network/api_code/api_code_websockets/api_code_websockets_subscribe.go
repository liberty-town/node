package api_code_websockets

import (
	msgpack "github.com/vmihailenco/msgpack/v5"
	"liberty-town/node/network/api_code/api_code_types"
	"liberty-town/node/network/websocks/connection"
)

func Subscribe(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	request := &api_code_types.APISubscriptionRequest{[]byte{}, api_code_types.SUBSCRIPTION_CHAT_ACCOUNT, api_code_types.RETURN_SERIALIZED}
	if err := msgpack.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.AddSubscription(request.Type, request.Key, request.ReturnType)
}

func SubscribedNotificationReceived(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	notification := &api_code_types.APISubscriptionNotification{}
	if err := msgpack.Unmarshal(values, notification); err != nil {
		return nil, err
	}

	SubscriptionNotifications.Broadcast(notification)

	return nil, nil
}

func Unsubscribe(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	unsubscribeRequest := &api_code_types.APIUnsubscriptionRequest{}
	if err := msgpack.Unmarshal(values, unsubscribeRequest); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.RemoveSubscription(unsubscribeRequest.Type, unsubscribeRequest.Key)
}
