package invoices

import (
	"errors"
	"liberty-town/node/addresses"
	"liberty-town/node/config"
	"liberty-town/node/cryptography"
	"liberty-town/node/federations"
)

type InvoiceIdBuyerAddress struct {
	Address          *addresses.Address `json:"address" msgpack:"address"`
	Nonce            []byte             `json:"nonce" msgpack:"nonce"`
	ConversionAsset  string             `json:"conversionAsset" msgpack:"conversionAsset"`
	ConversionAmount uint64             `json:"conversionAmount" msgpack:"conversionAmount"`
}

type InvoiceIdSellerAddress struct {
	Address *addresses.Address `json:"address" msgpack:"address"`
	Nonce   []byte             `json:"nonce" msgpack:"nonce"`
}

type InvoiceId struct {
	Federation *addresses.Address      `json:"federation" msgpack:"federation"`
	Moderator  *addresses.Address      `json:"moderator" msgpack:"moderator"`
	Buyer      *InvoiceIdBuyerAddress  `json:"buyer" msgpack:"buyer"`
	Seller     *InvoiceIdSellerAddress `json:"seller" msgpack:"seller"`
	Total      uint64                  `json:"total" msgpack:"total"`
}

func (this *InvoiceId) Validate() error {

	if this.Buyer == nil {
		return errors.New("buyer missing")
	}
	if this.Seller == nil {
		return errors.New("seller missing")
	}

	if len(this.Buyer.Nonce) != cryptography.HashSize || len(this.Seller.Nonce) != cryptography.HashSize {
		return errors.New("invalid nonce")
	}

	if this.Buyer.Address.Network != config.NETWORK_SELECTED || this.Seller.Address.Network != config.NETWORK_SELECTED || this.Federation.Network != config.NETWORK_SELECTED || this.Moderator.Network != config.NETWORK_SELECTED {
		return errors.New("invalid network")
	}

	if this.Buyer.Address.Encoded == this.Seller.Address.Encoded {
		return errors.New("buyer and seller are both the same")
	}

	if this.Total == 0 {
		return errors.New("invalid total")
	}

	f, _ := federations.FederationsDict.Load(this.Federation.Encoded)
	if f == nil {
		return errors.New("federation not found")
	}

	mod := f.FindModerator(this.Moderator.Encoded)
	if mod == nil {
		return errors.New("moderator not found")
	}

	return nil
}
