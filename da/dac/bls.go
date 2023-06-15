package dac

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	bls "github.com/ethereum/go-ethereum/crypto/bls12381"
)

var (
  ErrKeySetTooLarge = errors.New("key set size is 64 maximum")
)

const (
  MaxKeySetSize = 64
  AllKeysMask = uint64(0)
)

// PublicKey represenets a serializable public key
type PublicKey struct {
  p *bls.PointG1
}

func PublicKeyFromString(s string) (PublicKey, error) {
  bytes, err := hex.DecodeString(s)
  if err != nil {
    return PublicKey{}, fmt.Errorf("could not decode hex string: %w", err)
  }

  g1 := bls.NewG1()
  p, err := g1.FromBytes(bytes)
  if err != nil {
    return PublicKey{}, fmt.Errorf("could not decode point: %w", err)
  }
  return PublicKey{p}, nil
}

func (p PublicKey) VerifyMessage(message []byte, signature []byte) (bool, error) {
  return verify(message, signature, p)
}

func (p PublicKey) ToBytes() []byte {
  g1 := bls.NewG1()
  return g1.ToBytes(p.p)
}


// Signer represents a BLS signer over G2
type Signer struct {
  b *big.Int
}

func (s Signer) GetPublicKey() PublicKey {
  g1 := bls.NewG1()
  g := g1.One()

  p := g1.MulScalar(g1.New(), g, s.b)
  return PublicKey{p}
}

// Signature represents a serializable signature
type Signature struct {
  p *bls.PointG2
}

func NewSignature(b []byte) (Signature, error) {
  g2 := bls.NewG2()
  sigPoint, err := g2.FromBytes(b)
  if err != nil {
    return Signature{}, err
  }
  return Signature{sigPoint}, nil
}

func (s Signature) ToBytes() []byte {
  g2 := bls.NewG2()
  return g2.ToBytes(s.p)
}

func NewSigner(privateKey string) (Signer, error) {
  key, err := hexutil.DecodeBig(privateKey)
  if err != nil {
    return Signer{}, err
  }
  return Signer{key}, nil
}

func (s Signer) Sign(message []byte) (Signature, error) {
  msgPoint, err := hashToG2(message)
  if err != nil {
    return Signature{}, err
  }
  g2 := bls.NewG2()
  result := g2.MulScalar(g2.New(), msgPoint, s.b)
  return Signature{result}, nil
}

// KeySet represents a set of public keys over G1
type KeySet []PublicKey

// NewKeySet creates a key set
// It is the caller's duty to verify that private keys corresponding to the given public keys exist and that every key in the set is unique
func NewKeySet(uncompressedKeys [][]byte) (KeySet, error) {
  if len(uncompressedKeys) > MaxKeySetSize {
    return nil, ErrKeySetTooLarge
  }

  set := make([]PublicKey, len(uncompressedKeys))

  g1 := bls.NewG1()
  for index, uncompressedKey := range uncompressedKeys {
    pubKey, err := g1.FromBytes(uncompressedKey)
    if err != nil {
      return nil, fmt.Errorf("invalid %v public key %w", index, err)
    }
    set[index] = PublicKey{pubKey}
  }

  return KeySet(set), nil
}

func NewKeySetFromString(uncompressedKeys []string) (KeySet, error) {
  encoded := make([][]byte, len(uncompressedKeys))

  for i, str := range uncompressedKeys {
    var err error
    encoded[i], err = hex.DecodeString(str)
    if err != nil {
      return nil, fmt.Errorf("could not decode key %v of the set: %w", i, err)
    }
  }
  return NewKeySet(encoded)
}

// ComputeMask computes a uint64 mask of the keyset. i'th bit is set iif
// the i'th public key of the keyset is in the given array of keys
func (s KeySet) ComputeMask(keys []PublicKey) uint64 {
  mask := uint64(0)

  input := map[string]struct{}{}
  for _, inputKey := range keys {
    input[string(inputKey.ToBytes())] = struct{}{}
  }

  for i, key := range s {
    if _, isKeyInInput := input[string(key.ToBytes())]; isKeyInInput {
      mask |= 1 << i
    }
  }

  return mask
}

func (s KeySet) Aggregate(mask uint64) PublicKey {
  g1 := bls.NewG1()

  aggKey := g1.Zero()

  if mask == 0 {
    mask = math.MaxUint64
  }
  for i, pubKey := range s {
    if (mask >> i) & 0x1 == 0 {
      continue
    }
    g1.Add(aggKey, aggKey, pubKey.p)
  }
  return PublicKey{aggKey}
}

func (s KeySet) VerifyMessage(message []byte, signature []byte, mask uint64) (bool, error) {
  aggKey := s.Aggregate(mask)
  return verify(message, signature, aggKey)
}

func verify(message []byte, signature []byte, key PublicKey) (bool, error) {
  msgPoint, err := hashToG2(message)
  if err != nil {
    return false, fmt.Errorf("could not map message to curve: %w", err)
  }

  sig, err := NewSignature(signature)
  if err != nil {
    return false, fmt.Errorf("could not parse signature: %w", err)
  }

  // e(P, H(m)) == e(G, S)`
  engine := bls.NewPairingEngine()
  engine.Reset()
  engine.AddPair(key.p, msgPoint)
	leftSide := engine.Result()
  engine.AddPair(engine.G1.One(), sig.p)
	rightSide := engine.Result()
	return leftSide.Equal(rightSide), nil

}

func hashToG2(message []byte) (*bls.PointG2, error) {
  // hash so that we don't sign an arbitrary value
  hashed := crypto.Keccak256(message)

  g2 := bls.NewG2()
  padding := [64]byte{}
  msgPoint, err := g2.MapToCurve(append(padding[:], hashed...))
  if err != nil {
    return nil, fmt.Errorf("could not map message to curve: %w", err)
  }
  return msgPoint, nil
}

func AggregateSignatures(signatures []Signature) Signature {
  g2 := bls.NewG2()
  agg := g2.Zero()

  for _, signature := range signatures {
    g2.Add(agg, agg, signature.p)
  }

  return Signature{agg}
}
