package globals

import (
	"errors"
	"liberty-town/node/config/globals/oracle"
	"liberty-town/node/config/globals/oracle/oracle_feed"
	"liberty-town/node/pandora-pay/helpers"
	"liberty-town/node/pandora-pay/helpers/events"
	"liberty-town/node/pandora-pay/helpers/generics"
)

var (
	MainEvents  = events.NewEvents[any]()
	MainStarted = generics.Value[bool]{}
)

var Assets = &struct {
	Currencies map[string]*Asset `json:"currencies"`
	Assets     map[string]*Asset `json:"assets"`
}{
	map[string]*Asset{
		"DOLLAR": {"Pandora Cash", 2, nil, []*oracle.Oracle{{oracle.ORACLE_VERSION_FEED, oracle_feed.NewOracleFeed("https://imprezaftx.com:2053/price?pair=PCASH/USDT"), true}}},
	},
	map[string]*Asset{
		"PCASH": {"Pandora Cash", 5, helpers.DecodeHex("0000000000000000000000000000000000000000"), []*oracle.Oracle{{oracle.ORACLE_VERSION_FEED, oracle_feed.NewOracleFeed("https://imprezaftx.com:2053/price?pair=PCASH/USDT"), false}}},
	},
}

func ConvertCurrencyToAsset(currency, asset string, amount uint64) (uint64, error) {

	c := Assets.Currencies[currency]
	if c == nil {
		return 0, errors.New("currency not found")
	}

	a := Assets.Assets[asset]
	if a == nil {
		return 0, errors.New("asset not found")
	}

	return c.Convert(amount, c.DecimalSeparator, a.DecimalSeparator)
}

func ConvertAssetToCurrency(asset, currency string, amount uint64) (uint64, error) {

	c := Assets.Currencies[currency]
	if c == nil {
		return 0, errors.New("currency not found")
	}

	a := Assets.Assets[asset]
	if a == nil {
		return 0, errors.New("asset not found")
	}

	return a.Convert(amount, a.DecimalSeparator, c.DecimalSeparator)
}

func init() {
	MainStarted.Store(false)
	//ConvertCurrencyToAsset("DOLLAR", "PCASH", 10000)
	//ConvertAssetToCurrency("PCASH", "DOLLAR", t)
}
