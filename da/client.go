package da

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Client struct {
  url *url.URL
}

func NewClient(apiUrl string) *Client {
  parsed, err := url.Parse(apiUrl)
  if err != nil {
    panic(fmt.Errorf("invalid DA url: %w", err))
  }
  return &Client{parsed}
}

func (c *Client) PostBatch(data []byte) ([]byte, error) {

  fmt.Println("----------data in PostBatch function", data)
  apiUrl := *c.url
  apiUrl.Path = "batch"

  httpClient := http.DefaultClient

  type payload struct {
    Data string `json:"data"`
  }
  p := payload{
    Data: hexutil.Encode(data),
  }
  encoded, _ := json.Marshal(p)
  resp, err := httpClient.Post(apiUrl.String(), "application/json", bytes.NewReader(encoded))
  if err != nil {
    return []byte{}, fmt.Errorf("could not post batch: %w", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    return []byte{}, fmt.Errorf("invalid post batch response code: %v", resp.StatusCode)
  }

  type response struct {
    DataHash []byte `json:"id"`
  }
  r := response{}
  if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
    return []byte{}, fmt.Errorf("invalid post batch response data: %w", err)
  }

  return r.DataHash, nil
}

func (c *Client) GetBatch(dataHash []byte) ([]byte, error) {
  apiUrl := *c.url
  apiUrl.Path = fmt.Sprintf("batch/%s", dataHash)

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
    Data []byte `json:"data"`
  }
  r := response{}
  if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
    return nil, fmt.Errorf("invalid get batch response data: %w", err)
  }

  return r.Data, nil
}
