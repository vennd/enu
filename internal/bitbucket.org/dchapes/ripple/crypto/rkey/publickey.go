package rkey

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
)

// An AcctPublicKey is a specific ECDSA public key within an account family.
//
// Its Ripple base58 encoding starts with `a`.
type AcctPublicKey ecdsa.PublicKey

func (p *AcctPublicKey) raw() []byte { return curve.CompressPoint(p.X, p.Y) }

func (p *AcctPublicKey) setraw(raw []byte) error {
	x, y, err := curve.DecompressPoint(raw)
	if err == nil {
		p.Curve = curve
		p.X, p.Y = x, y
	}
	return err
}

// MarshalJSON implements the json.Marshaler interface.
// Ripple uses a DER style encoding in JSON Ripple transactions rather than
// base58 encoding.
func (p *AcctPublicKey) MarshalJSON() ([]byte, error) {
	r := p.raw()
	text := make([]byte, hex.EncodedLen(len(r))+2)
	text[0] = '"'
	hex.Encode(text[1:], r)
	text[len(text)-1] = '"'
	return bytes.ToUpper(text), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Ripple uses a DER style encoding in JSON Ripple transactions rather than
// base58 encoding.
func (p *AcctPublicKey) UnmarshalJSON(in []byte) error {
	in = bytes.Trim(in, "\"")
	r := make([]byte, hex.DecodedLen(len(in)))
	if _, err := hex.Decode(r, in); err != nil {
		return err
	}
	return p.setraw(r)
}

// AccountId returns the account id (aka Ripple address) for this public key.
//
// TODO(dchapes): remove?
func (p *AcctPublicKey) AccountId() *AccountId {
	return accountIdHash(p.raw())
}

// Address returns the Ripple address of this public key.
func (p *AcctPublicKey) Address() string {
	addr, err := p.AccountId().MarshalText() // XXX
	if err != nil {
		panic(err) // XXX
	}
	return string(addr)
}
