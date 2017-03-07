// Developed to be stand alone to enable creation of Counterwallet HD wallets, retrieval of the public, private keys and the addresses.
package counterpartycrypto

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/chaincfg"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcutil/hdkeychain"
	"github.com/vennd/enu/internal/github.com/vennd/mneumonic"
)

type CounterpartyWallet struct {
	Passphrase string   `json:"passphrase"`
	HexSeed    string   `json:"hexSeed"`
	Addresses  []string `json:"addresses"`
	RequestId  string   `json:"requestId"`
}

type CounterpartyAddress struct {
	Value      string `json:"value"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

func getAddressFromPassphrase(passphrase string, position uint32) (CounterpartyAddress, error) {
	var returnValue CounterpartyAddress

	m := mneumonic.FromWords(strings.Split(passphrase, " "))
	hexSeed := m.ToHex()

	hexValue, err := hex.DecodeString(hexSeed)

	if err != nil {
		return returnValue, err
	}

	masterKey, err := hdkeychain.NewMaster(hexValue)
	if err != nil {
		return returnValue, err
	}

	// get m/0'/0/0
	// Hardened key for account 0. ie 0'
	acct0, err := masterKey.Child(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return returnValue, err
	}

	// External account for 0'
	extAcct0, err := acct0.Child(0)
	if err != nil {
		return returnValue, err
	}

	key, err := extAcct0.Child(uint32(position))
	if err != nil {
		return returnValue, err
	}

	// Get the address
	address, err := key.Address(&chaincfg.MainNetParams)
	if err != nil {
		return returnValue, err
	}

	//		log.Printf("Address: %s", address)

	// Get the pubkey and serialise the compressed public key
	privKey, err := key.ECPrivKey()
	if err != nil {
		return returnValue, err
	}

	returnValue.Value = fmt.Sprintf("%s", address)
	returnValue.PrivateKey = hex.EncodeToString(privKey.Serialize())
	returnValue.PublicKey = hex.EncodeToString(privKey.PubKey().SerializeCompressed())

	return returnValue, nil
}

func CreateWallet(numberOfAddressesToGenerate int) (CounterpartyWallet, error) {
	var wallet CounterpartyWallet
	var numAddresses int

	if numberOfAddressesToGenerate <= 0 {
		numAddresses = 20
	} else if numberOfAddressesToGenerate > 100 {
		numAddresses = 100
	} else {
		numAddresses = numberOfAddressesToGenerate
	}

	m := mneumonic.GenerateRandom(128)
	wallet.Passphrase = strings.Join(m.ToWords(), " ")
	wallet.HexSeed = m.ToHex()

	hexValue, err := hex.DecodeString(wallet.HexSeed)

	if err != nil {
		return wallet, err
	}

	masterKey, err := hdkeychain.NewMaster(hexValue)
	if err != nil {
		return wallet, err
	}

	// get m/0'/0/0
	// Hardened key for account 0. ie 0'
	acct0, err := masterKey.Child(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return wallet, err
	}

	// External account for 0'
	extAcct0, err := acct0.Child(0)
	if err != nil {
		return wallet, err
	}

	// Derive extended key (repeat this from 0 to number of addresses-1)
	for i := 0; i <= numAddresses-1; i++ {
		var counterpartyAddress CounterpartyAddress

		key, err := extAcct0.Child(uint32(i))
		if err != nil {
			return wallet, err
		}

		// Get the address
		address, err := key.Address(&chaincfg.MainNetParams)
		if err != nil {
			return wallet, err
		}

		//		log.Printf("Address: %s", address)

		// Get the pubkey and serialise the compressed public key
		privKey, err := key.ECPrivKey()
		if err != nil {
			return wallet, err
		}

		counterpartyAddress.Value = fmt.Sprintf("%s", address.String())
		counterpartyAddress.PrivateKey = hex.EncodeToString(privKey.Serialize())
		counterpartyAddress.PublicKey = hex.EncodeToString(privKey.PubKey().SerializeCompressed())

		wallet.Addresses = append(wallet.Addresses, counterpartyAddress.Value)
	}

	return wallet, nil
}

// GetPrivateKey_Counterparty will retrieve the private key that corresponds to the address given.
// The hierarchical master key is derived from the passphrase and then searches up to the first
// 20 addresses for a match
func GetPrivateKey(passphrase string, address string) (string, error) {
	keys, err := GetPublicPrivateKey(passphrase, address)

	return keys.PrivateKey, err
}

// GetPublicKey_Counterparty will retrieve the public key that corresponds to the address given.
// The hierarchical master key is derived from the passphrase and then searches up to the first
// 20 addresses for a match
func GetPublicKey(passphrase string, address string) (string, error) {
	keys, err := GetPublicPrivateKey(passphrase, address)

	return keys.PublicKey, err
}

func GetPublicPrivateKey(passphrase string, address string) (CounterpartyAddress, error) {
	var result CounterpartyAddress

	for i := 0; i <= 19; i++ {
		generatedAddress, err := getAddressFromPassphrase(passphrase, uint32(i))

		if err != nil {
			errorMessage := fmt.Sprintf("Error with passphrase value: %s\n", err)
			return result, errors.New(errorMessage)
		}

		if generatedAddress.Value == address {
			result.Value = address
			result.PrivateKey = generatedAddress.PrivateKey
			result.PublicKey = generatedAddress.PublicKey

			return result, nil
		}
	}

	errorMessage := fmt.Sprintf("Private and public keys not found for address: %s", address)

	return result, errors.New(errorMessage)
}
