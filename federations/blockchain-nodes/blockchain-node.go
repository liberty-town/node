package blockchain_nodes

import (
	"liberty-town/node/pandora-pay/helpers/advanced_buffers"
	"net/url"
)

type BlockchainNode struct {
	URL string `json:"url"`
}

func (this *BlockchainNode) Validate() error {
	_, err := url.Parse(this.URL)
	if err != nil {
		return err
	}
	return nil
}

func (this *BlockchainNode) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteString(this.URL)
}

func (this *BlockchainNode) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if this.URL, err = r.ReadString(100); err != nil {
		return
	}
	return
}
