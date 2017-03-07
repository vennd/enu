package counterpartyhandlers

import (
	"testing"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var c context.Context

func setContext() {
	c = context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "test"+enulib.GenerateRequestId())
	c = context.WithValue(c, consts.AccessKeyKey, "unittesting")
	c = context.WithValue(c, consts.BlockchainIdKey, "counterparty")
	c = context.WithValue(c, consts.EnvKey, "dev")
}

func TestActivateAddress(t *testing.T) {
	var testData = []struct {
		AddressToActivate string
		Amount            uint64
		ActivationId      string
		ExpectedResult    string
		ExpectedErrorCode int64
		CaseDescription   string
	}{
		{"1KgUFkLpypNbNsJJKsTN5qjwq76gKWsH7d", 10, "TestActivateAddress1", "success", 0, "Successful case"},
		{"1KgUFkLpypNbNsJJKsTN5qjwq76gKWsH7d", 10000000000, "TestActivateAddress2", "success", 0, "Successful. Defaults kick in"},
	}

	setContext()

	for _, s := range testData {
		txId, errorCode, err := delegatedActivateAddress(c, s.AddressToActivate, s.Amount, s.ActivationId)

		if txId != s.ExpectedResult || errorCode != s.ExpectedErrorCode {
			t.Errorf("Expected: %s errorCode: %d, Got: %s errorCode: %d\nCase: %s\n", s.ExpectedResult, s.ExpectedErrorCode, txId, errorCode, s.CaseDescription)

			// Additionally log the error if we got an error
			if err != nil {
				t.Error(err.Error())
			}
		}
	}
}
