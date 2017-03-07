package sha512half

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"hash"
	"testing"
)

func TestSize(t *testing.T) {
	if g, e := Size, sha512.Size/2; g != e {
		t.Errorf("const sha512half.Size = %d, expected %d", g, e)
	}
	var h hash.Hash = New()
	if g, e := h.Size(), sha512.Size/2; g != e {
		t.Errorf("sha512half.New().Size() = %d, expected %d", g, e)
	}
}

// SHA512 half of a single zero byte: B8244D028981D693AF7B456AF8EFA4CAD63D282E19FF14942C246E50D9351D22
// SHA512 half of 100,000 as a 32-bit integer in big-endian form: 8EEE2EA9E7F93AB0D9E66EE4CE696D6824922167784EC7F340B3567377B1CE64

func TestSum(t *testing.T) {
	data := []byte{0}
	h := New()
	h.Write(data)
	sum := h.Sum(nil)
	if g, e := len(sum), sha512.Size/2; g != e {
		t.Errorf("sha512half.New().Sum(nil) len = %d, expected %d", g, e)
	} else {
		hfull := sha512.New()
		hfull.Write(data)
		sumfull := hfull.Sum(nil)
		if g, e := sum, sumfull[:sha512.Size/2]; !bytes.Equal(g, e) {
			t.Errorf("sha512half.New().Sum(nil) gave 0x%x, expected 0x%x", g, e)
		}
		e := "B8244D028981D693AF7B456AF8EFA4CAD63D282E19FF14942C246E50D9351D22"
		if g := fmt.Sprintf("%X", sum); g != e {
			t.Errorf("sha512half.New().Sum(nil) gave %q, expected %q", g, e)
		}
	}
}

func TestSum256(t *testing.T) {
	sum := Sum256(nil)
	if g, e := len(sum), sha512.Size/2; g != e {
		t.Errorf("sha512half.Sum256(nil) len = %d, expected %d", g, e)
	} else {
		sumfull := sha512.Sum512(nil)
		if g, e := sum, sumfull[:sha512.Size/2]; !bytes.Equal(g[:], e) {
			t.Errorf("sha512half.Sum256(nil) gave 0x%x, expected 0x%x", g, e)
		}
	}
}

func BenchmarkSum(b *testing.B) {
	var buf []byte
	data := []byte("Some random test data")
	h := New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Reset()
		h.Write(data)
		buf = h.Sum(buf[0:])
	}
}
func BenchmarkSum256(b *testing.B) {
	data := []byte("Some random test data")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Sum256(data)
	}
}
