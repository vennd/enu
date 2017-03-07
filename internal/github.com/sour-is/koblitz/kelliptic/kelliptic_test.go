// Copyright 2010 The Go Authors. All rights reserved.
// Copyright 2011 ThePiachu. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kelliptic

import (
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"math/big"
	"testing"
)

var _ elliptic.Curve = S256() // verify we satisfy the elliptic.Curve interface

func TestOnCurve(t *testing.T) {
	s160 := S160()
	if !s160.IsOnCurve(s160.Gx, s160.Gy) {
		t.Errorf("FAIL S160")
	}
	s192 := S192()
	if !s192.IsOnCurve(s192.Gx, s192.Gy) {
		t.Errorf("FAIL S192")
	}
	s224 := S224()
	if !s224.IsOnCurve(s224.Gx, s224.Gy) {
		t.Errorf("FAIL S224")
	}
	s256 := S256()
	if !s256.IsOnCurve(s256.Gx, s256.Gy) {
		t.Errorf("FAIL S256")
	}
}

type baseMultTest struct {
	k    string
	x, y string
}

//TODO: add more test vectors
var s256BaseMultTests = []baseMultTest{
	{
		"AA5E28D6A97A2479A65527F7290311A3624D4CC0FA1578598EE3C2613BF99522",
		"34F9460F0E4F08393D192B3C5133A6BA099AA0AD9FD54EBCCFACDFA239FF49C6",
		"B71EA9BD730FD8923F6D25A7A91E7DD7728A960686CB5A901BB419E0F2CA232",
	},
	{
		"7E2B897B8CEBC6361663AD410835639826D590F393D90A9538881735256DFAE3",
		"D74BF844B0862475103D96A611CF2D898447E288D34B360BC885CB8CE7C00575",
		"131C670D414C4546B88AC3FF664611B1C38CEB1C21D76369D7A7A0969D61D97D",
	},
	{
		"6461E6DF0FE7DFD05329F41BF771B86578143D4DD1F7866FB4CA7E97C5FA945D",
		"E8AECC370AEDD953483719A116711963CE201AC3EB21D3F3257BB48668C6A72F",
		"C25CAF2F0EBA1DDB2F0F3F47866299EF907867B7D27E95B3873BF98397B24EE1",
	},
	{
		"376A3A2CDCD12581EFFF13EE4AD44C4044B8A0524C42422A7E1E181E4DEECCEC",
		"14890E61FCD4B0BD92E5B36C81372CA6FED471EF3AA60A3E415EE4FE987DABA1",
		"297B858D9F752AB42D3BCA67EE0EB6DCD1C2B7B0DBE23397E66ADC272263F982",
	},
	{
		"1B22644A7BE026548810C378D0B2994EEFA6D2B9881803CB02CEFF865287D1B9",
		"F73C65EAD01C5126F28F442D087689BFA08E12763E0CEC1D35B01751FD735ED3",
		"F449A8376906482A84ED01479BD18882B919C140D638307F0C0934BA12590BDE",
	},
}

//TODO: test different curves as well?
func TestBaseMult(t *testing.T) {
	s256 := S256()
	for i, e := range s256BaseMultTests {
		k, ok := new(big.Int).SetString(e.k, 16)
		if !ok {
			t.Errorf("%d: bad value for k: %s", i, e.k)
		}
		x, y := s256.ScalarBaseMult(k.Bytes())
		if fmt.Sprintf("%X", x) != e.x || fmt.Sprintf("%X", y) != e.y {
			t.Errorf("%d: bad output for k=%s: got (%X, %X), want (%s, %s)", i, e.k, x, y, e.x, e.y)
		}
		if testing.Short() && i > 5 {
			break
		}
	}
}

//TODO: test more curves?
func BenchmarkBaseMult(b *testing.B) {
	s256 := S224()
	e := s256BaseMultTests[0] //TODO: check, used to be 25 instead of 0, but it's probably ok
	k, _ := new(big.Int).SetString(e.k, 16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s256.ScalarBaseMult(k.Bytes())
	}
}

//TODO: test more curves?
func TestMarshal(t *testing.T) {
	s256 := S256()

	// 04e1aa5f8cdeb412bcd043c4df0c9d0d4e5132b728f521ffe8590a25775cb5cd956ca2518d87b502d7971f5c7b68afbb8d7406ba580eb683fd57466a6c5a7982bb

	_, x, y, err := elliptic.GenerateKey(s256, rand.Reader)
	if err != nil {
		t.Error(err)
		return
	}
	serialised := elliptic.Marshal(s256, x, y)
	xx, yy := elliptic.Unmarshal(s256, serialised)
	if xx == nil {
		t.Error("failed to unmarshal")
		return
	}
	if xx.Cmp(x) != 0 || yy.Cmp(y) != 0 {
		t.Error("unmarshal returned different values")
		return
	}
}

