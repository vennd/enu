package rkey

import (
	"github.com/vennd/enu/internal/bitbucket.org/dchapes/ripple/crypto/sha512half"
	"crypto/ecdsa"
	"encoding/binary"
	"math/big"
)

// A PublicGenerator, also called the root/master public key or family
// generator, is effectively a ECDSA public key used to make the
// individual AcctPublicKey's within the account family.
//
// The PublicGenerator is so named because it generates the public
// keys, not because it should be made public.
//
// With the public generator, anyone can determine which accounts are in the family.
//
// Caution: With the public generator and any one private key, the
// private generator can be determined. For this reason, export of
// individual private keys should not be allowed when the accounts are
// part of a family.
//
// Its Ripple base58 encoding starts with `f`
type PublicGenerator struct {
	X, Y *big.Int
}

func (g *PublicGenerator) hashGenerate(idx uint32) *big.Int {
	k := new(big.Int)
	hash := sha512half.New()
	var subseq uint32
	var sum []byte
	raw := g.raw()
	for k.BitLen() == 0 || k.Cmp(curve.N) >= 0 {
		hash.Reset()
		hash.Write(raw)
		binary.Write(hash, binary.BigEndian, idx)
		binary.Write(hash, binary.BigEndian, subseq)
		subseq++
		sum = hash.Sum(sum[:0])
		k = k.SetBytes(sum)
	}
	return k
}

// Generate is used to generate the AcctPublicKey for sequence idx in
// this account family.
func (g *PublicGenerator) Generate(idx uint32) *AcctPublicKey {
	k := g.hashGenerate(idx)
	x2, y2 := curve.ScalarBaseMult(k.Bytes())

	x, y := curve.Add(g.X, g.Y, x2, y2)
	//cp := curve.CompressPoint(x, y)
	key := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}
	return (*AcctPublicKey)(key)
}

func (g *PublicGenerator) raw() []byte { return curve.CompressPoint(g.X, g.Y) }

func (g *PublicGenerator) setraw(raw []byte) error {
	x, y, err := curve.DecompressPoint(raw)
	if err == nil {
		g.X, g.Y = x, y
	}
	return err
}
