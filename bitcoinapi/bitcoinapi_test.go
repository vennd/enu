package bitcoinapi

import (
	"testing"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var destinationAddress string = "1Bd5wrFxHYRkk4UCFttcPNMYzqJnQKfXUE"

func TestGetBalance(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	result, err := GetBalance(c, destinationAddress)

	if err != nil {
		t.Errorf(err.Error())
	}

	println(result)

	if result < 0 {
		t.Errorf("Balance is too small\n", result)

	}

}

func TestGetGetBlockCount(t *testing.T) {
	result, err := GetBlockCount()

	if err != nil {
		t.Errorf(err.Error())
	}

	//

	if result < 367576 {
		t.Errorf("Expected: block height > 367576, received: %d\n", result)

	}
}

func TestHttpGet(t *testing.T) {
	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "testing_"+enulib.GenerateRequestId())

	_, _, err := httpGet(c, "http://btc.blockr.io/api/v1/address/balance/198aMn6ZYAczwrE5NvNTUMyJ5qkfy4g3Hi")

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetRawTransaction(t *testing.T) {
	result, err := GetRawTransaction("32e81511a39788cf1c47e6749842e63261ec405614478dbe30dfaac61fee0a93")

	if err != nil {
		t.Errorf(err.Error())
	}

	if result.Hex != "0100000001778260dc82b5a81fc468a385e41c570aa2e2a0ab392ee6ad8d52644eabbd8ea1000000008a473044022010f902abf4d874a71f3d4ed09a8e3bbf7dae0b0d5fc634460af317dd9a8cb3ae0220379d50e4ffdbee1a72b99fb8989736fad1a9fd3f385a6679fbb1ad787e9a569a0141044092b63c102c6dbc07724df065fb2d04e4ccb3d4184a8b8982c1ed5be50b29ce1bdfe0ca3de83c0349d70d8b56adc93e500e108004ac3a7406800321f390d0aeffffffff0280a60b1e000000001976a91485a6c91126e9f70d989ad83abf48d6395b4dff5688ac00e1f505000000001976a914307393dde5fe43d9bc4b39ff668266fb9108ac4f88ac00000000" {
		t.Errorf("Expected: 0100000001778260dc82b5a81fc468a385e41c570aa2e2a0ab392ee6ad8d52644eabbd8ea1000000008a473044022010f902abf4d874a71f3d4ed09a8e3bbf7dae0b0d5fc634460af317dd9a8cb3ae0220379d50e4ffdbee1a72b99fb8989736fad1a9fd3f385a6679fbb1ad787e9a569a0141044092b63c102c6dbc07724df065fb2d04e4ccb3d4184a8b8982c1ed5be50b29ce1bdfe0ca3de83c0349d70d8b56adc93e500e108004ac3a7406800321f390d0aeffffffff0280a60b1e000000001976a91485a6c91126e9f70d989ad83abf48d6395b4dff5688ac00e1f505000000001976a914307393dde5fe43d9bc4b39ff668266fb9108ac4f88ac00000000, got: %s\n", result.Hex)
	}
}

func TestGetConfirmations(t *testing.T) {
	result, err := GetConfirmations("32e81511a39788cf1c47e6749842e63261ec405614478dbe30dfaac61fee0a93")
	if err != nil {
		t.Errorf(err.Error())
	}

	if result <= 0 {
		t.Errorf("Expected > 0, got: %d\n", result)
	}

	result2, err2 := GetConfirmations("invalid")

	if result2 != 0 {
		t.Errorf("Expected != 0, got: %d\n", result2)
	}

	if err2 == nil {
		t.Errorf("Expected err2 != nil, got: %s\n", err2.Error())
	}
}
