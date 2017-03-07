package database

import (
	"fmt"
	"testing"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

func TestUpdateNonceByAccessKey(t *testing.T) {
	// Code to test update nonce

	nonce := int64(100000)

	accessKey := "71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16"

	err := UpdateNonce(accessKey, nonce)

	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetNonceByAccessKey(t *testing.T) {
	// Code to test nonce check

	// non existing acessKey
	//nonce := GetNonceByAccessKey("73a7b844c80c3c5cf532d1dd843321b1c733c0c67e5b5ab162ca283da4cfc182")
	nonce := GetNonceByAccessKey("71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16")

	if nonce == 0 {
		t.Errorf("Unable to retrieve nonce. Expected != 0, got: %d\n", nonce)
	}
}

func TestUserKeyExists(t *testing.T) {
	exists := UserKeyExists("71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16")
	notExists := UserKeyExists("narebeko")

	if exists == false {
		t.Errorf("User test key doesn't exist. Expected: true, got: %t\n", exists)
	}

	if notExists == true {
		t.Errorf("User test key doesn't exist. Expected: false, got: %t\n", exists)
	}
}

func TestCreateUserKey(t *testing.T) {
	// Create a user key
	key, _, err := CreateUserKey(777, "", consts.CounterpartyBlockchainId, "", "")
	if err != nil {
		t.Errorf("Unable to create user: %s\n", err.Error())
	}

	// Get the blockchainId and check it was set correctly
	blockchainId := GetBlockchainIdByUserKey(key)
	if blockchainId != consts.CounterpartyBlockchainId {
		t.Errorf("userKey blockchainId not set correctly. Expected: %s, got: %s\n", consts.CounterpartyBlockchainId, blockchainId)
	}

	// Update user key with all possible statuses
	for _, value := range consts.AccessKeyStatuses {
		err2 := UpdateUserKeyStatus(key, value)
		if err2 != nil {
			t.Errorf("Unable to update userKey status: %s\n", err2.Error())
		}

		status := GetStatusByUserKey(key)
		if status != value {
			t.Errorf("User key status not set correctly. Expected: %s, got: %s\n", value, status)
		}
	}

	// Disable the user key that we created previously
	err3 := UpdateUserKeyStatus(key, consts.AccessKeyInvalidStatus)
	if err3 != nil {
		t.Errorf("Unable to update userKey status: %s\n", err3.Error())
	}

	// Attempt to set status to an invalid value
	err4 := UpdateUserKeyStatus(key, "this_should_not_work")
	if err4 == nil {
		t.Errorf("userKey status could be updated to an invalid value: %s\n", "this_should_not_work")
	}
}

// Also tests insert payment
func TestInsertActivationandInsertPayment(t *testing.T) {
	activationId := "test_" + enulib.GenerateActivationId()
	requestId := "test_" + enulib.GenerateRequestId()

	ctx := context.TODO()
	ctx = context.WithValue(ctx, consts.RequestIdKey, requestId)

	// Okay insertion
	InsertActivation(ctx, "TestAccessKey", activationId, "BlockchainId", "AddressToActive", 100)

	// Insert a corresponding payment
	InsertPayment(ctx, "TestAccessKey", 0, "testblockchain", activationId, "InternalAddress", "AddressToActive", "BTC", "", 1, "testing", 0, 1500, "")

	// need to cover more columns and how to test that payment actually works?

	// Retrieve the payment
	payment := GetPaymentByPaymentId(ctx, "TestAccessKey", activationId)
	if payment.SourceAddress != "InternalAddress" || payment.DestinationAddress != "AddressToActive" || payment.Asset != "BTC" || payment.Amount != 1 || payment.TxFee != 1500 || payment.PaymentTag != "" {
		t.Errorf("Expected: %s, %s, %s, %d, %d, %s, %s. Got: %s, %s, %s, %d, %d, %s, %s", "InternalAddress", "AddressToActive", "BTC", 1, 1500, "", payment.SourceAddress, payment.DestinationAddress, payment.Asset, payment.Amount, payment.TxFee, payment.PaymentTag)
	}

	// Retrieve the activation
	activation := GetActivationByActivationId(ctx, "TestAccessKey", activationId)
	//		var result = map[string]interface{}{
	//		"address":       addressToActivate,
	//		"amount":        amount,
	//		"activationId":  activationId,
	//		"broadcastTxId": broadcastTxId,
	//		"status":        status,
	//		"errorMessage":  errorMessage,
	//		"requestId":     requestId,
	//	}
	if activation["address"].(string) != "AddressToActive" {
		t.Errorf("Expected: %s. Got: %s\n", "AddressToActive", activation["address"])
	}
	if activation["status"] != "testing" {
		t.Errorf("Expected: %s. Got: %s\n", "valid", activation["status"])
	}
}

func TestGetPaymentsByAddress(t *testing.T) {
	var testData = []struct {
		Address         string
		ExpectedResult  []enulib.SimplePayment
		CaseDescription string
	}{
		{"unittesting1", []enulib.SimplePayment{
			{SourceAddress: "unittesting1", DestinationAddress: "unittesting2", Asset: "CHANGE", Amount: 124364, TxFee: 1500, BroadcastTxId: "", Status: "valid", PaymentId: "unittesting1", ErrorMessage: "", PaymentTag: ""},
			{SourceAddress: "unittesting3", DestinationAddress: "unittesting1", Asset: "GOLD", Amount: 7853, TxFee: 2000, BroadcastTxId: "success", Status: "complete", PaymentId: "unittesting12", ErrorMessage: "", PaymentTag: ""},
			{SourceAddress: "unittesting1", DestinationAddress: "unittesting3", Asset: "SILVER", Amount: 23523456, TxFee: 1000, BroadcastTxId: "", Status: "error", PaymentId: "unittesting123", ErrorMessage: "Not enough silver", PaymentTag: "invoice123"},
		}, "successful"},
		{"addressdoesntexist", []enulib.SimplePayment{{SourceAddress: "", DestinationAddress: "", Asset: "", Amount: 0, TxFee: 0, BroadcastTxId: "", Status: "", PaymentId: "", ErrorMessage: "", PaymentTag: ""}}, "Address doesn't exist"},
	}

	Init()

	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "test"+enulib.GenerateRequestId())
	c = context.WithValue(c, consts.AccessKeyKey, "71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16")
	c = context.WithValue(c, consts.BlockchainIdKey, "counterparty")
	c = context.WithValue(c, consts.EnvKey, "dev")

	for _, s := range testData {
		result := GetPaymentsByAddress(c, c.Value(consts.AccessKeyKey).(string), s.Address)

		for i, paymentResult := range result {
			if paymentResult.SourceAddress != s.ExpectedResult[i].SourceAddress ||
				paymentResult.DestinationAddress != s.ExpectedResult[i].DestinationAddress ||
				paymentResult.Asset != s.ExpectedResult[i].Asset ||
				paymentResult.Amount != s.ExpectedResult[i].Amount ||
				paymentResult.TxFee != s.ExpectedResult[i].TxFee ||
				paymentResult.BroadcastTxId != s.ExpectedResult[i].BroadcastTxId ||
				paymentResult.Status != s.ExpectedResult[i].Status ||
				paymentResult.PaymentId != s.ExpectedResult[i].PaymentId ||
				paymentResult.ErrorMessage != s.ExpectedResult[i].ErrorMessage ||
				paymentResult.PaymentTag != s.ExpectedResult[i].PaymentTag {
				paymentResultString := fmt.Sprintf("%#v", paymentResult)
				expectedResultString := fmt.Sprintf("%#v", s.ExpectedResult[i])

				t.Errorf("Expected: %s, Got: %s\nCase: %s\n", paymentResultString, expectedResultString, s.CaseDescription)
			}
		}
	}
}

func TestUpdateDividendWithErrorByDividendId(t *testing.T) {
	Init()

	c := context.TODO()
	c = context.WithValue(c, consts.RequestIdKey, "test"+enulib.GenerateRequestId())
	c = context.WithValue(c, consts.AccessKeyKey, "71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16")
	c = context.WithValue(c, consts.BlockchainIdKey, "counterparty")
	c = context.WithValue(c, consts.EnvKey, "dev")

	// Update existing dividend
	err := UpdateDividendWithErrorByDividendId(c, c.Value(consts.AccessKeyKey).(string), "3bff0a8e661d8000388971564d0f82ff", 999999999999999, "Unit testing error")
	if err != nil {
		t.Errorf(err.Error())
	}

	// Dividend doesn't exist
	err = UpdateDividendWithErrorByDividendId(c, c.Value(consts.AccessKeyKey).(string), "3bff0a8e661d8000388971564d0f82ffxxxxxx", 999999999999999, "Unit testing error")
	if err == nil || err.Error() != consts.GenericErrors.NotFound.Description {
		t.Errorf("Expected: %s, Got: %s", consts.GenericErrors.NotFound.Description, err)
	}
}

// Also tests insert payment
func TestInsertTrustAsset(t *testing.T) {
	activationId := "test_" + enulib.GenerateActivationId()
	requestId := "test_" + enulib.GenerateRequestId()

	ctx := context.TODO()
	ctx = context.WithValue(ctx, consts.RequestIdKey, requestId)

	// Insert trustasset
	InsertTrustAsset(ctx, "TestAccessKey", activationId, "BlockchainId", "coolasset", "niceissuer", 1000000)

	// Retrieve the trustasset
	//	payment := GetPaymentByPaymentId(ctx, "TestAccessKey", activationId)
	//	if payment.SourceAddress != "InternalAddress" || payment.DestinationAddress != "AddressToActive" || payment.Asset != "BTC" || payment.Amount != 1 || payment.TxFee != 1500 || payment.PaymentTag != "" {
	//		t.Errorf("Expected: %s, %s, %s, %d, %d, %s, %s. Got: %s, %s, %s, %d, %d, %s, %s", "InternalAddress", "AddressToActive", "BTC", 1, 1500, "", payment.SourceAddress, payment.DestinationAddress, payment.Asset, payment.Amount, payment.TxFee, payment.PaymentTag)
	//	}
}

//
func TestGetAssetByAssetId(t *testing.T) {
	var assetId string = "1e33a6997f0ab51d30713ede33abd0c5"
	var accessKey string = "71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16"
	requestId := "test_" + enulib.GenerateRequestId()

	ctx := context.TODO()
	ctx = context.WithValue(ctx, consts.RequestIdKey, requestId)

	assetRequest, err := GetAssetByAssetId(ctx, accessKey, assetId)
	if err != nil {
		t.Error(err.Error())
	}

	t.Logf("%+#v", assetRequest)
}
