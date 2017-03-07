package rkey

import (
	"testing"
)

func TestPublicGenerator(t *testing.T) {
	checkDepends(t, "FamilySeed", "PrivateGenerator")
	testResults["PublicGenerator"] = false
	for i, d := range testKeyData {
		//t.Log(i, d.secret, d.publicGen)

		f, err := NewPublicGenerator(d.publicGen)
		if err != nil {
			t.Errorf("%2d: NewPublicGenerator() failed: %v", i, err)
			continue
		}
		if g, err := f.MarshalText(); err != nil {
			t.Errorf("%2d: PublicGenerator.MarshalText() failed: %v", i, err)
		} else if string(g) != d.publicGen {
			t.Errorf("%2d: PublicGenerator.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.publicGen)
		}

		s, err := NewFamilySeed(d.secret)
		if err != nil {
			t.Skip("Skipping due to NewFamilySeed failure:", err)
		}
		f2 := &s.PrivateGenerator.PublicGenerator
		if f2.X.Cmp(f.X) != 0 {
			t.Errorf("%2d: PublicGenerator\n\t  result X=%#x\n\texpected X=%#x",
				i, f2.X, f.X)
		}
		if f2.Y.Cmp(f.Y) != 0 {
			t.Errorf("%2d: PublicGenerator\n\t  result Y=%#x\n\texpected Y=%#x",
				i, f2.Y, f.Y)
		}
		// XXX
		if g, err := f2.MarshalText(); err != nil {
			t.Errorf("%2d: PublicGenerator.MarshalText() failed: %v", i, err)
		} else if string(g) != d.publicGen {
			t.Errorf("%2d: PublicGenerator.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.publicGen)
		}
	}
	testResults["PublicGenerator"] = !t.Failed()
}
