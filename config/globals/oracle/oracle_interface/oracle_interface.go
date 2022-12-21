package oracle_interface

import "math/big"

type OracleInterface interface {
	GetPrice() (*big.Float, error)
}
