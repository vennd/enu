package counterpartyhandlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartyapi"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
)

func WalletCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var wallet counterpartycrypto.CounterpartyWallet
	var err error

	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	wallet.RequestId = requestId

	var number int
	if m["numberOfAddresses"] != nil {
		number = int(m["numberOfAddresses"].(float64))
	}

	// Create the wallet
	wallet, err = counterpartycrypto.CreateWallet(number)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in CreateWallet(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}
	log.FluentfContext(consts.LOGINFO, c, "Created a new wallet with first address: %s for access key: %s\n (requestID: %s)", wallet.Addresses[0], c.Value(consts.AccessKeyKey).(string), requestId)

	// Return the wallet
	wallet.RequestId = requestId
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(wallet); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	return nil
}

func WalletSend(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var walletPayment enulib.WalletPayment
	var paymentTag string

	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	walletPayment.RequestId = requestId

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletPayment")

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	destinationAddress := m["destinationAddress"].(string)
	asset := m["asset"].(string)
	quantity := uint64(m["quantity"].(float64))

	if m["paymentTag"] != nil {
		paymentTag = m["paymentTag"].(string)
	}

	log.FluentfContext(consts.LOGINFO, c, "WalletSend: received request sourceAddress: %s, destinationAddress: %s, asset: %s, quantity: %d, paymentTag: %s from accessKey: %s\n", sourceAddress, destinationAddress, asset, quantity, c.Value(consts.AccessKeyKey).(string), paymentTag)
	// Generate a paymentId
	paymentId := enulib.GeneratePaymentId()

	log.FluentfContext(consts.LOGINFO, c, "Generated paymentId: %s", paymentId)

	// Return to the client the walletPayment containing requestId and paymentId and unblock the client
	walletPayment.PaymentId = paymentId
	walletPayment.Asset = asset
	walletPayment.SourceAddress = sourceAddress
	walletPayment.DestinationAddress = destinationAddress
	walletPayment.Quantity = quantity
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(walletPayment); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	go delegatedSend(c, c.Value(consts.AccessKeyKey).(string), passphrase, sourceAddress, destinationAddress, asset, quantity, paymentId, paymentTag)

	return nil
}

// Concurrency safe to create and send transactions from a single address.
func delegatedSend(c context.Context, accessKey string, passphrase string, sourceAddress string, destinationAddress string, asset string, quantity uint64, paymentId string, paymentTag string) (string, int64, error) {
	// Write the payment with the generated payment id to the database
	go database.InsertPayment(c, accessKey, 0, c.Value(consts.BlockchainIdKey).(string), paymentId, sourceAddress, destinationAddress, asset, "", quantity, "valid", 0, 1500, paymentTag)

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in GetPublicKey(): %s\n", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.CounterpartyErrors.InvalidPassphrase.Code, consts.CounterpartyErrors.InvalidPassphrase.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	// Mutex lock this address
	counterparty_Mutexes.Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if counterparty_Mutexes.m[sourceAddress] == nil {
		log.FluentfContext(consts.LOGINFO, c, "Created new entry in map for %s", sourceAddress)
		counterparty_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	counterparty_Mutexes.m[sourceAddress].Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked: %s\n", sourceAddress)

	defer counterparty_Mutexes.Unlock()
	defer counterparty_Mutexes.m[sourceAddress].Unlock()

	// We must sleep for at least the time it takes for any transactions to propagate through to the counterparty mempool
	log.FluentfContext(consts.LOGINFO, c, "Sleeping %d milliseconds", counterparty_BackEndPollRate+10000)
	time.Sleep(time.Duration(counterparty_BackEndPollRate+10000) * time.Millisecond)

	log.FluentfContext(consts.LOGINFO, c, "Sleep complete")

	// Create the send
	createResult, errorCode, err := counterpartyapi.CreateSend(c, sourceAddress, destinationAddress, asset, quantity, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in CreateSend(): %s", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, errorCode, err.Error())
		return "", errorCode, err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created send of %d %s to %s: %s", quantity, asset, destinationAddress, createResult)

	// Sign the transactions
	signed, err := counterpartyapi.SignRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in SignRawTransaction(): %s\n", err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.CounterpartyErrors.SigningError.Code, consts.CounterpartyErrors.SigningError.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s", signed)

	// Update the DB with the raw signed TX. This will allow re-transmissions if something went wrong with sending on the network
	database.UpdatePaymentSignedRawTxByPaymentId(c, accessKey, paymentId, signed)

	//	 Transmit the transaction
	txIdSignedTx, err := bitcoinapi.SendRawTransaction(c, signed)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		database.UpdatePaymentWithErrorByPaymentId(c, accessKey, paymentId, consts.CounterpartyErrors.BroadcastError.Code, consts.CounterpartyErrors.BroadcastError.Description)
		return "", consts.CounterpartyErrors.BroadcastError.Code, errors.New(consts.CounterpartyErrors.BroadcastError.Description)
	}

	database.UpdatePaymentCompleteByPaymentId(c, accessKey, paymentId, txIdSignedTx)
	log.FluentfContext(consts.LOGINFO, c, "Complete.")

	return txIdSignedTx, 0, nil
}

