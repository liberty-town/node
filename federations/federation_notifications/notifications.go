package federation_notifications

import (
	"liberty-town/node/federations/chat/chat_message"
	"liberty-town/node/pandora-pay/helpers/multicast"
)

var NewMessageSubscriptionsNotifications *multicast.MulticastChannel[*chat_message.ChatMessage]

func init() {
	NewMessageSubscriptionsNotifications = multicast.NewMulticastChannel[*chat_message.ChatMessage]()
}
