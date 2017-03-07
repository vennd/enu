// Package sha512half implements the sha512 hash but returning only the first half
// of the sum as used by Ripple.
package sha512half

import (
	"crypto/sha512"
	"hash"
)

const (
	BlockSize = sha512.BlockSize
	Size      = sha512.Size / 2
)

type digest struct {
	hash.Hash
}

// New returns a new hash.Hash computing the half of the SHA512 checksum.
func New() hash.Hash      { return &digest{Hash: sha512.New()} }
func (*digest) Size() int { return Size }

func (d *digest) Sum(in []byte) []byte {
	// We could pass 'in' through to sha512.Sum so it appends directly
	// onto it (and avoid the allocation of h) and then reslice the
	// result, but then our result could be larger then required.
	// Using a temporary is simpler.
	h := d.Hash.Sum(nil)
	return append(in, h[:Size]...)
}

// Sum256 returns the first half of the SHA512 checksum of the data.
func Sum256(data []byte) (sum512half [Size]byte) {
	tmp := sha512.Sum512(data)
	copy(sum512half[:], tmp[:Size])
	return
}
