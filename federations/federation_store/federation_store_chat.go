package federation_store

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/federations/chat/chat_message"
	"liberty-town/node/federations/federation_notifications"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/store/small_sorted_set"
	"liberty-town/node/store/store_db/store_db_interface"
	"liberty-town/node/store/store_utils"
)

func storeChatMessage(tx store_db_interface.StoreDBTransactionInterface, msg *chat_message.ChatMessage, remove bool, id string) error {

	if len(id) == 0 {
		id = msg.GetUniqueId()
	}

	if remove {
		tx.Delete("messages_last:" + msg.First.Encoded + ":" + msg.Second.Encoded)
	} else {
		tx.Put("messages_last:"+msg.First.Encoded+":"+msg.Second.Encoded, []byte(id))
	}

	ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "messages_all:"+msg.First.Encoded+":"+msg.Second.Encoded, tx)
	if err := ss.Read(); err != nil {
		return err
	}
	if len(ss.Data) >= config.LIST_SIZE-1 {
		return errors.New("publisher has way to many listings")
	}
	if remove {
		ss.Delete(id)
	} else {
		ss.Add(id, float64(msg.Validation.Timestamp))
	}
	ss.Save()

	ss2 := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "conversations:"+msg.First.Encoded, tx)
	if err := ss2.Read(); err != nil {
		return err
	}
	if len(ss2.Data) >= config.LIST_SIZE-1 {
		return errors.New("publisher has way to many listings")
	}
	if remove {
		if len(ss.Dict) == 0 {
			ss2.Delete(msg.Second.Encoded)
		}
	} else {
		ss2.Add(msg.Second.Encoded, float64(msg.Validation.Timestamp))
	}
	ss2.Save()

	ss2 = small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "conversations:"+msg.Second.Encoded, tx)
	if err := ss2.Read(); err != nil {
		return err
	}
	if len(ss.Data) >= config.LIST_SIZE-1 {
		return errors.New("publisher has way to many listings")
	}
	if remove {
		if len(ss.Dict) == 0 {
			ss2.Delete(msg.First.Encoded)
		}
	} else {
		ss2.Add(msg.First.Encoded, float64(msg.Validation.Timestamp))
	}
	ss2.Save()

	return nil
}

func StoreChatMessage(msg *chat_message.ChatMessage) error {

	f := federation_serve.ServeFederation.Load()

	if f == nil || !f.Federation.Ownership.Address.Equals(msg.FederationIdentity) {
		return errors.New("not serving this federation")
	}

	if err := msg.Validate(); err != nil {
		return err
	}
	if err := msg.ValidateSignatures(); err != nil {
		return err
	}
	if !f.Federation.IsValidationAccepted(msg.Validation) {
		return errors.New("validation signature is not accepted")
	}

	if err := f.Store.DB.Update(func(tx store_db_interface.StoreDBTransactionInterface) (err error) {

		id := msg.GetUniqueId()

		score, err := store_utils.GetBetterScore("messages", id, tx)
		if err != nil {
			return err
		}
		if msg.GetBetterScore() < score {
			return errors.New("data is older")
		}

		tx.Put("messages:"+string(id), helpers.SerializeToBytes(msg))

		if err = storeChatMessage(tx, msg, false, id); err != nil {
			return
		}

		if err = store_utils.IncreaseCount("messages", id, msg.GetBetterScore(), tx); err != nil {
			return
		}

		return nil
	}); err != nil {
		return err
	}

	type Notification struct {
		Type    string `json:"type"`
		Message []byte `json:"message"`
	}

	federation_notifications.NewMessageSubscriptionsNotifications.Broadcast(msg)

	return nil
}

func GetChatConversations(address string, start int) (list []*api_types.APIMethodFindListItem, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "conversations:"+address, tx)
		if err = ss.Read(); err != nil {
			return err
		}

		for i := start; i < len(ss.Data) && len(list) < config.CHAT_CONVERSATIONS_LIST_COUNT; i++ {
			result := ss.Data[i]

			list = append(list, &api_types.APIMethodFindListItem{
				result.Key,
				result.Score,
			})
		}

		return nil
	})
	return
}

func GetChatLastMessage(first, second string) (lastMessage []byte, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	first2, second2, _ := chat_message.SortKeys(first, second)

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		id := tx.Get("messages_last:" + first2 + ":" + second2)
		if len(id) > 0 {
			lastMessage = tx.Get("messages:" + string(id))
		}

		return nil
	})
	return
}

func GetChatMessages(first, second *addresses.Address, start int) (list []*api_types.APIMethodFindListItem, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	first2, second2, _ := chat_message.SortKeys(first.Encoded, second.Encoded)

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {

		ss := small_sorted_set.NewSmallSortedSet(config.LIST_SIZE, "messages_all:"+first2+":"+second2, tx)
		if err = ss.Read(); err != nil {
			return err
		}

		for i := start; i < len(ss.Data) && len(list) < config.CHAT_MESSAGES_LIST_COUNT; i++ {
			result := ss.Data[i]

			list = append(list, &api_types.APIMethodFindListItem{
				result.Key,
				result.Score,
			})
		}

		return nil
	})
	return
}

func GetChatMessage(id string) (message []byte, err error) {

	f := federation_serve.ServeFederation.Load()
	if f == nil {
		return nil, errors.New("not serving this federation")
	}

	err = f.Store.DB.View(func(tx store_db_interface.StoreDBTransactionInterface) error {
		message = tx.Get("messages:" + id)
		return nil
	})
	return
}
