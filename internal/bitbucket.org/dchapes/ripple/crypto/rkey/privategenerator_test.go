package rkey

import (
	"math/big"
	"testing"
)

func TestPrivateGenerator(t *testing.T) {
	checkDepends(t, "FamilySeed")
	testResults["PrivateGenerator"] = false
	for i, d := range testKeyData {
		//t.Log(i, d.secret, d.privateGen)
		s, err := NewFamilySeed(d.secret)
		if err != nil {
			t.Skip("Skipping due to NewFamilySeed failure:", err)
		}
		e, _ := new(big.Int).SetString(d.privateGen, 0)
		if g := s.PrivateGenerator.D; g.Cmp(e) != 0 {
			t.Errorf("%2d: PrivateGenerator\n\t  result %#x,\n\texpected %#x",
				i, g, e)
		}
	}
	testResults["PrivateGenerator"] = !t.Failed()
}
