package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ethereum-optimism/optimism/da/dac"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/mux"
)

type member struct {
  storage Storage
  signer dac.Signer
  // just so we don't recompute it too often
  publicKey dac.PublicKey
}

func newMember(storage Storage, privateKey string) (*member, error) {
  signer, err := dac.NewSigner(privateKey)
  if err != nil {
    return nil, fmt.Errorf("could not instanciate the signer: %w", err)
  }
  return &member{
    storage: storage,
    signer: signer,
    publicKey: signer.GetPublicKey(),
  }, nil
}

func (m *member) handleGet(w http.ResponseWriter, req *http.Request) {
  dataHash := mux.Vars(req)["dataHash"]

  log.Info("retrieving batch", "data_hash", dataHash)

  data, err := m.storage.Fetch(dataHash)
  if err != nil {
    if err == ErrNotFound {
      w.WriteHeader(http.StatusNotFound)
      return
    }
    log.Warn("could not fetch batch", "err", err, "data_hash", dataHash)
    w.WriteHeader(http.StatusInternalServerError)
  }
  defer data.Close()

  w.Header().Set("Content-Type", "text/plain")
  if written, err := io.Copy(hex.NewEncoder(w), data); err != nil {
    log.Warn("could not write batch", "err", err, "written", written)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
}

func (m *member) handlePost(w http.ResponseWriter, req *http.Request) {
  type request struct {
    Data string `json:"data"`
  }
  defer req.Body.Close()

  payload := &request{}
  if err := json.NewDecoder(req.Body).Decode(payload); err != nil || len(payload.Data) == 0 {
    log.Info("payload is not valid json", "encoded_data_len", len(payload.Data))
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  data, err := hex.DecodeString(payload.Data)
  if err != nil {
    log.Info("batch is not valid hex", "encoded_data_len", len(payload.Data))
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  dataHash := crypto.Keccak256(data)
  dataHashHex := hex.EncodeToString(dataHash)
  if err := m.storage.Store(dataHashHex, bytes.NewReader(data)); err != nil {
    log.Error("could not store batch", "err", err, "data_len", len(data))
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  signature, err := m.signer.Sign(dataHash)
  if err != nil {
    log.Error("could not sign batch", "err", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  type response struct {
    DataHash string `json:"data_hash"`
    Signature string `json:"signature"`
    // PublicKey is purely informational. It should not serve as truth, it is just a convenience
    // for the aggregator to build a bitmap. The aggregator has to check a proof of ownership of
    // the private keys
    PublicKey string `json:"public_key"` 
  }

  json.NewEncoder(w).Encode(response{
    DataHash: dataHashHex,
    Signature: hex.EncodeToString(signature.ToBytes()),
    PublicKey: hex.EncodeToString(m.publicKey.ToBytes()),
  })
}
