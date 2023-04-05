package websocks

import (
	"liberty-town/node/federations/federation_notifications"
	"liberty-town/node/network/api_code/api_code_types"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/network_config"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/network/websocks/connection/advanced_connection_types"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/msgpack"
	"liberty-town/node/pandora-pay/helpers/recovery"
)

type WebsocketSubscriptions struct {
	websocketClosedCn     chan *connection.AdvancedConnection
	newSubscriptionCn     chan *connection.SubscriptionNotification
	removeSubscriptionCn  chan *connection.SubscriptionNotification
	accountsSubscriptions map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification
}

func newWebsocketSubscriptions() (subs *WebsocketSubscriptions) {

	subs = &WebsocketSubscriptions{
		make(chan *connection.AdvancedConnection),
		make(chan *connection.SubscriptionNotification),
		make(chan *connection.SubscriptionNotification),
		make(map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification),
	}

	if network_config.NETWORK_ENABLE_SUBSCRIPTIONS {
		recovery.SafeGo(subs.processSubscriptions)
	}

	return
}

func (this *WebsocketSubscriptions) send(subscriptionType api_code_types.SubscriptionType, apiRoute []byte, key []byte, list map[advanced_connection_types.UUID]*connection.SubscriptionNotification, element helpers.SerializableInterface, elementBytes []byte, extra interface{}) {

	var err error
	var extraMarshalled []byte
	var serialized, marshalled *api_code_types.APISubscriptionNotification

	if extra != nil {
		if extraMarshalled, err = msgpack.Marshal(extra); err != nil {
			panic(err)
		}
	}

	for _, subNot := range list {

		if element == nil && elementBytes == nil && extra == nil {
			_ = subNot.Conn.Send(key, nil, 0)
			continue
		}

		if subNot.Subscription.ReturnType == api_code_types.RETURN_SERIALIZED {
			var bytes []byte
			if element != nil {
				bytes = helpers.SerializeToBytes(element)
			} else {
				bytes = elementBytes
			}

			if serialized == nil {
				serialized = &api_code_types.APISubscriptionNotification{subscriptionType, key, bytes, extraMarshalled}
			}
			_ = subNot.Conn.SendJSON(apiRoute, serialized, 0)
		} else if subNot.Subscription.ReturnType == api_code_types.RETURN_JSON {
			if marshalled == nil {
				var bytes []byte
				if element != nil {
					if bytes, err = msgpack.Marshal(element); err != nil {
						panic(err)
					}
				} else {
					bytes = elementBytes
				}
				marshalled = &api_code_types.APISubscriptionNotification{subscriptionType, key, bytes, extraMarshalled}
			}
			_ = subNot.Conn.SendJSON(apiRoute, marshalled, 0)
		}

	}
}

func (this *WebsocketSubscriptions) getSubsMap(subscriptionType api_code_types.SubscriptionType) (subsMap map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification) {
	switch subscriptionType {
	case api_code_types.SUBSCRIPTION_CHAT_ACCOUNT:
		subsMap = this.accountsSubscriptions
	}
	return
}

func (this *WebsocketSubscriptions) removeConnection(conn *connection.AdvancedConnection, subscriptionType api_code_types.SubscriptionType) {

	subsMap := this.getSubsMap(subscriptionType)

	var deleted []string
	for key, value := range subsMap {
		if value[conn.UUID] != nil {
			delete(value, conn.UUID)
		}
		if len(value) == 0 {
			deleted = append(deleted, key)
		}
	}
	for _, key := range deleted {
		delete(subsMap, key)
	}
}

func (this *WebsocketSubscriptions) processSubscriptions() {

	updateNewMessageNotificationsCn := federation_notifications.NewMessageSubscriptionsNotifications.AddListener()
	defer federation_notifications.NewMessageSubscriptionsNotifications.RemoveChannel(updateNewMessageNotificationsCn)

	var subsMap map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification

	for {

		select {
		case subscription := <-this.newSubscriptionCn:

			if subsMap = this.getSubsMap(subscription.Subscription.Type); subsMap == nil {
				continue
			}

			keyStr := string(subscription.Subscription.Key)
			if subsMap[keyStr] == nil {
				subsMap[keyStr] = make(map[advanced_connection_types.UUID]*connection.SubscriptionNotification)
			}
			subsMap[keyStr][subscription.Conn.UUID] = subscription

		case subscription := <-this.removeSubscriptionCn:

			if subsMap = this.getSubsMap(subscription.Subscription.Type); subsMap == nil {
				continue
			}

			keyStr := string(subscription.Subscription.Key)
			if subsMap[keyStr] != nil {
				delete(subsMap[keyStr], subscription.Conn.UUID)
				if len(subsMap[keyStr]) == 0 {
					delete(subsMap, keyStr)
				}
			}

		case newMessage := <-updateNewMessageNotificationsCn:

			if list := this.accountsSubscriptions[string(helpers.SerializeToBytes(newMessage.First))]; list != nil {
				this.send(api_code_types.SUBSCRIPTION_CHAT_ACCOUNT, []byte("sub/notify"), nil, list, newMessage, nil, &api_types.APISubscriptionNotificationAccountExtra{})
			}

			if list := this.accountsSubscriptions[string(helpers.SerializeToBytes(newMessage.Second))]; list != nil {
				this.send(api_code_types.SUBSCRIPTION_CHAT_ACCOUNT, []byte("sub/notify"), nil, list, newMessage, nil, &api_types.APISubscriptionNotificationAccountExtra{})
			}

		case conn, ok := <-this.websocketClosedCn:
			if !ok {
				return
			}

			this.removeConnection(conn, api_code_types.SUBSCRIPTION_CHAT_ACCOUNT)

		}

	}

}
