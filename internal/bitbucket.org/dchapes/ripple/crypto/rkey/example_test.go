package rkey_test

import (
	"github.com/vennd/enu/internal/bitbucket.org/dchapes/ripple/crypto/rkey"
	"fmt"
)

const secret = "sp6JS7f14BuwFY8Mw6bTtLKWauoUs"

func Example_address() {
	s, _ := rkey.NewFamilySeed(secret)
	pubkey := s.PrivateGenerator.PublicGenerator.Generate(0)
	addr := pubkey.Address()
	fmt.Println("secret:", secret, "address:", addr)
	// Output:
	// secret: sp6JS7f14BuwFY8Mw6bTtLKWauoUs address: rJq5ce8cdbWBsysXx32rvLMV6DUxMwruMT
}