func TestCompression(t *testing.T) {
	s256 := S256()

	Convey(`Decompressed Points should equal original point.`, t, func() {
		points := []string{"02a8b7effaef9a36d0fe3b3c218e0c3b621c957166954d8d00c485ce3818196745",
			"037e8accb04b42d262de134589e9e18f5f9f03c136ff6efdb3f57601365f743cc6"}

		for _, px := range points {
			p, _ := hex.DecodeString(px)

			x, y, err := s256.DecompressPoint(p)

			cp := CompressPoint(s256, x, y)

			So(cp, ShouldResemble, p)

			xx, yy, err := s256.DecompressPoint(cp)
			if err != nil {
				t.Error(err)
				return
			}

			So(x, ShouldResemble, xx)
			So(y, ShouldResemble, yy)
		}
	})

	Convey(`Unserialize point that is not compressed.`, t, func() {
		p, _ := hex.DecodeString("04a8b7effaef9a36d0fe3b3c218e0c3b621c957166954d8d00c485ce38181967451040be9bd94c9de7b8296d9293073e4b7e91c7ebf80a3137e355fc314bcc06de")

		x, y, err := s256.DecompressPoint(p)

		sp := elliptic.Marshal(s256, x, y)
		if err != nil {
			t.Error(err)
			return
		}

		So(sp, ShouldResemble, p)
	})

	Convey(`Test Falure Cases`, t, func() {
		badheader, _ := hex.DecodeString("077e8accb04b42d262de134589e9e18f5f9f03c136ff6efdb3f57601365f743cc3")
		badlength, _ := hex.DecodeString("037accb04b42d262de134589e9e18f5f9f03c136ff6efdb3f57601365f743cc3")
		notcurve, _ := hex.DecodeString("030000000000000000000000000000000000000000000000000000000000000000")

		Convey(`Bad Header should fail`, func() {
			_, _, err := s256.DecompressPoint(badheader)
			So(err, ShouldNotBeNil)
		})

		Convey(`Bad Length should fail`, func() {
			_, _, err := s256.DecompressPoint(badlength)
			So(err, ShouldNotBeNil)
		})

		Convey(`Not on Curve`, func() {
			_, _, err := s256.DecompressPoint(notcurve)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestLegendreSymbol(t *testing.T) {

	zero := make([]int64, 0, 14)
	residue := make([]int64, 0, 14)
	nonresidue := make([]int64, 0, 14)

	p := big.NewInt(17)

	for i := int64(0); i < 17; i++ {
		a := big.NewInt(i)
		l := LegendreSymbol(a, p)

		if l == 0 {
			zero = append(zero, i)
		} else if l > 0 {
			residue = append(residue, i)
		} else {
			nonresidue = append(nonresidue, i)
		}
	}

	Convey(`Find the residues for numbers mod 17`, t, func() {
		So([]int64{0}, ShouldResemble, zero)
		So([]int64{1, 2, 4, 8, 9, 13, 15, 16}, ShouldResemble, residue)
		So([]int64{3, 5, 6, 7, 10, 11, 12, 14}, ShouldResemble, nonresidue)
	})

}

func TestModuloSqrt(t *testing.T) {

	ZERO := big.NewInt(0)
	ONE := big.NewInt(1)

	c := new(Curve)

	Convey(`Check sqrt mod 2 cases`, t, func() {
		c.P = big.NewInt(2)

		So(c.Sqrt(ZERO).Int64(), ShouldEqual, 0)
		So(c.Sqrt(ONE).Int64(), ShouldEqual, 1)
	})

	Convey(`Verify that sqrt mod 17 results are correct.`, t, func() {
		c.P = big.NewInt(17)
		for i := int64(0); i < 17; i++ {

			rt := c.Sqrt(big.NewInt(i))
			if rt.Cmp(ZERO) == 0 {
				continue
			}

			rt.Mul(rt, rt)
			rt.Mod(rt, c.P)

			So(i, ShouldEqual, rt.Int64())
		}
	})

	// Mod 2 and 17 cover all code for Sqrt.
	// These tests further test the dragon code.

	SkipConvey(`Verify that sqrt mod 73 results are correct.`, t, func() {
		c.P = big.NewInt(73)
		for i := int64(0); i < 73; i++ {

			rt := c.Sqrt(big.NewInt(i))
			if rt.Cmp(ZERO) == 0 {
				continue
			}

			rt.Mul(rt, rt)
			rt.Mod(rt, c.P)

			So(i, ShouldEqual, rt.Int64())
		}
	})

}
