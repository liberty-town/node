package globals

import (
	"errors"
	"liberty-town/node/config/globals/oracle"
)

type Asset struct {
	Name             string           `json:"name"`
	DecimalSeparator byte             `json:"decimalSeparator"`
	Hash             []byte           `json:"hash"`
	Oracles          []*oracle.Oracle `json:"oracles"`
}

func (this *Asset) Convert(amount uint64, decimalSeparator, finalDecimalSeparator byte) (uint64, error) {
	if len(this.Oracles) == 0 {
		return 0, errors.New("no oracles")
	}
	return this.Oracles[0].Convert(amount, decimalSeparator, finalDecimalSeparator)
}