func WalletBalance(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var walletbalance enulib.AddressBalances

	requestId := c.Value(consts.RequestIdKey).(string)
	walletbalance.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" || len(address) != 34 {
		log.FluentfContext(consts.LOGERROR, c, "Invalid address")
		handlers.ReturnBadRequest(c, w, consts.GenericErrors.InvalidAddress.Code, consts.GenericErrors.InvalidAddress.Description)

		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "WalletBalance: received request address: %s from accessKey: %s\n", address, c.Value(consts.AccessKeyKey).(string))

	// Get counterparty balances
	result, errorCode, err := counterpartyapi.GetBalancesByAddress(c, address)
	if err != nil {
		handlers.ReturnServerErrorWithCustomError(c, w, errorCode, err.Error())
		return nil
	}

	// Iterate and gather the balances to return
	walletbalance.Address = address
	walletbalance.BlockchainId = consts.CounterpartyBlockchainId
	for _, item := range result {
		var balance enulib.Amount

		balance.Asset = item.Asset
		balance.Quantity = item.Quantity

		walletbalance.Balances = append(walletbalance.Balances, balance)
	}

	// Add BTC balances
	btcbalance, err := bitcoinapi.GetBalance(c, address)
	walletbalance.Balances = append(walletbalance.Balances, enulib.Amount{Asset: "BTC", Quantity: btcbalance})

	// Calculate number of transactions possible
	numberOfTransactions, err := counterpartyapi.CalculateNumberOfTransactions(c, btcbalance)
	if err != nil {
		numberOfTransactions = 0
		log.FluentfContext(consts.LOGERROR, c, "Unable to calculate number of transactions: %s", err.Error())
	}
	walletbalance.NumberOfTransactions = numberOfTransactions

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(walletbalance); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	return nil
}

func ActivateAddress(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	requestId := c.Value(consts.RequestIdKey).(string)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "activateaddress")

	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
		returnCode := enulib.ReturnCode{RequestId: c.Value(consts.RequestIdKey).(string), Code: consts.GenericErrors.InvalidAddress.Code, Description: consts.GenericErrors.InvalidAddress.Description}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			handlers.ReturnServerError(c, w)

			return nil
		}
		return nil

	}

	// Get the amount from the URL
	var amount uint64
	if m["amount"] == nil {
		amount = consts.CounterpartyAddressActivationAmount
	} else {
		amount = uint64(m["amount"].(float64))
	}

	log.FluentfContext(consts.LOGINFO, c, "ActivateAddress: received request address to activate: %s, number of transactions to activate: %d", address, amount)
	// Generate an activationId
	activationId := enulib.GenerateActivationId()

	log.FluentfContext(consts.LOGINFO, c, "Generated activationId: %s", activationId)

	// Return to the client the activationId and requestId and unblock the client
	var result = map[string]interface{}{
		"address":       address,
		"amount":        amount,
		"activationId":  activationId,
		"broadcastTxId": "",
		"status":        "valid",
		"errorMessage":  "",
		"requestId":     requestId,
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	go delegatedActivateAddress(c, address, amount, activationId)

	return nil
}

// Concurrency safe to create and send transactions from a single address.
func delegatedActivateAddress(c context.Context, addressToActivate string, amount uint64, activationId string) (string, int64, error) {
	var complete bool = false
	var txId string
	//	var retries int = 0

	// Copy same context values to local variables which are often accessed
	accessKey := c.Value(consts.AccessKeyKey).(string)
	blockchainId := c.Value(consts.BlockchainIdKey).(string)

	// Need a better way to secure internal wallets
	// Array of internal wallets that can be round robined to activate addresses
	var wallets = []struct {
		Address      string
		Passphrase   string
		BlockchainId string
	}{
		{"1E5YgFkC4HNHwWTF5iUdDbKpzry1SRLv8e", "one two three four five six seven eight nine ten eleven twelve", "counterparty"},
	}

	for complete == false {
		// Pick an internal address to send from
		var randomNumber int = 0
		var sourceAddress = wallets[randomNumber].Address

		// Write the dividend with the generated dividend id to the database
		database.InsertActivation(c, accessKey, activationId, blockchainId, sourceAddress, amount)

		// Calculate the quantity of BTC to send by the amount specified
		// For Counterparty: each transaction = dust_size + miners_fee
		quantity, asset, err := counterpartyapi.CalculateFeeAmount(c, amount)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Could not calculate fee: %s", err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.CounterpartyErrors.MiscError.Code, consts.CounterpartyErrors.MiscError.Description)
			return "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
		}

		txId, _, err = delegatedSend(c, accessKey, wallets[randomNumber].Passphrase, wallets[randomNumber].Address, addressToActivate, asset, quantity, activationId, "")
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in DelegatedSend: %s", err.Error())
			database.UpdatePaymentWithErrorByPaymentId(c, accessKey, activationId, consts.CounterpartyErrors.MiscError.Code, consts.CounterpartyErrors.MiscError.Description)

			complete = false
		} else {
			complete = true
		}
	}

	database.UpdatePaymentCompleteByPaymentId(c, accessKey, activationId, txId)

	return txId, 0, nil
}
