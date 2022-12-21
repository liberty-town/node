package main

import (
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations/federation"
	"liberty-town/node/federations/federation_store/store_data/listings"
	"liberty-town/node/federations/moderator"
	"liberty-town/node/invoices"
	"liberty-town/node/network/api_code/api_code_types"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_type"
	"liberty-town/node/pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
	pandora_pay_config_coins "liberty-town/node/pandora-pay/config/config_coins"
	pandora_pay_cryptography "liberty-town/node/pandora-pay/cryptography"
	"liberty-town/node/start"
	"liberty-town/node/validator/validation/validation_type"
	"syscall/js"
)

var subscriptionsIndex uint64

func Initialize() {

	js.Global().Set("LibertyTown", js.ValueOf(map[string]any{
		"helpers": js.ValueOf(map[string]any{
			"test":         js.FuncOf(test),
			"start":        js.FuncOf(startLibrary),
			"getIdenticon": js.FuncOf(getIdenticon),
		}),
		"addresses": js.ValueOf(map[string]any{
			"decodeAddress": js.FuncOf(decodeAddress),
		}),
		"settings": js.ValueOf(map[string]any{
			"get": js.FuncOf(settingsGet),
			"manager": js.ValueOf(map[string]any{
				"getSecretWords":      js.FuncOf(settingsGetSecretWords),
				"getSecretEntropy":    js.FuncOf(settingsGetSecretEntropy),
				"importSecretWords":   js.FuncOf(settingsImportSecretWords),
				"importSecretEntropy": js.FuncOf(settingsImportSecretEntropy),
				"clear":               js.FuncOf(settingsClear),
				"exportJSON":          js.FuncOf(settingsExportJSON),
				"exportJSONAll":       js.FuncOf(settingsExportJSONAll),
				"importJSON":          js.FuncOf(settingsImportJSON),
				"rename":              js.FuncOf(settingsRename),
			}),
		}),
		"app": js.ValueOf(map[string]any{
			"federationReplaceValidatorContactAddresses": js.FuncOf(appFederationReplaceValidatorContactAddresses),
			"getFederations":        js.FuncOf(appGetFederations),
			"setSelectedFederation": js.FuncOf(appSetSelectedFederation),
			"assets": js.ValueOf(map[string]any{
				"get":                    js.FuncOf(appGetAssets),
				"convertCurrencyToAsset": js.FuncOf(appConvertCurrencyToAsset),
				"convertAssetToCurrency": js.FuncOf(appConvertAssetToCurrency),
			}),
		}),
		"events": js.ValueOf(map[string]any{
			"listenEvents":               js.FuncOf(listenEvents),
			"listenNetworkNotifications": js.FuncOf(listenNetworkNotifications),
		}),
		"crypto": js.ValueOf(map[string]any{
			"randomBytes":     js.FuncOf(cryptoRandomBytes),
			"HASH_SIZE":       js.ValueOf(cryptography.HashSize),
			"SIGNATURE_SIZE":  js.ValueOf(cryptography.SignatureSize),
			"PUBLIC_KEY_SIZE": js.ValueOf(cryptography.PublicKeySize),
			"sign":            js.FuncOf(sign),
			"verify":          js.FuncOf(verify),
		}),
		"accounts": js.ValueOf(map[string]any{
			"store": js.FuncOf(accountStore),
			"get":   js.FuncOf(accountGet),
		}),
		"listings": js.ValueOf(map[string]any{
			"store":  js.FuncOf(listingStore),
			"get":    js.FuncOf(listingGet),
			"getAll": js.FuncOf(listingsGetAll),
			"search": js.FuncOf(listingsSearch),
		}),
		"accountsSummaries": js.ValueOf(map[string]any{
			"get":   js.FuncOf(accountSummaryGet),
			"store": js.FuncOf(accountSummaryStore),
		}),
		"listingsSummaries": js.ValueOf(map[string]any{
			"get":   js.FuncOf(listingSummaryGet),
			"store": js.FuncOf(listingSummaryStore),
		}),
		"reviews": js.ValueOf(map[string]any{
			"store":  js.FuncOf(reviewStore),
			"getAll": js.FuncOf(reviewsGetAll),
		}),
		"invoices": js.ValueOf(map[string]any{
			"validate":           js.FuncOf(invoiceValidate),
			"validateConfirmed":  js.FuncOf(invoiceValidateConfirmed),
			"sign":               js.FuncOf(invoiceSign),
			"createId":           js.FuncOf(invoiceCreateId),
			"messageToSignItems": js.FuncOf(invoiceMessageToSignItems),
			"serialize":          js.FuncOf(invoiceSerialize),
			"deserialize":        js.FuncOf(invoiceDeserialize),
			"multisig": js.ValueOf(map[string]any{
				"compute":       js.FuncOf(invoiceMultisigCompute),
				"sign":          js.FuncOf(invoiceMultisigSign),
				"moderatorSign": js.FuncOf(invoiceModeratorMultisigSign),
				"verify":        js.FuncOf(invoiceMultisigVerify),
				"claim":         js.FuncOf(invoiceMultisigClaimTx),
			}),
		}),
		"chat": js.ValueOf(map[string]any{
			"getConversations": js.FuncOf(chatGetConversations),
			"sendMessage":      js.FuncOf(chatSendMessage),
			"getMessages":      js.FuncOf(chatGetMessages),
			"decryptMessage":   js.FuncOf(chatDecryptMessage),
		}),
		"config": js.ValueOf(map[string]any{
			"LISTING_IMAGES_MAX_COUNT":       js.ValueOf(config.LISTING_IMAGES_MAX_COUNT),
			"LISTING_CATEGORIES_MAX_COUNT":   js.ValueOf(config.LISTING_CATEGORIES_MAX_COUNT),
			"LISTING_SHIPPING_TO_MAX_COUNT":  js.ValueOf(config.LISTING_SHIPPING_TO_MAX_COUNT),
			"LISTING_TITLE_MIN_LENGTH":       js.ValueOf(config.LISTING_TITLE_MIN_LENGTH),
			"LISTING_TITLE_MAX_LENGTH":       js.ValueOf(config.LISTING_TITLE_MAX_LENGTH),
			"LISTING_IMAGE_MAX_LENGTH":       js.ValueOf(config.LISTING_IMAGE_MAX_LENGTH),
			"LISTING_DESCRIPTION_MAX_LENGTH": js.ValueOf(config.LISTING_DESCRIPTION_MAX_LENGTH),
			"LISTING_OFFERS_MAX_COUNT":       js.ValueOf(config.LISTING_OFFERS_MAX_COUNT),
			"LISTING_OFFER_MAX_LENGTH":       js.ValueOf(config.LISTING_OFFER_MAX_LENGTH),
			"LISTING_OFFER_MIN_LENGTH":       js.ValueOf(config.LISTING_OFFER_MIN_LENGTH),
			"LISTING_SHIPPING_MAX_COUNT":     js.ValueOf(config.LISTING_SHIPPING_MAX_COUNT),
			"LISTING_SHIPPING_MAX_LENGTH":    js.ValueOf(config.LISTING_SHIPPING_MAX_LENGTH),
			"LISTING_SHIPPING_MIN_LENGTH":    js.ValueOf(config.LISTING_SHIPPING_MIN_LENGTH),
			"REVIEW_TITLE_MAX_LENGTH":        js.ValueOf(config.REVIEW_TITLE_MAX_LENGTH),
			"REVIEWS_LIST_COUNT":             js.ValueOf(config.REVIEWS_LIST_COUNT),
			"CHAT_MESSAGES_LIST_COUNT":       js.ValueOf(config.CHAT_MESSAGES_LIST_COUNT),
			"CHAT_CONVERSATIONS_LIST_COUNT":  js.ValueOf(config.CHAT_CONVERSATIONS_LIST_COUNT),
			"CHAT_MESSAGE_MAX_LENGTH":        js.ValueOf(config.CHAT_MESSAGE_MAX_LENGTH),
			"LISTINGS_LIST_COUNT":            js.ValueOf(config.LISTINGS_LIST_COUNT),
		}),
		"enums": js.ValueOf(map[string]any{
			"moderators": js.ValueOf(map[string]any{
				"MODERATOR_PANDORA": js.ValueOf(uint64(moderator.MODERATOR_PANDORA)),
			}),
			"listings": js.ValueOf(map[string]any{
				"LISTING_VERSION": js.ValueOf(uint64(listings.LISTING_VERSION)),
			}),
			"federations": js.ValueOf(map[string]any{
				"FEDERATION_VERSION": js.ValueOf(uint64(federation.FEDERATION_VERSION)),
			}),
			"invoices": js.ValueOf(map[string]any{
				"INVOICE_VERSION_0": js.ValueOf(uint64(invoices.INVOICE_VERSION_0)),
			}),
			"invoiceItems": js.ValueOf(map[string]any{
				"INVOICE_ITEM_NEW": js.ValueOf(uint64(invoices.INVOICE_ITEM_NEW)),
				"INVOICE_ITEM_ID":  js.ValueOf(uint64(invoices.INVOICE_ITEM_ID)),
			}),
			"api": js.ValueOf(map[string]any{
				"websockets": js.ValueOf(map[string]any{
					"subscriptionType": js.ValueOf(map[string]any{
						"SUBSCRIPTION_CHAT_ACCOUNT": js.ValueOf(int(api_code_types.SUBSCRIPTION_CHAT_ACCOUNT)),
					}),
				}),
			}),
			"validators": js.ValueOf(map[string]any{
				"validations": js.ValueOf(map[string]any{
					"VALIDATOR_CHALLENGE_NO_CAPTCHA": js.ValueOf(uint64(validation_type.VALIDATOR_CHALLENGE_NO_CAPTCHA)),
					"VALIDATOR_CHALLENGE_HCAPTCHA":   js.ValueOf(uint64(validation_type.VALIDATOR_CHALLENGE_HCAPTCHA)),
					"VALIDATOR_CHALLENGE_CUSTOM":     js.ValueOf(uint64(validation_type.VALIDATOR_CHALLENGE_CUSTOM)),
				}),
			}),
		}),
		"PandoraPay": js.ValueOf(map[string]any{
			"addresses": js.ValueOf(map[string]any{
				"decodeAddress": js.FuncOf(pandoraPayDecodeAddress),
				"createAddress": js.FuncOf(pandoraPayCreateAddress),
			}),
			"cryptography": js.ValueOf(map[string]any{
				"SIGNATURE_SIZE":  js.ValueOf(pandora_pay_cryptography.SignatureSize),
				"PUBLIC_KEY_SIZE": js.ValueOf(pandora_pay_cryptography.PublicKeySize),
			}),
			"config": js.ValueOf(map[string]any{
				"coins": js.ValueOf(map[string]any{
					"ASSET_LENGTH":                    js.ValueOf(pandora_pay_config_coins.ASSET_LENGTH),
					"NATIVE_ASSET_FULL_STRING_BASE64": js.ValueOf(pandora_pay_config_coins.NATIVE_ASSET_FULL_STRING_BASE64),
				}),
			}),
			"enums": js.ValueOf(map[string]any{
				"transactions": js.ValueOf(map[string]any{
					"TransactionVersion": js.ValueOf(map[string]any{
						"TX_SIMPLE": js.ValueOf(uint64(transaction_type.TX_SIMPLE)),
						"TX_ZETHER": js.ValueOf(uint64(transaction_type.TX_ZETHER)),
					}),
					"transactionSimple": js.ValueOf(map[string]any{
						"ScriptType": js.ValueOf(map[string]any{
							"SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY":     js.ValueOf(uint64(transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY)),
							"SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT": js.ValueOf(uint64(transaction_simple.SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT)),
						}),
					}),
					"transactionZether": js.ValueOf(map[string]any{
						"PayloadScriptType": js.ValueOf(map[string]any{
							"SCRIPT_TRANSFER":              js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_TRANSFER)),
							"SCRIPT_STAKING":               js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_STAKING)),
							"SCRIPT_STAKING_REWARD":        js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_STAKING_REWARD)),
							"SCRIPT_SPEND":                 js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_SPEND)),
							"SCRIPT_ASSET_CREATE":          js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_ASSET_CREATE)),
							"SCRIPT_ASSET_SUPPLY_INCREASE": js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_ASSET_SUPPLY_INCREASE)),
							"SCRIPT_PLAIN_ACCOUNT_FUND":    js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_PLAIN_ACCOUNT_FUND)),
							"SCRIPT_CONDITIONAL_PAYMENT":   js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_CONDITIONAL_PAYMENT)),
						}),
					}),
				}),
			}),
		}),
	}))

}

func main() {
	if err := start.InitMain(func() {
		Initialize()
		js.Global().Call("WASMLoaded")
	}); err != nil {
		panic(err)
	}
}
