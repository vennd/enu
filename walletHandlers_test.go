package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vennd/enu/counterpartycrypto"
	//	"github.com/vennd/enu/log"
)

func TestCounterpartyWalletCreate(t *testing.T) {
	// Make URL from base URL
	var url = baseURL + "/wallet"
	var wallet counterpartycrypto.CounterpartyWallet

	var send = map[string]interface{}{
		"nonce": time.Now().Unix(),
	}

	assetJsonBytes, err := json.Marshal(send)
	if err != nil {
		t.Errorf("TestWalletCreate(): Unable to create payload")
	}

	responseData, statusCode, err := DoEnuAPITesting("POST", url, assetJsonBytes)

	// deserialise the response if the status is 0
	if err != nil && statusCode != 0 {
		t.Errorf("Error in API call. Error: %s, statusCode: %d\n", err, statusCode)
	}

	if err := json.Unmarshal(responseData, &wallet); err != nil {
		t.Errorf("Error in API call. Unable to unmarshal responseData. Error: %s", err)
	}

	// 20 Addresses should be returned when numberOfAddresses is not specified
	if len(wallet.Addresses) != 20 {
		t.Errorf("Expected 20 addresses to be generated. Got: %d", len(wallet.Addresses))
	}

	// Try a custom number of addresses
	send = map[string]interface{}{
		"nonce":             time.Now().Unix(),
		"numberOfAddresses": 1,
	}
	assetJsonBytes, err = json.Marshal(send)
	if err != nil {
		t.Errorf("TestWalletCreate(): Unable to create payload")
	}
	responseData, statusCode, err = DoEnuAPITesting("POST", url, assetJsonBytes)
	// deserialise the response if the status is 0
	if err != nil && statusCode != 0 {
		t.Errorf("Error in API call. Error: %s, statusCode: %d\n", err, statusCode)
	}
	if err := json.Unmarshal(responseData, &wallet); err != nil {
		t.Errorf("Error in API call. Unable to unmarshal responseData. Error: %s", err)
	}
	// 20 Addresses should be returned when numberOfAddresses is not specified
	if len(wallet.Addresses) != 1 {
		t.Errorf("Expected 1 addresses to be generated. Got: %d", len(wallet.Addresses))
	}

	// Try a an invalid value for numberOfAddresses
	send = map[string]interface{}{
		"nonce":             time.Now().Unix(),
		"numberOfAddresses": "a",
	}
	assetJsonBytes, err = json.Marshal(send)
	if err != nil {
		t.Errorf("TestWalletCreate(): Unable to create payload")
	}
	responseData, statusCode, err = DoEnuAPITesting("POST", url, assetJsonBytes)
	// deserialise the response if the status is 0
	//	t.Errorf("statusCode: %d\n", statusCode)
	if statusCode != 422 {
		t.Errorf("Was expecting for the request to be rejected but it wasn't")
	}
}

func TestWalletBalance(t *testing.T) {
	// Make URL from base URL
	var url = baseURL + "/wallet/balances/1GaZfh9VhxL4J8tBt2jrDvictZEKc8kcHx" // Balance from test address which is used for composing unit testing transactions
	//	var url = baseURL + "/wallet/balances/19kXH7PdizT1mWdQAzY9H4Yyc4iTLTVT5A" // Zero wallet

	var wallet counterpartycrypto.CounterpartyWallet

	var send = map[string]interface{}{
		"nonce": time.Now().Unix(),
	}

	assetJsonBytes, err := json.Marshal(send)
	if err != nil {
		t.Errorf("TestWalletBalance(): Unable to create payload")
	}

	responseData, statusCode, err := DoEnuAPITesting("GET", url, assetJsonBytes)

	// deserialise the response if the status is 0
	if err != nil && statusCode != 0 {
		t.Errorf("Error in API call. Error: %s, statusCode: %d\n", err, statusCode)
	}

	if err := json.Unmarshal(responseData, &wallet); err != nil {
		t.Errorf("Error in API call. Unable to unmarshal responseData. Error: %s", err)
	}
}

func TestRippleWalletCreate(t *testing.T) {
	// Make URL from base URL
	var url = baseURL + "/wallet"
	var wallet map[string]interface{}

	var send = map[string]interface{}{
		"nonce":        time.Now().Unix(),
		"blockchainId": "ripple",
	}

	assetJsonBytes, err := json.Marshal(send)
	if err != nil {
		t.Errorf("TestWalletCreate(): Unable to create payload")
	}

	responseData, statusCode, err := DoEnuAPITesting("POST", url, assetJsonBytes)

	// deserialise the response if the status is 0
	if err != nil && statusCode != 0 {
		t.Errorf("Error in API call. Error: %s, statusCode: %d\n", err, statusCode)
	}

	if err := json.Unmarshal(responseData, &wallet); err != nil {
		t.Errorf("Error in API call. Unable to unmarshal responseData. Error: %s", err)
	}
}
