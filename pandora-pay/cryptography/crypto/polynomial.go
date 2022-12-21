package crypto

import (
	"liberty-town/node/pandora-pay/cryptography/bn256"
	"math/big"
)

type Polynomial struct {
	coefficients []*big.Int
}

func NewPolynomial(input []*big.Int) *Polynomial {
	if input == nil {
		return &Polynomial{coefficients: []*big.Int{new(big.Int).SetInt64(1)}}
	}
	return &Polynomial{coefficients: input}
}

func (p *Polynomial) Length() int {
	return len(p.coefficients)
}

func (p *Polynomial) Mul(m *Polynomial) *Polynomial {
	var product []*big.Int
	for i := range p.coefficients {
		product = append(product, new(big.Int).Mod(new(big.Int).Mul(p.coefficients[i], m.coefficients[0]), bn256.Order))
	}
	product = append(product, new(big.Int)) // add 0 element

	if m.coefficients[1].IsInt64() && m.coefficients[1].Int64() == 1 {
		for i := range product {
			if i > 0 {
				tmp := new(big.Int).Add(product[i], p.coefficients[i-1])

				product[i] = new(big.Int).Mod(tmp, bn256.Order)

			} else { // do nothing

			}
		}
	}
	return NewPolynomial(product)
}
