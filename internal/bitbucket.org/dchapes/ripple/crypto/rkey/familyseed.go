package rkey

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var maxSeed, _ = new(big.Int).SetString("100000000000000000000000000000000", 16)

// A FamilySeed, also called the Ripple secret, is a 128 bit random number used
// to generate all the private keys, public keys, and Ripple addresses within
// the account family.
//
// Its Ripple base58 encoding starts with `s`.
type FamilySeed struct {
	Seed *big.Int
	PrivateGenerator
}

// GenerateSeed uses crypto/rand to generate a new random FamilySeed.
func GenerateSeed() (*FamilySeed, error) {
	i, err := rand.Int(rand.Reader, maxSeed)
	if err != nil {
		return nil, err
	}
	return NewSeed(i)
}

func (fs *FamilySeed) raw() []byte { return fs.Seed.Bytes() }

func (fs *FamilySeed) setraw(raw []byte) error {
	if fs.Seed == nil {
		fs.Seed = new(big.Int)
	}
	fs.Seed.SetBytes(raw)
	fs.PrivateGenerator.init(fs.Seed)
	return nil
}

// NewSeed uses the provided Int, which must be <=128 bits long, as a FamilySeed.
func NewSeed(i *big.Int) (*FamilySeed, error) {
	if i.Cmp(maxSeed) >= 0 {
		return nil, errors.New("tbd too big")
	} else if i.Sign() < 0 {
		return nil, errors.New("tbd too small")
	}
	fs := &FamilySeed{Seed: i}
	fs.PrivateGenerator.init(i)
	return fs, nil
}
