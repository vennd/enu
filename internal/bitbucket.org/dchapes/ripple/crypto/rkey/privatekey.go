package rkey

import (
	"crypto/ecdsa"
	"math/big"
)

// An AcctPrivateKey is a specific ECDSA private key within an account family.
//
// Its Ripple base58 encoding starts with `p`.
type AcctPrivateKey ecdsa.PrivateKey

func (p *AcctPrivateKey) raw() []byte { return p.D.Bytes() }

func (p *AcctPrivateKey) setraw(raw []byte) error {
	if p.D == nil {
		p.D = new(big.Int)
	}
	p.D.SetBytes(raw)
	p.PublicKey.Curve = curve
	p.PublicKey.X, p.PublicKey.Y = curve.ScalarBaseMult(p.D.Bytes())
	return nil
}
