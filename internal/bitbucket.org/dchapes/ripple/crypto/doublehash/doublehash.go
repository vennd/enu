// Package doublehash implements double hashing via a hash.Hash.
package doublehash

import (
	"crypto"
	"crypto/sha256"
	"hash"
)

// New returns a new hash.Hash that computes a double hash.
// The hashes are specified by name, e.g. crypto.SHA256.
func New(inner, outer crypto.Hash) hash.Hash {
	return &doubleHash{inner.New(), outer.New()}
}

// NewHash returns a new hash.Hash that computes a double hash.
func NewHash(inner, outer hash.Hash) hash.Hash {
	return &doubleHash{inner, outer}
}

type doubleHash struct {
	hash.Hash
	outer hash.Hash
}

func (dbl *doubleHash) Size() int { return dbl.outer.Size() }

func (dbl *doubleHash) Sum(b []byte) []byte {
	dbl.outer.Reset()
	dbl.outer.Write(dbl.Hash.Sum(nil))
	return dbl.outer.Sum(b)
}

// SumDoubleSha256 is a convience function that returns the
// double sha256 hash of the data.
func SumDoubleSha256(data []byte) [sha256.Size]byte {
	h := sha256.Sum256(data)
	return sha256.Sum256(h[:])
}

// Avoid depending on the non-base ripemd160 package for ripemd160.Size
const ripemd160Size = 20

// SumSha256Ripemd160 is a convience function that returns the
// ripemd160 hash of the sha256 hash of the data.
// SumSha256Ripemd160 panics if the crypto.RIPEMD160 hash is not linked into the binary.
func SumSha256Ripemd160(data []byte) (sum160 [ripemd160Size]byte) {
	// As of go1.2 the ripemd160 package does not provide a Sum160()
	// function so we need to use it's hash.Hash implementation.
	// This also avoids us having to import golang.org/x/crypto/ripemd160
	ripemd160 := crypto.RIPEMD160.New()
	h := sha256.Sum256(data)
	ripemd160.Write(h[:])
	sum := ripemd160.Sum(nil)
	copy(sum160[:], sum[:ripemd160Size])
	return
}
