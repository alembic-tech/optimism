package dac_test

import (
	"fmt"
	"testing"

	"github.com/ethereum-optimism/optimism/da/dac"
)

func TestBLS(t *testing.T) {
}

func TestBLSPublicKey(t *testing.T) {
  key := "0000000000000000000000000000000010c36f69c5f73a0ae95fa1768e68a58973d0a3a61f1e9bf889050217388ebb24c57341fb5528b8f2b6138d5149d88c61000000000000000000000000000000001003f241a22da86d76e15cdfdc06d6ea86845d5f662e3209044716add654d98aa6a9c99632b2d647ac280e36d9da5756"

  pubKey, err := dac.PublicKeyFromString(key)
  if err != nil {
    t.Errorf("pubKey: got an error: %v", err)
  }


  keyset, err := dac.NewKeySetFromString([]string{key})
  if err != nil {
    t.Errorf("keyset: got an error: %v", err)
  }

  fmt.Println(pubKey, keyset)
}
