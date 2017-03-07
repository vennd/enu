package rkey

import (
	"github.com/vennd/enu/internal/bitbucket.org/dchapes/ripple/crypto/doublehash"
	"bytes"
	"encoding"
	"errors"
	"github.com/vennd/enu/internal/github.com/spearson78/guardian/encoding/base58"
)

const encodeRipple = "rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz"

var rippleEncoding = base58.NewEncoding(encodeRipple)

// Key is ???
// TODO(dchapes): should this even be exported? Is there a better/cleaner
// way to implement this.
type Key interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	version() version
	raw() []byte
	setraw([]byte) error
}

// encode returns a Ripple base58 encoded version of k.
func encode(k Key) ([]byte, error) {
	ver := k.version()
	psize := ver.PayloadSize()
	tsize := 1 + psize + 4
	raw := k.raw()
	if len(raw) > psize {
		return nil, errors.New("tbd")
	}
	tmp := make([]byte, tsize)
	tmp[0] = byte(ver)
	copy(tmp[len(tmp)-4-len(raw):], raw)
	sum := doublehash.SumDoubleSha256(tmp[:1+psize])
	copy(tmp[1+psize:], sum[:4])
	return rippleEncoding.Encode(tmp)
}

// padbytes zero pads b on the left so that the length is >= padlen.
// TODO(dchapes): move this
func padbytes(b []byte, padlen int) []byte {
	n := len(b)
	if n >= padlen {
		return b
	}
	tmp := make([]byte, padlen)
	copy(tmp[padlen-n:], b)
	return tmp
}

// New returns a new Key decoded from a Ripple base58 encoded text.
func New(text string) (Key, error) {
	return decode(nil, []byte(text))
}

// decodeInto decodes a Ripple base58 encoded text into k.
// In addition to decoding errors, if the encoding is a
// non-matching version/type an error is returned.
func decodeInto(k Key, text []byte) error {
	_, err := decode(k, text)
	return err
}

// decode decodes a Ripple base58 encoded text.
// If k is non-nil then the encoded version/type must match and the value is
// decoded into the existing k.
// If k is nil then a new Key of the version/type appropriate to the
// encoded text is initialised.
func decode(k Key, text []byte) (Key, error) {
	tmp, err := rippleEncoding.Decode(text)
	if err != nil {
		return nil, err
	}
	v := version(tmp[0])
	if !v.valid() {
		return nil, errors.New("rkey decode: bad version") // XXX
	}
	if k != nil && v != k.version() {
		return nil, errors.New("rkey decode: version mismatch") // XXX
	}
	if k == nil {
		k = v.newKey()
		if k == nil {
			return nil, errors.New("rkey decode: version not supported") // XXX
		}
	}
	psize := v.PayloadSize()
	tsize := 1 + psize + 4
	tmp = padbytes(tmp, tsize) // XXX
	if len(tmp) != tsize {
		return nil, errors.New("rkey decode: bad size") // XXX
	}
	sum := doublehash.SumDoubleSha256(tmp[:1+psize])
	if !bytes.Equal(tmp[1+psize:], sum[:4]) {
		return nil, errors.New("rkey decode: bad checksum bytes") // XXX
	}
	err = k.setraw(tmp[1 : 1+v.PayloadSize()])
	return k, err
}

// XXX

func (v version) newKey() Key {
	switch v {
	case verFamilySeed:
		return new(FamilySeed)
	case verFamilyGenerator:
		return new(PublicGenerator)
	case verAcctPrivate:
		return new(AcctPrivateKey)
	case verAcctPublic:
		return new(AcctPublicKey)
	}
	return nil
}

func (*FamilySeed) version() version      { return verFamilySeed }
func (*PublicGenerator) version() version { return verFamilyGenerator }
func (*AcctPrivateKey) version() version  { return verAcctPrivate }
func (*AcctPublicKey) version() version   { return verAcctPublic }
func (*AccountId) version() version       { return verAcctId }

func (s *FamilySeed) MarshalText() ([]byte, error)      { return encode(s) }
func (f *PublicGenerator) MarshalText() ([]byte, error) { return encode(f) }
func (p *AcctPrivateKey) MarshalText() ([]byte, error)  { return encode(p) }
func (a *AcctPublicKey) MarshalText() ([]byte, error)   { return encode(a) }
func (r *AccountId) MarshalText() ([]byte, error)       { return encode(r) }

func (s *FamilySeed) UnmarshalText(text []byte) error      { return decodeInto(s, text) }
func (f *PublicGenerator) UnmarshalText(text []byte) error { return decodeInto(f, text) }
func (p *AcctPrivateKey) UnmarshalText(text []byte) error  { return decodeInto(p, text) }
func (a *AcctPublicKey) UnmarshalText(text []byte) error   { return decodeInto(a, text) }
func (r *AccountId) UnmarshalText(text []byte) error       { return decodeInto(r, text) }

func NewFamilySeed(text string) (*FamilySeed, error) {
	s := new(FamilySeed)
	return s, decodeInto(s, []byte(text))
}
func NewPublicGenerator(text string) (*PublicGenerator, error) {
	f := new(PublicGenerator)
	return f, decodeInto(f, []byte(text))
}
func NewAcctPrivateKey(text string) (*AcctPrivateKey, error) {
	p := new(AcctPrivateKey)
	return p, decodeInto(p, []byte(text))
}
func NewAcctPublicKey(text string) (*AcctPublicKey, error) {
	a := new(AcctPublicKey)
	return a, decodeInto(a, []byte(text))
}
func NewAccountId(text string) (*AccountId, error) {
	r := new(AccountId)
	return r, decodeInto(r, []byte(text))
}
