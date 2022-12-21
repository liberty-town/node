package oracle_feed

import (
	"errors"
	"liberty-town/node/network/request"
	"math/big"
	"time"
)

type OracleFeed struct {
	Address  string `json:"address"`
	price    *big.Float
	lastTime uint64
}

func NewOracleFeed(address string) *OracleFeed {
	return &OracleFeed{address, nil, 0}
}

func (this *OracleFeed) GetPrice() (*big.Float, error) {

	t := uint64(time.Now().Unix())
	if t > this.lastTime*10*1000 {

		b, err := request.RequestGetData(this.Address)
		if err != nil {
			return nil, err
		}

		final, ok := new(big.Float).SetString(string(b))
		if !ok {
			return nil, errors.New("invalid price")
		}

		if final.Sign() == 0 {
			return nil, errors.New("price should not be zero")
		}

		this.price = final
		this.lastTime = t
	}

	return this.price, nil
}
