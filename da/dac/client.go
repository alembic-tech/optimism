package dac

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ethereum-optimism/optimism/da"
	"github.com/ethereum/go-ethereum/common"
)

const (
  DACBatchHeaderID uint8 = 1
)

var (
  ErrInvalidBatchSignature = errors.New("invalid batch signature")
)

type client struct {
  url *url.URL
  addr common.Address
  keyset KeySet
}

type batchRef struct {
  addr common.Address
  dataHash []byte
  signature []byte
  mask uint64
}

func (r *batchRef) ToTx() (da.Tx, error) {
  data := make([]byte, 0, 1 + len(r.dataHash) + len(r.signature) + 8)

  data = append(data, DACBatchHeaderID)
  data = append(data, r.dataHash...)
  data = append(data, r.signature...)
  data = binary.BigEndian.AppendUint64(data, r.mask)

  return da.Tx{
    To: &r.addr,
    Data: data,
  }, nil
}


// FIXME: remove addr
func NewClient(apiUrl string, addr common.Address, keyset KeySet) da.Client {
  parsed, err := url.Parse(apiUrl)
  if err != nil {
    panic(fmt.Errorf("invalid DA url: %w", err))
  }
  return &client{parsed, addr, keyset}
}

func (c *client) PostBatch(data []byte) (da.BatchRef, error) {
  apiUrl := *c.url
  apiUrl.Path = "batch"

  httpClient := http.DefaultClient

  type payload struct {
    Data string `json:"data"`
  }
  p := payload{
    Data: hex.EncodeToString(data),
  }
  encoded, _ := json.Marshal(p)
  resp, err := httpClient.Post(apiUrl.String(), "application/json", bytes.NewReader(encoded))
  if err != nil {
    return nil, fmt.Errorf("could not post batch: %w", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("invalid post batch response code: %v", resp.StatusCode)
  }

  type response struct {
    DataHash string `json:"data_hash"`
    PublicKeys []string `json:"public_keys"`
    Signature string `json:"signature"`
  }
  r := response{}
  if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
    return nil, fmt.Errorf("invalid post batch response data: %w", err)
  }
  dataHash, err := hex.DecodeString(r.DataHash)
  if err != nil {
    return nil, fmt.Errorf("data hash is not valid hex")
  }

  publicKeys := make([]PublicKey, 0, len(r.PublicKeys))
  for i, encodedPublicKey := range r.PublicKeys {
    publicKey, err := PublicKeyFromString(encodedPublicKey)
    // FIXME: maybe we should ignore and error here.
    // A broken DAC member should not break the entire batch posting
    if err != nil {
      return nil, fmt.Errorf("invalid signer public key %v: %w", i, err)
    }
    publicKeys = append(publicKeys, publicKey)
  }

  signature, err := hex.DecodeString(r.Signature)
  if err != nil {
    return nil, fmt.Errorf("invalid signature: %w", err)
  }


  // FIXME: absolutely wrong to rely on the dataHash of the aggregator service
  // We should compute the hash locally and verify the signature against it
  isValid, mask, err := c.verifySignature(dataHash, publicKeys, signature)
  if err != nil {
    return nil, fmt.Errorf("could not verify batch signature: %w", err)
  }
  if !isValid {
    return nil, ErrInvalidBatchSignature
  }

  return &batchRef{c.addr, dataHash, signature, mask}, nil
}

func (c *client) verifySignature(
  dataHash []byte, publicKeys []PublicKey, signature []byte,
) (bool, uint64, error) {
  mask := c.keyset.ComputeMask(publicKeys)
  isValid, err := c.keyset.VerifyMessage(dataHash, signature, mask)
  return isValid, mask, err
}

func (c *client) GetBatch(dataRef []byte) ([]byte, error) {
  if len(dataRef) < 33 || dataRef[0] != DACBatchHeaderID {
    return nil, fmt.Errorf("invalid DAC batch header %x", dataRef[0])
  }

  // <       1          ><    32   ><     128    ><  8   >
  // < DACBatchHeaderID >< dataHash >< signature >< mask >
  dataHash := dataRef[1:33]

  apiUrl := *c.url
  apiUrl.Path = fmt.Sprintf("batch/%s", hex.EncodeToString(dataHash))

  fmt.Println(apiUrl.Path)
  httpClient := http.DefaultClient

  resp, err := httpClient.Get(apiUrl.String())
  if err != nil {
    return nil, fmt.Errorf("could not get batch: %w", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("invalid get batch response code: %v", resp.StatusCode)
  }

  type response struct {
    Data string `json:"data"`
  }
  r := response{}
  if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
    return nil, fmt.Errorf("invalid get batch response data: %w", err)
  }

  rawData, err := hex.DecodeString(r.Data)
  if err != nil {
    return nil, fmt.Errorf("invalid batch data: %w", err)
  }
  return rawData, nil
}
