package oracle

import (
	"liberty-town/node/config/globals/oracle/oracle_interface"
	"math/big"
)

type Oracle struct {
	Version OracleVersion                    `json:"version"`
	Data    oracle_interface.OracleInterface `json:"data"`
	Invert  bool                             `json:"invert"` // true 1/x
}

func pow(exp byte) *big.Float {
	final := new(big.Float).SetUint64(1)
	for i := byte(0); i < exp; i++ {
		final = new(big.Float).Mul(final, new(big.Float).SetUint64(10))
	}
	return final
}

func (this *Oracle) Convert(amount uint64, decimalsAmount, finalDecimalsAmount byte) (uint64, error) {

	price, err := this.Data.GetPrice()
	if err != nil {
		return 0, err
	}

	final := pow(finalDecimalsAmount)
	final = new(big.Float).Mul(final, new(big.Float).SetUint64(amount))

	if this.Invert {
		final = new(big.Float).Quo(final, price)
	} else {
		final = new(big.Float).Mul(final, price)
	}

	final = new(big.Float).Quo(final, pow(decimalsAmount))

	result, _ := final.Uint64()

	return result, err
}
