package rkey

import (
	"github.com/vennd/enu/internal/bitbucket.org/dchapes/ripple/crypto/sha512half"
	"crypto/ecdsa"
	"encoding/binary"
	"github.com/vennd/enu/internal/github.com/sour-is/koblitz/kelliptic"
	"math/big"
)

// All Ripple operations are the secp256k1 Koblitz elliptic curve.
var curve = kelliptic.S256()

// A PrivateGenerator, also called the root private key or master
// private key, is effectively a ECDSA private key used to make the
// PrivateGenerator (the matching ECDSA public key) and the
// individual AcctPrivateKey's within the account family.
//
// With the private generator, all the private keys can be determined.
type PrivateGenerator struct {
	D *big.Int
	PublicGenerator
}

// Not needed, there is no Ripple base58 encoding of these.
//func (g *PrivateGenerator) raw() []byte { return g.D.Bytes() }

// Generate is used to generate the AcctPrivateKey for sequence idx in
// this account family.
func (g *PrivateGenerator) Generate(idx uint32) *AcctPrivateKey {
	k := g.PublicGenerator.hashGenerate(idx)
	k.Add(k, g.D)
	k.Mod(k, curve.N)
	x, y := curve.ScalarBaseMult(k.Bytes())
	key := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: k,
	}
	return (*AcctPrivateKey)(key)
}

// init sets this generator from the from provided seed.
// The seed is from a FamilySeed and is used to find a value for D, X, Y;
// effectively a ECDSA private/public key pair.
func (g *PrivateGenerator) init(seed *big.Int) {
	k := new(big.Int)
	hash := sha512half.New()
	var seq uint32
	var sum []byte
	raw := seed.Bytes()
	for k.BitLen() == 0 || k.Cmp(curve.N) >= 0 {
		hash.Reset()
		hash.Write(raw)
		binary.Write(hash, binary.BigEndian, seq)
		seq++
		sum = hash.Sum(sum[:0])
		k = k.SetBytes(sum)
	}
	g.setraw(k.Bytes())
}

func (g *PrivateGenerator) setraw(raw []byte) error {
	if g.D == nil {
		g.D = new(big.Int)
	}
	g.D.SetBytes(raw)
	x, y := curve.ScalarBaseMult(raw)
	g.PublicGenerator.X = x
	g.PublicGenerator.Y = y
	return nil
}
