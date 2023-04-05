package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/chat/chat_message"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_conversations"
	"liberty-town/node/network/api_implementation/api_common/api_method_find_messages"
	"liberty-town/node/network/api_implementation/api_common/api_method_get_last_msg"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/network/websocks/connection"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/msgpack"
	"liberty-town/node/settings"
	"liberty-town/node/validator/validation"
	"syscall/js"
)

func chatGetConversations(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation selected")
		}

		req := &struct {
			Start int `json:"start"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		sender := settings.Settings.Load().Account.Address

		type SearchResult struct {
			Key     string                    `json:"key"`
			Score   float64                   `json:"score"`
			Message *chat_message.ChatMessage `json:"message"`
		}

		count := 0
		err := federation_network.AggregateListData[api_types.APIMethodGetResult]("find-conversations", &api_method_find_conversations.APIMethodFindConversationsRequest{
			sender.Encoded,
			req.Start,
		}, "get-last-msg", func(it *federation_network.AggregationListResult) (any, error) {
			return &api_method_get_last_msg.APIMethodGetLastMessageRequest{
				sender.Encoded, it.Key,
			}, nil
		}, func(answer *api_types.APIMethodGetResult, key string, score float64) error {

			msg := &chat_message.ChatMessage{}
			if err := msg.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
				return err
			}

			if msg.Validate() != nil || msg.ValidateSignatures() != nil || !f.Federation.IsValidationAccepted(msg.Validation) {
				return errors.New("msg invalid")
			}

			if score > float64(msg.Validation.Timestamp) {
				return errors.New("msg score invalid")
			}

			if (msg.First.Equals(sender) && msg.Second.Encoded == key) || (msg.First.Encoded == key && msg.Second.Equals(sender)) {
				result := &SearchResult{key, score, msg}
				b, err := webassembly_utils.ConvertJSONBytes(result)
				if err != nil {
					return err
				}

				args[1].Invoke(b)
				count += 1
				return nil
			}

			return errors.New("msg invalid")

		}, nil)

		return count, err
	})
}

func chatSendMessage(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Message struct {
				Type      uint64 `json:"type"`
				Text      []byte `json:"text"`
				Data      []byte `json:"data"`
				Signature []byte `json:"signature,omitempty"`
			} `json:"message"`
			Receiver *addresses.Address `json:"receiver"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		chatSettings := settings.Settings.Load()
		sender := chatSettings.Account.Address

		b, err := json.Marshal(&chat_message.ChatMessageToSign{
			req.Message.Type,
			req.Message.Text,
			req.Message.Data,
		})
		if err != nil {
			return nil, err
		}

		if req.Message.Signature, err = chatSettings.Account.PrivateKey.Sign(b); err != nil {
			return nil, nil
		}

		if b, err = msgpack.Marshal(req.Message); err != nil {
			return nil, err
		}

		//ask haOr3n
		encryptedReceiver, err := req.Receiver.EncryptMessage(b)
		if err != nil {
			return nil, err
		}

		encryptedSender, err := chatSettings.Account.Address.EncryptMessage(b)
		if err != nil {
			return nil, err
		}

		_, _, ok := chat_message.SortKeys(sender.Encoded, req.Receiver.Encoded)

		msg := &chat_message.ChatMessage{
			chat_message.CHAT_MESSAGE,
			f.Federation.Ownership.Address,
			nil,
			nil,
			cryptography.RandomBytes(cryptography.HashSize),
			encryptedSender,
			encryptedReceiver,
			&validation.Validation{},
		}

		if ok {
			msg.First = req.Receiver
			msg.Second = sender
			msg.FirstMessage = encryptedReceiver
			msg.SecondMessage = encryptedSender
		} else {
			msg.First = sender
			msg.Second = req.Receiver
		}

		if msg.Validation, _, err = federationValidate(f.Federation, msg.GetMessageForSigningValidator, args[1], nil); err != nil {
			return nil, err
		}

		if err = msg.Validate(); err != nil {
			return nil, err
		}

		results := 0
		if err = federation_network.FetchData[api_types.APIMethodStoreResult]("store-msg", api_types.APIMethodStoreRequest{helpers.SerializeToBytes(msg)}, func(a *api_types.APIMethodStoreResult, b *connection.AdvancedConnection) bool {
			if a != nil && a.Result {
				results++
			}
			return true
		}); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			ChatMessage *chat_message.ChatMessage `json:"message"`
			Id          string                    `json:"id"`
			Results     int                       `json:"results"`
		}{msg, msg.GetUniqueId(), results})

	})
}

