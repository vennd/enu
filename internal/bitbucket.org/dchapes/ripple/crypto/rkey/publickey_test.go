package rkey

import (
	"math/big"
	"testing"
)

func TestAccountIdEncode(t *testing.T) {
	acctId := &AccountId{Id: new(big.Int)}
	for i, d := range []struct{ n, r string }{
		{"0", "rrrrrrrrrrrrrrrrrrrrrhoLvTp"},
		{"1", "rrrrrrrrrrrrrrrrrrrrBZbvji"},
		{"0xC3B0D7B9A232B393AB7E393BCB078114FDDABCAA",
			"rJq5ce8cdbWBsysXx32rvLMV6DUxMwruMT"},
		{"0xDBDCF3B9F064802968137945CEA10726AF313E19",
			"rMsXgpgs6kuisJyy7nsNWmsJyfw8ydYh3w"},
		{"0x278C4AC8FEAB993E85BBA8FF8401E7C37E062C56",
			"rhcfR9Cg98qCxHpCcPBmMonbDBXo84wyTn"},
	} {
		//t.Log(i, d)
		acctId.Id.SetString(d.n, 0)
		if g, err := acctId.MarshalText(); err != nil {
			t.Errorf("%2d: AccountId.MarshalText() failed: %v", i, err)
		} else if string(g) != d.r {
			t.Errorf("%2d: AccountId.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.r)
		}

		acctId2 := new(AccountId)
		if err := acctId2.UnmarshalText([]byte(d.r)); err != nil {
			t.Errorf("%2d: AccountId.UnmarshalText(%q) failed: %v",
				i, d.r, err)
		} else if acctId2.Id.Cmp(acctId.Id) != 0 {
			t.Errorf("%2d: AccountId.UnmarshalText\n\tresulted %#x,\n\texpected %#x",
				i, acctId2.Id, acctId.Id)
		}
	}
}

func TestPublicKey(t *testing.T) {
	checkDepends(t, "FamilySeed", "PrivateGenerator", "PublicGenerator")
	testResults["PublicKey"] = false
	for i, d := range testKeyData {
		//t.Log(i, d.secret, d.publicKey0)

		a, err := NewAcctPublicKey(d.publicKey0)
		if err != nil {
			t.Errorf("%2d: NewAcctPublicKey() failed: %v", i, err)
			continue
		}
		if g, err := a.MarshalText(); err != nil {
			t.Errorf("%2d: PublicKey.MarshalText() failed: %v", i, err)
		} else if string(g) != d.publicKey0 {
			t.Errorf("%2d: PublicKey.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.publicKey0)
		}

		s, err := NewFamilySeed(d.secret)
		if err != nil {
			t.Skip("Skipping due to NewFamilySeed failure:", err)
		}
		a2 := s.PrivateGenerator.PublicGenerator.Generate(0)
		if a2.X.Cmp(a.X) != 0 {
			t.Errorf("%2d: PublicKey\n\t  result X=%#x\n\texpected X=%#x",
				i, a2.X, a.X)
		}
		if a2.Y.Cmp(a.Y) != 0 {
			t.Errorf("%2d: PublicKey\n\t  result Y=%#x\n\texpected Y=%#x",
				i, a2.Y, a.Y)
		}

		// XXX
		if g, err := a2.MarshalText(); err != nil {
			t.Errorf("%2d: PublicKey.MarshalText() failed: %v", i, err)
		} else if string(g) != d.publicKey0 {
			t.Errorf("%2d: PublicKey.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.publicKey0)
		}

		if g := a.Address(); g != d.address {
			t.Errorf("%2d: PublicKey.Address()\n\treturned %q\n\texpected %q",
				i, g, d.address)
		}
	}
	testResults["PublicKey"] = !t.Failed()
}
