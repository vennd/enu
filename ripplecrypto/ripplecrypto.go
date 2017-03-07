package ripplecrypto

import (
	"fmt"

	"math/big"
	"strings"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/internal/github.com/vennd/mneumonic"
	"github.com/vennd/enu/log"

	"github.com/vennd/enu/internal/bitbucket.org/dchapes/ripple/crypto/rkey"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

type RippleWallet struct {
	Passphrase string   `json:"passphrase"`
	HexSeed    string   `json:"hexSeed"`
	Addresses  []string `json:"addresses"`
	RequestId  string   `json:"requestId"`
}

type RippleAddress struct {
	Value      string `json:"value"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

// Generate hex seed from ripple secret
func ToHexSeedString(secret string) (string, error) {
	s, err := rkey.NewFamilySeed(secret)
	seedHex := fmt.Sprintf("%x", s.Seed)

	return seedHex, err
	//	log.Printf("From secret: %s, hex seed: %x", secret, s.Seed)
}

// Generate ripple encoded secret from hex string of the seed
func ToSecret(hexString string) (string, error) {
	ints2 := new(big.Int)
	ints2.SetString(hexString, 16)
	s, err := rkey.NewSeed(ints2)
	if err != nil {
		return "", err
	}

	address, err := s.MarshalText()
	if err != nil {
		return "", err
	}
	//	log.Printf("From seedHex: %s, address: %s", seedHex, string(address))
	return string(address), nil
}

// Convert passphrase to Ripple secret
func PassphraseToSecret(c context.Context, passphrase string) string {
	var result string

	seed := mneumonic.FromWords(strings.Split(passphrase, " "))
	secret, err := ToSecret(seed.ToHex())
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in ripplecrypto.ToSecret(): %s", err.Error())

		return result
	}

	return secret
}

// Generates a ripple wallet offline
func CreateWallet(numberOfAddressesToGenerate int) (RippleWallet, error) {
	var wallet RippleWallet
	var numAddresses int

	if numberOfAddressesToGenerate <= 0 {
		numAddresses = 1
	} else if numberOfAddressesToGenerate > 100 {
		numAddresses = 100
	} else {
		numAddresses = numberOfAddressesToGenerate
	}

	m := mneumonic.GenerateRandom(128)
	wallet.Passphrase = strings.Join(m.ToWords(), " ")
	wallet.HexSeed = m.ToHex()

	// Generate ripple master secret
	masterSecret, err := ToSecret(wallet.HexSeed)
	if err != nil {
		return wallet, err
	}

	s, err := rkey.NewFamilySeed(masterSecret)
	if err != nil {
		return wallet, err
	}

	// Derive keys
	for i := 0; i <= numAddresses-1; i++ {
		var rippleAddress RippleAddress
		var pos = uint32(i)

		privateKey, err := s.PrivateGenerator.Generate(pos).MarshalText()
		if err != nil {
			return wallet, err
		}
		rippleAddress.PrivateKey = string(privateKey)

		publicKey, err := s.PrivateGenerator.PublicGenerator.Generate(pos).MarshalText()
		if err != nil {
			return wallet, err
		}
		rippleAddress.PublicKey = string(publicKey)

		rippleAddress.Value = s.PrivateGenerator.PublicGenerator.Generate(pos).Address()

		//		fmt.Println("privateKey:", rippleAddress.PrivateKey, "address:", rippleAddress.Value)

		wallet.Addresses = append(wallet.Addresses, rippleAddress.Value)
	}

	return wallet, nil
}