func chatGetMessages(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (any, error) {

		f := federation_serve.ServeFederation.Load()
		if f == nil {
			return nil, errors.New("no federation")
		}

		req := &struct {
			Receiver *addresses.Address `json:"receiver"`
			Start    int                `json:"start"`
		}{}
		if err := webassembly_utils.UnmarshalBytes(args[0], req); err != nil {
			return nil, err
		}

		sender := settings.Settings.Load().Account.Address

		type SearchResult struct {
			Key     string                    `json:"key"`
			Score   float64                   `json:"score"`
			Message *chat_message.ChatMessage `json:"message"`
		}

		count := 0
		err := federation_network.AggregateListData[api_types.APIMethodGetResult]("find-msgs", &api_method_find_messages.APIMethodFindMessagesRequest{
			sender,
			req.Receiver,
			req.Start,
		}, "get-msg", nil, func(answer *api_types.APIMethodGetResult, key string, score float64) error {

			msg := &chat_message.ChatMessage{}
			if err := msg.Deserialize(advanced_buffers.NewBufferReader(answer.Result)); err != nil {
				return err
			}

			if msg.Validate() != nil || msg.ValidateSignatures() != nil || !f.Federation.IsValidationAccepted(msg.Validation) {
				return errors.New("msg invalid")
			}

			if score > float64(msg.Validation.Timestamp) {
				return errors.New("msg score invalid")
			}

			if msg.GetUniqueId() != key {
				return errors.New("msg id is invalid")
			}

			if (msg.First.Equals(sender) && msg.Second.Equals(req.Receiver)) || (msg.First.Equals(req.Receiver) && msg.Second.Equals(sender)) {
				result := &SearchResult{key, score, msg}
				b, err := webassembly_utils.ConvertJSONBytes(result)
				if err != nil {
					return err
				}

				args[1].Invoke(b)
				count++
				return nil
			}

			return errors.New("msg invalid")

		}, nil)

		return count, err

	})
}

func chatDecryptMessage(this js.Value, args []js.Value) any {
	return webassembly_utils.PromiseFunction(func() (out any, err error) {

		type ChatMessageDecrypted struct {
			Type      uint64             `json:"type"`
			Data      []byte             `json:"data"`
			Text      []byte             `json:"text"`
			Signature []byte             `json:"signature,omitempty"`
			Address   *addresses.Address `json:"address,omitempty"`
		}

		data := &struct {
			Message  *chat_message.ChatMessage `json:"message"`
			Receiver string                    `json:"receiver"`
		}{}

		if err = webassembly_utils.UnmarshalBytes(args[0], data); err != nil {
			return nil, err
		}

		receiverAddr, err := addresses.DecodeAddr(data.Receiver)
		if err != nil {
			return nil, err
		}

		settings := settings.Settings.Load()

		m1, err1 := settings.Account.PrivateKey.Decrypt(data.Message.FirstMessage)
		m2, err2 := settings.Account.PrivateKey.Decrypt(data.Message.SecondMessage)

		decrypted := new(ChatMessageDecrypted)
		if err1 == nil {
			err = msgpack.Unmarshal(m1, decrypted)
		} else if err2 == nil {
			err = msgpack.Unmarshal(m2, decrypted)
		} else {
			err = errors.New("Invalid message")
		}

		if err != nil {
			return nil, err
		}

		b, err := json.Marshal(&chat_message.ChatMessageToSign{
			decrypted.Type,
			decrypted.Text,
			decrypted.Data,
		})
		if err != nil {
			return nil, err
		}

		publicKey, err := cryptography.EcrecoverCompressed(cryptography.SHA3(b), decrypted.Signature)
		if err != nil {
			return nil, err
		}

		if bytes.Equal(publicKey, settings.Account.Address.PublicKey) {
			decrypted.Address = settings.Account.Address
		} else if bytes.Equal(publicKey, receiverAddr.PublicKey) {
			decrypted.Address = receiverAddr
		} else {
			return nil, errors.New("signer is unknown")
		}

		return webassembly_utils.ConvertJSONBytes(decrypted)
	})
}
