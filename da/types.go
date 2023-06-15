package da

import "github.com/ethereum/go-ethereum/common"

type Tx struct {
  To *common.Address
  Data []byte
}

type BatchRef interface {
  ToTx() (Tx, error)
}

type Client interface {
  PostBatch(data []byte) (BatchRef, error)
  GetBatch(ref []byte) ([]byte, error)
}
