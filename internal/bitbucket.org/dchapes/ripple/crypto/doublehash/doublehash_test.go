package doublehash

import (
	"crypto"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"testing"

	"github.com/vennd/enu/internal/golang.org/x/crypto/ripemd160"
)

func TestSize(t *testing.T) {
	s256 := sha256.New()
	s512 := sha512.New()

	if g, e := ripemd160Size, ripemd160.Size; g != e {
		t.Errorf("ripemd160Size = %d, expected %d", g, e)
	}

	h := New(crypto.SHA512, crypto.SHA256)
	if g, e := h.Size(), crypto.SHA256.Size(); g != e {
		t.Errorf("doublehash.New(crypto.SHA512, crypto.SHA256).Size() = %d, expected %d", g, e)
	}
	if g, e := h.BlockSize(), s512.BlockSize(); g != e {
		t.Errorf("doublehash.New(crypto.SHA512, crypto.SHA256).BlockSize() = %d, expected %d", g, e)
	}
	h = NewHash(s512, s256)
	if g, e := h.Size(), crypto.SHA256.Size(); g != e {
		t.Errorf("doublehash.New(crypto.SHA512, crypto.SHA256).Size() = %d, expected %d", g, e)
	}
	if g, e := h.BlockSize(), s512.BlockSize(); g != e {
		t.Errorf("doublehash.New(crypto.SHA512, crypto.SHA256).BlockSize() = %d, expected %d", g, e)
	}

	h = New(crypto.SHA256, crypto.SHA512)
	if g, e := h.Size(), crypto.SHA512.Size(); g != e {
		t.Errorf("doublehash.New(crypto.SHA256, crypto.SHA512).Size() = %d, expected %d", g, e)
	}
	if g, e := h.BlockSize(), s256.BlockSize(); g != e {
		t.Errorf("doublehash.New(crypto.SHA256, crypto.SHA512).BlockSize() = %d, expected %d", g, e)
	}
	h = NewHash(s256, s512)
	if g, e := h.Size(), crypto.SHA512.Size(); g != e {
		t.Errorf("doublehash.New(crypto.SHA256, crypto.SHA512).Size() = %d, expected %d", g, e)
	}
	if g, e := h.BlockSize(), s256.BlockSize(); g != e {
		t.Errorf("doublehash.New(crypto.SHA256, crypto.SHA512).BlockSize() = %d, expected %d", g, e)
	}
}

var data = []byte("Some random test data")

func TestSum(t *testing.T) {
	h := New(crypto.SHA512, crypto.SHA256)
	h.Write(data)
	sum := []byte("HEAD\000")
	sum = h.Sum(sum)
	e := "4845414400F462F54EA31DD62E26E49678FA45680411285C2B363F93D0FB718152953EBA7A"
	if g := fmt.Sprintf("%X", sum); g != e {
		t.Errorf("Sum()\n\treturned %q,\n\texpected %q", g, e)
	}
}

func TestDoubleSha256(t *testing.T) {
	sum := SumDoubleSha256(data)
	e := "3ED4E653B03D4BCC92734CBADCAE452D7B58F4C4AC53AEA085BAA27AD47FB833"
	if g := fmt.Sprintf("%X", sum); g != e {
		t.Errorf("SumDoubleSha256()\n\treturned %q,\n\texpected %q", g, e)
	}
}

func TestSumSha256Ripemd160(t *testing.T) {
	sum := SumSha256Ripemd160(data)
	e := "6151CC5625FB9F0F31F7D38518D1388D260E56D4"
	if g := fmt.Sprintf("%X", sum); g != e {
		t.Errorf("SumSha256Ripemd160()\n\treturned %q,\n\texpected %q", g, e)
	}
}

func BenchmarkSumSha256Ripemd160(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SumSha256Ripemd160(data)
	}
}
