package rkey

import (
	"testing"
)

func TestPrivateKey(t *testing.T) {
	checkDepends(t, "FamilySeed", "PrivateGenerator")
	testResults["PrivateKey"] = false
	for i, d := range testKeyData {
		//t.Log(i, d.secret, d.privateKey0)

		p, err := NewAcctPrivateKey(d.privateKey0)
		if err != nil {
			t.Errorf("%2d: NewAcctPrivateKey() failed: %v", i, err)
			continue
		}
		if g, err := p.MarshalText(); err != nil {
			t.Errorf("%2d: PrivateKey.MarshalText() failed: %v", i, err)
		} else if string(g) != d.privateKey0 {
			t.Errorf("%2d: PrivateKey.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.privateKey0)
		}

		s, err := NewFamilySeed(d.secret)
		if err != nil {
			t.Skip("Skipping due to NewFamilySeed failure:", err)
		}
		p2 := s.PrivateGenerator.Generate(0)
		if p2.D.Cmp(p.D) != 0 {
			t.Errorf("%2d: PrivateKey.MarshalText()\n\treturned D=%#x\n\texpected D=%#x",
				i, p2.D, p.D)
		}
		// XXX
		if g, err := p2.MarshalText(); err != nil {
			t.Errorf("%2d: PrivateKey.MarshalText() failed: %v", i, err)
		} else if string(g) != d.privateKey0 {
			t.Errorf("%2d: PrivateKey.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.privateKey0)
		}

		a := (*AcctPublicKey)(&p.PublicKey)
		a2 := (*AcctPublicKey)(&p2.PublicKey)
		if a2.X.Cmp(a.X) != 0 {
			t.Errorf("%2d: PublicKey\n\t  result X=%#x\n\texpected X=%#x",
				i, a2.X, a.X)
		}
		if a2.Y.Cmp(a.Y) != 0 {
			t.Errorf("%2d: PublicKey\n\t  result Y=%#x\n\texpected Y=%#x",
				i, a2.Y, a.Y)
		}

		if g, err := a.MarshalText(); err != nil {
			t.Errorf("%2d: PublicKey.MarshalText() failed: %v", i, err)
		} else if string(g) != d.publicKey0 {
			t.Errorf("%2d: PublicKey.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.publicKey0)
		}
	}
	testResults["PrivateKey"] = !t.Failed()
}
