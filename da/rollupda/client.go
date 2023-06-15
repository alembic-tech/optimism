package rollupda

import "github.com/ethereum/go-ethereum/common"
import "github.com/ethereum-optimism/optimism/da"

type client struct {
  addr common.Address
}

type batchRef struct {
  to common.Address
  data []byte
}

func (c *batchRef) ToTx() (da.Tx, error) {
  return da.Tx{
    To: &c.to,
    Data: c.data,
  }, nil
}


func NewClient(addr common.Address) da.Client {
  return &client{}
}

func (c *client) PostBatch(data []byte) (da.BatchRef, error) {
  return &batchRef{c.addr, data}, nil
}

func (c *client) GetBatch(data []byte) ([]byte, error) {
  return data, nil
}
