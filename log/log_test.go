package log

import (
	"testing"
)

func TestPrintf(t *testing.T) {
	Printf("TestPrintf. %s x %d = %f", "two", 2, 4.0)
}

func TestObject(t *testing.T) {
	type amount struct {
		Address  string `json:"address"`
		Quantity uint64 `json:"quantity"`
	}
	var assetBalances struct {
		Asset     string   `json:"asset"`
		Divisible bool     `json:"divisible"`
		Balances  []amount `json:"balances"`
	}

	assetBalances.Asset = "testAsset"
	assetBalances.Divisible = false

	var balance1 amount
	balance1.Address = "address1"
	balance1.Quantity = 1000

	var balance2 amount
	balance2.Address = "address2"
	balance2.Quantity = 2000

	assetBalances.Balances = append(assetBalances.Balances, balance1)
	assetBalances.Balances = append(assetBalances.Balances, balance2)

	object("enu", assetBalances, "Testing fluentd logger", false)
}
