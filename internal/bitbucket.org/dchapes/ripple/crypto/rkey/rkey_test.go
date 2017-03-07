package rkey

import (
	//crand "crypto/rand"
	"math/big"
	"testing"
)

func BenchmarkGeneration(b *testing.B) {
	//r, err := crand.Int(crand.Reader, maxSeed)
	//if err != nil {
	//	b.Error("Failed to get random value:", err)
	//}
	r, _ := new(big.Int).SetString("0xF0E0D0C0B0A090807060504030201000", 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seed, err := NewSeed(r)
		if err != nil {
			b.Error("GenerateSeed failed:", err)
		}
		pubkey := seed.PrivateGenerator.PublicGenerator.Generate(0)
		addr := pubkey.Address()
		_ = addr
	}
}
