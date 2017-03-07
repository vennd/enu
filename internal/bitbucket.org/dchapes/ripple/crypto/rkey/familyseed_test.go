package rkey

import (
	"math/big"
	"testing"
)

var testKeyData = []struct {
	secret      string
	secretVal   string
	privateGen  string
	publicGen   string
	privateKey0 string
	publicKey0  string
	address     string
}{
	{
		"sp6JS7f14BuwFY8Mw6bTtLKWauoUs",
		"0",
		"0xEC2D57691D9B2D40182AC565032054B7D784BA96B18BCB5BE0BB4E70E3FB041E",
		"fhiz7iEpEpkw1T9toGXWNGUZnSNg5erTmMTET5w3EnCFrtMh74CB",
		"p97YySUSRcdomJxbZKUKKEokJRzDEZMcra9goGqdQ1d4VWYbUYS",
		"aBQ7bHCe5MqMWCR4ihtdKvxgiBS1Mdd3A1tVyR1D1Ersoej421TA",
		"rJq5ce8cdbWBsysXx32rvLMV6DUxMwruMT",
	},
	{
		"sp6JS7f14BuwFY8Mw6bTtLKdMm3s8",
		"1",
		"0x29F61516876C25379A7BF4FAA2B3CA6F6B53EAC90E7DE47671FEC4A818D51441",
		"fht1rtAAn5Q7yhXJbLP1AeYx9kCyhVTK6qXvZwLFgeNHkMvYbh57",
		"pwPUn8JTAScYoye8ZTZ7DpiUAn9PbtUfK1wXFouhYPDe288knKo", // ??
		"aBP46hwGLfq2VVjYhqwR9WQ4F3LEsSbzBxxv7jDA7Vv2nXbx3hMB",
		"rMsXgpgs6kuisJyy7nsNWmsJyfw8ydYh3w",
	},
	{
		"shHM53KPZ87Gwdqarm1bAmPeXg8Tn",
		"0x71ED064155FFADFA38782C5E0158CB26",
		"0x7CFBA64F771E93E817E15039215430B53F7401C34931D111EAB3510B22DBB0D8",
		"fht5yrLWh3P8DrJgQuVNDPQVXGTMyPpgRHFKGQzFQ66o3ssesk3o",
		"pwMPbuE25rnajigDPBEh9Pwv8bMV2ebN9gVPTWTh4c3DtB14iGL",
		"aBRoQibi2jpDofohooFuzZi9nEzKw9Zdfc4ExVNmuXHaJpSPh8uJ",
		"rhcfR9Cg98qCxHpCcPBmMonbDBXo84wyTn",
	},
	{
		"snoPBrXtMeMyMHUVTgbuqAfg1SUTb",
		"0xDEDCE9CE67B451D852FD4E846FCDE31C",
		"0x395898665728F57DE5D90F1DE102278A967D6941A45A6C9A98CB123394489E55",
		"fhuJKrhSDzV2SkjLn9qbwm5AaRmrxDPfFsHDCP6yfDZWcxDFz4mt",
		"p9JfM6HHi64m6mvB6v5k7G2b1cXzGmYiCNJf6GHPKvFTWdeRVjh",
		"aBQG8RQAzjs1eTKFEAQXr2gS4utcDiEC9wmi7pfUPTi27VCahwgw",
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
	},
	{
		"saGwBRReqUNKuWNLpUAq8i8NkXEPN",
		"0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
		"0x48977BDA680B92A425F797193CBE32D10554BF340AB02B03C0136D8FF6BAA03D",
		"fh1JNg5JhHcJvEfBusc4Mdi65KBjeaFcWoUycN8oeEVWQxYZKJkB",
		"p9BofrT8Qw8r6j8xKdPqAdUa3PuES8XwXwv2pNtMmViRvktgiPb", // ??
		"aBPh4FgrSkzC3aUuyExJ8Ss9kJxHMXVqLnuSu9Lnjissa7231Q6Y",
		"rGzzsoiXsM3RmHU9Fu9eDQ6hoNoouyQiNT",
	},
}

var testResults = make(map[string]bool)

func checkDepends(t testing.TB, deps ...string) {
	for _, test := range deps {
		if result, ok := testResults[test]; ok && !result {
			t.Skipf("Skipping due to Test%s failure", test)
		} else if !ok {
			//t.Logf("warning: depends on Test%s", test)
		}
	}
}

func TestFamilySeed(t *testing.T) {
	testResults["FamilySeed"] = false
	v := new(big.Int)
	for i, d := range testKeyData {
		//t.Log(i, d.secret, d.secretVal)

		v.SetString(d.secretVal, 0)
		s, err := NewSeed(v)
		if err != nil {
			t.Errorf("%2d: NewSeed() failed: %v", i, err)
			continue
		}
		if g, err := s.MarshalText(); err != nil {
			t.Errorf("%2d: FamilySeed.MarshalText() failed: %v", i, err)
		} else if string(g) != d.secret {
			t.Errorf("%2d: FamilySeed.MarshalText()\n\treturned %q,\n\texpected %q",
				i, string(g), d.secret)
		}

		s2 := new(FamilySeed)
		if err := s2.UnmarshalText([]byte(d.secret)); err != nil {
			t.Errorf("%2d: FamilySeed.UnmarshalText(%q) failed: %v",
				i, d.secret, err)
		} else if s2.Seed.Cmp(v) != 0 {
			t.Errorf("%2d: FamilySeed.UnmarshalText\n\t  result %#x,\n\texpected %#x",
				i, s2.Seed, v)
		}
	}
	testResults["FamilySeed"] = !t.Failed()
}
