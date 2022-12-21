package connection

import (
	"bytes"
	"errors"
	"golang.org/x/exp/slices"
	"liberty-town/node/network/api_code/api_code_types"
	"liberty-town/node/network/network_config"
	"sync"
)

type Subscriptions struct {
	conn                 *AdvancedConnection
	list                 []*Subscription
	newSubscriptionCn    chan<- *SubscriptionNotification
	removeSubscriptionCn chan<- *SubscriptionNotification
	index                uint64
	sync.Mutex
}

func checkSubscriptionLength(key []byte, subscriptionType api_code_types.SubscriptionType) error {
	switch subscriptionType {
	case api_code_types.SUBSCRIPTION_CHAT_ACCOUNT:
	}
	return nil
}

func (s *Subscriptions) AddSubscription(subscriptionType api_code_types.SubscriptionType, key []byte, returnType api_code_types.APIReturnType) error {

	if err := checkSubscriptionLength(key, subscriptionType); err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	if len(s.list) > network_config.WEBSOCKETS_MAX_SUBSCRIPTIONS {
		return errors.New("Too many subscriptions")
	}

	for _, subscription := range s.list {
		if subscription.Type == subscriptionType && bytes.Equal(subscription.Key, key) {
			return errors.New("Already subscribed")
		}
	}

	s.index += 1

	subscription := &Subscription{subscriptionType, key, returnType}
	s.list = append(s.list, subscription)

	s.newSubscriptionCn <- &SubscriptionNotification{subscription, s.conn}

	return nil
}

func (s *Subscriptions) RemoveSubscription(subscriptionType api_code_types.SubscriptionType, key []byte) error {

	if err := checkSubscriptionLength(key, subscriptionType); err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, subscription := range s.list {
		if subscription.Type == subscriptionType && bytes.Equal(subscription.Key, key) {
			s.list = slices.Delete(s.list, i, i+1)
			s.removeSubscriptionCn <- &SubscriptionNotification{subscription, s.conn}
			return nil
		}
	}

	return errors.New("Subscription not found")
}

func NewSubscriptions(conn *AdvancedConnection, newSubscriptionCn, removeSubscriptionCn chan<- *SubscriptionNotification) (s *Subscriptions) {
	return &Subscriptions{
		conn:                 conn,
		newSubscriptionCn:    newSubscriptionCn,
		removeSubscriptionCn: removeSubscriptionCn,
	}
}
