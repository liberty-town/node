package main

import (
	"encoding/base64"
	"errors"
	msgpack "github.com/vmihailenco/msgpack/v5"
	"liberty-town/node/builds/webassembly/webassembly_utils"
	"liberty-town/node/config/globals"
	"liberty-town/node/federations/chat/chat_message"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/network/api_code/api_code_types"
	"liberty-town/node/network/api_code/api_code_websockets"
	"liberty-town/node/network/api_implementation/api_common/api_types"
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"liberty-town/node/pandora-pay/helpers/recovery"
	"sync/atomic"
	"syscall/js"
)

func listenEvents(this js.Value, args []js.Value) any {

	if len(args) == 0 || args[0].Type() != js.TypeFunction {
		return errors.New("Argument must be a callback")
	}

	index := atomic.AddUint64(&subscriptionsIndex, 1)
	channel := globals.MainEvents.AddListener()

	callback := args[0]
	var err error

	recovery.SafeGo(func() {
		for {

			data, ok := <-channel
			if !ok {
				return
			}

			var final any

			switch v := data.Data.(type) {
			case string:
				final = data.Data
			case any:
				if final, err = webassembly_utils.ConvertJSONBytes(v); err != nil {
					panic(err)
				}
			default:
				final = data.Data
			}

			callback.Invoke(data.Name, final)
		}
	})

	return index
}

func listenNetworkNotifications(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 1 || args[0].Type() != js.TypeFunction {
			return nil, errors.New("Argument must be a callback function")
		}
		callback := args[0]

		subscriptionsCn := api_code_websockets.SubscriptionNotifications.AddListener()

		recovery.SafeGo(func() {

			defer api_code_websockets.SubscriptionNotifications.RemoveChannel(subscriptionsCn)

			var err error
			for {
				data, ok := <-subscriptionsCn
				if !ok {
					return
				}

				func() {

					var object, extra any

					switch data.SubscriptionType {
					case api_code_types.SUBSCRIPTION_CHAT_ACCOUNT:
						if data.Data == nil {
							return
						}
						msg := &chat_message.ChatMessage{}
						if err = msg.Deserialize(advanced_buffers.NewBufferReader(data.Data)); err != nil {
							return
						}

						f := federation_serve.ServeFederation.Load()
						if f == nil {
							return
						}

						if msg.Validate() != nil || msg.ValidateSignatures() != nil || !f.Federation.IsValidationAccepted(msg.Validation) {
							return
						}

						object = struct {
							Message *chat_message.ChatMessage `json:"message"`
							Id      string                    `json:"id"`
						}{msg, msg.GetUniqueId()}
						extra = &api_types.APISubscriptionNotificationAccountExtra{}
					default:
						return
					}

					if err = msgpack.Unmarshal(data.Extra, extra); err != nil {
						return
					}

					jsOutData, err1 := webassembly_utils.ConvertJSONBytes(object)
					jsOutExtra, err2 := webassembly_utils.ConvertJSONBytes(extra)

					if err1 != nil || err2 != nil {
						return
					}

					callback.Invoke(int(data.SubscriptionType), base64.StdEncoding.EncodeToString(data.Key), jsOutData, jsOutExtra)

				}()

			}
		})

		return true, nil
	})
}
