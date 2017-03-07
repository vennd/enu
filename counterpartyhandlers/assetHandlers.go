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

func AssetCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	var assetStruct enulib.Asset
	requestId := c.Value(consts.RequestIdKey).(string)
	assetStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	asset := m["asset"].(string)
	//	description := m["description"].(string)
	quantity := uint64(m["quantity"].(float64))
	divisible := m["divisible"].(bool)

	log.FluentfContext(consts.LOGINFO, c, "AssetCreate: received request sourceAddress: %s, asset: %s, quantity: %s, divisible: %b from accessKey: %s\n", sourceAddress, asset, quantity, divisible, c.Value(consts.AccessKeyKey).(string))

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in counterpartycrypto.GetPublicKey(): %s\n", err)

		handlers.ReturnServerError(c, w)
		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "retrieved publickey: %s", sourceAddressPubKey)

	// Generate random asset name
	randomAssetName, errorCode, err := counterpartyapi.GenerateRandomAssetName(c)
	if err != nil {
		handlers.ReturnServerErrorWithCustomError(c, w, errorCode, err.Error())

		return nil
	}

	// Generate an assetId
	assetId := enulib.GenerateAssetId()
	log.Printf("Generated assetId: %s", assetId)
	assetStruct.AssetId = assetId
	assetStruct.Asset = randomAssetName
	assetStruct.Description = asset
	assetStruct.Quantity = quantity
	assetStruct.Divisible = divisible
	assetStruct.SourceAddress = sourceAddress

	// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(assetStruct); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	// Start asset creation in async mode
	go delegatedCreateIssuance(c, c.Value(consts.AccessKeyKey).(string), passphrase, sourceAddress, assetId, randomAssetName, asset, quantity, divisible)

	return nil
}

// Concurrency safe to create and send transactions from a single address.
func delegatedCreateIssuance(c context.Context, accessKey string, passphrase string, sourceAddress string, assetId string, asset string, assetDescription string, quantity uint64, divisible bool) (string, int64, error) {
	// Write the asset with the generated asset id to the database
	go database.InsertAsset(accessKey, c.Value(consts.BlockchainIdKey).(string), assetId, sourceAddress, "", asset, assetDescription, quantity, divisible, "valid")

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error with GetPublicKey(): %s", err)
		return "", consts.CounterpartyErrors.InvalidPassphrase.Code, errors.New(consts.CounterpartyErrors.InvalidPassphrase.Description)
	}

	// Mutex lock this address
	counterparty_Mutexes.Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if counterparty_Mutexes.m[sourceAddress] == nil {
		log.Printf("Created new entry in map for %s\n", sourceAddress)
		counterparty_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	counterparty_Mutexes.m[sourceAddress].Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked: %s\n", sourceAddress)

	// We must sleep for at least the time it takes for any transactions to propagate through to the counterparty mempool
	log.FluentfContext(consts.LOGINFO, c, "Sleeping")
	time.Sleep(time.Duration(counterparty_BackEndPollRate+3000) * time.Millisecond)

	defer counterparty_Mutexes.Unlock()
	defer counterparty_Mutexes.m[sourceAddress].Unlock()

	log.FluentfContext(consts.LOGINFO, c, "Composing the CreateNumericIssuance transaction")
	// Create the issuance
	createResult, errCode, err := counterpartyapi.CreateIssuance(c, sourceAddress, asset, assetDescription, quantity, divisible, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in CreateIssuance(): %s", err.Error())
		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, errCode, err.Error())
		return "", errCode, err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created issuance of %d %s (%s) at %s: %s\n", quantity, asset, assetDescription, sourceAddress, createResult)
	//	database.UpdateAssetNameByAssetId(c, accessKey, assetId, asset)

	// Sign the transactions
	signed, err := counterpartyapi.SignRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in SignRawTransaction(): %s", err.Error())

		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.CounterpartyErrors.SigningError.Code, consts.CounterpartyErrors.SigningError.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s\n", signed)

	//	 Transmit the transaction
	txIdSignedTx, err := bitcoinapi.SendRawTransaction(c, signed)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in SendRawTransaction(): %s", err.Error())
		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.CounterpartyErrors.BroadcastError.Code, consts.CounterpartyErrors.BroadcastError.Description)
		return "", consts.CounterpartyErrors.BroadcastError.Code, errors.New(consts.CounterpartyErrors.BroadcastError.Description)
	}

	database.UpdateAssetCompleteByAssetId(c, accessKey, assetId, txIdSignedTx)

	return txIdSignedTx, 0, nil
}

func DividendCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var dividendStruct enulib.Dividend
	requestId := c.Value(consts.RequestIdKey).(string)
	dividendStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "dividend")

	passphrase := m["passphrase"].(string)
	sourceAddress := m["sourceAddress"].(string)
	asset := m["asset"].(string)
	dividendAsset := m["dividendAsset"].(string)
	quantityPerUnit := uint64(m["quantityPerUnit"].(float64))

	log.FluentfContext(consts.LOGINFO, c, "DividendCreate: received request sourceAddress: %s, asset: %s, dividendAsset: %s, quantityPerUnit: %d from accessKey: %s\n", sourceAddress, asset, dividendAsset, quantityPerUnit, c.Value(consts.AccessKeyKey).(string))

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)

		returnCode := enulib.ReturnCode{RequestId: requestId, Code: -3, Description: err.Error()}
		if err := json.NewEncoder(w).Encode(returnCode); err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
			handlers.ReturnServerError(c, w)

			return nil
		}
		return nil
	}
	log.FluentfContext(consts.LOGINFO, c, "retrieved publickey: %s", sourceAddressPubKey)

	// Generate a dividendId
	dividendId := enulib.GenerateDividendId()
	log.FluentfContext(consts.LOGINFO, c, "Generated dividendId: %s", dividendId)
	dividendStruct.DividendId = dividendId

	dividendStruct.SourceAddress = sourceAddress
	dividendStruct.Asset = asset
	dividendStruct.DividendAsset = dividendAsset
	dividendStruct.QuantityPerUnit = quantityPerUnit

	// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(dividendStruct); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	// Start dividend creation in async mode
	go delegatedCreateDividend(c, c.Value(consts.AccessKeyKey).(string), passphrase, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit)

	return nil
}

// Concurrency safe to create and send transactions from a single address.
func delegatedCreateDividend(c context.Context, accessKey string, passphrase string, dividendId string, sourceAddress string, asset string, dividendAsset string, quantityPerUnit uint64) (string, int64, error) {
	// Write the dividend with the generated dividend id to the database
	go database.InsertDividend(accessKey, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, "valid")

	sourceAddressPubKey, err := counterpartycrypto.GetPublicKey(passphrase, sourceAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Err in GetPublicKey(): %s\n", err.Error())
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, consts.CounterpartyErrors.InvalidPassphrase.Code, consts.CounterpartyErrors.InvalidPassphrase.Description)
		return "", consts.CounterpartyErrors.InvalidPassphrase.Code, errors.New(consts.CounterpartyErrors.InvalidPassphrase.Description)
	}

	// Mutex lock this address
	counterparty_Mutexes.Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked the map") // The map of mutexes must be locked before we modify the mutexes stored in the map

	// If an entry doesn't currently exist in the map for that address
	if counterparty_Mutexes.m[sourceAddress] == nil {
		log.FluentfContext(consts.LOGINFO, c, "Created new entry in map for %s\n", sourceAddress)
		counterparty_Mutexes.m[sourceAddress] = new(sync.Mutex)
	}

	counterparty_Mutexes.m[sourceAddress].Lock()
	log.FluentfContext(consts.LOGINFO, c, "Locked: %s", sourceAddress)

	// We must sleep for at least the time it takes for any transactions to propagate through to the counterparty mempool
	log.FluentfContext(consts.LOGINFO, c, "Sleeping")
	time.Sleep(time.Duration(counterparty_BackEndPollRate+3000) * time.Millisecond)

	defer counterparty_Mutexes.Unlock()
	defer counterparty_Mutexes.m[sourceAddress].Unlock()

	// Create the dividend
	createResult, errorCode, err := counterpartyapi.CreateDividend(c, sourceAddress, asset, dividendAsset, quantityPerUnit, sourceAddressPubKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in CreateDividend(): %s errorCode: %d", err.Error(), errorCode)
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, consts.CounterpartyErrors.ComposeError.Code, consts.CounterpartyErrors.ComposeError.Description)
		return "", errorCode, err
	}

	log.FluentfContext(consts.LOGINFO, c, "Created dividend of %d %s for each %s from address %s: %s\n", quantityPerUnit, dividendAsset, asset, sourceAddress, createResult)

	// Sign the transactions
	signed, err := counterpartyapi.SignRawTransaction(c, passphrase, createResult)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in SignRawTransaction: %s", err.Error())
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, consts.CounterpartyErrors.SigningError.Code, consts.CounterpartyErrors.SigningError.Description)
		return "", consts.CounterpartyErrors.SigningError.Code, errors.New(consts.CounterpartyErrors.SigningError.Description)
	}

	log.FluentfContext(consts.LOGINFO, c, "Signed tx: %s", signed)

	//	 Transmit the transaction if not in dev, otherwise stub out the return
	txIdSignedTx, err := bitcoinapi.SendRawTransaction(c, signed)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in SendRawTransaction(): %s", err.Error())
		database.UpdateDividendWithErrorByDividendId(c, accessKey, dividendId, consts.CounterpartyErrors.BroadcastError.Code, consts.CounterpartyErrors.BroadcastError.Description)
		return "", consts.CounterpartyErrors.BroadcastError.Code, errors.New(consts.CounterpartyErrors.BroadcastError.Description)
	}

	database.UpdateDividendCompleteByDividendId(c, accessKey, dividendId, txIdSignedTx)

	return txIdSignedTx, 0, nil
}

func AssetIssuances(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var issuanceForAsset enulib.AssetIssuances

	requestId := c.Value(consts.RequestIdKey).(string)
	issuanceForAsset.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	asset := vars["asset"]

	if asset == "" || len(asset) < 5 {
		log.FluentfContext(consts.LOGERROR, c, "Invalid asset")
		handlers.ReturnBadRequest(c, w, consts.GenericErrors.InvalidAsset.Code, consts.GenericErrors.InvalidAsset.Description)

		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "AssetIssuances: received request asset: %s from accessKey: %s\n", asset, c.Value(consts.AccessKeyKey).(string))
	result, errorCode, err := counterpartyapi.GetIssuances(c, asset)
	if err != nil {
		handlers.ReturnServerErrorWithCustomError(c, w, errorCode, err.Error())

		return nil
	}

	// Iterate and gather the balances to return
	issuanceForAsset.Asset = asset

	if len(result) > 0 {
		if result[0].Divisible == 1 { // the first valid issuance always defines divisibility
			issuanceForAsset.Divisible = true
			issuanceForAsset.Divisibility = 100000000 // always divisible to 8 decimal places for counterparty divisible assets
		} else {
			issuanceForAsset.Divisible = false
			issuanceForAsset.Divisibility = 0
		}

		// If any issuances has locked the asset, then the supply of the asset is locked
		var isLocked = false
		for _, i := range result {
			if i.Locked == 1 {
				isLocked = true
			}
		}

		issuanceForAsset.Description = result[len(result)-1].Description // get the last description on the asset
		issuanceForAsset.Locked = isLocked
	}

	for _, item := range result {
		var issuance enulib.Issuance

		issuance.BlockIndex = item.BlockIndex
		issuance.Issuer = item.Issuer
		issuance.Quantity = item.Quantity
		//		issuance.Transfer = item.Transfer

		issuanceForAsset.Issuances = append(issuanceForAsset.Issuances, issuance)
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(issuanceForAsset); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	return nil
}

// Recommended call which summarises the ledger for a particular asset
func AssetLedger(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var assetBalances enulib.AssetBalances

	requestId := c.Value(consts.RequestIdKey).(string)
	assetBalances.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	asset := vars["asset"]

	if asset == "" || len(asset) < 5 {
		log.FluentfContext(consts.LOGERROR, c, "Invalid asset")
		handlers.ReturnBadRequest(c, w, consts.GenericErrors.InvalidAsset.Code, consts.GenericErrors.InvalidAsset.Description)

		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "AssetLedger: received request asset: %s from accessKey: %s\n", asset, c.Value(consts.AccessKeyKey).(string))

	result, errorCode, err := counterpartyapi.GetBalancesByAsset(c, asset)
	if err != nil {
		handlers.ReturnServerErrorWithCustomError(c, w, errorCode, err.Error())
		return nil
	}

	resultIssuances, errorCode, err := counterpartyapi.GetIssuances(c, asset)
	if err != nil {
		handlers.ReturnServerErrorWithCustomError(c, w, errorCode, err.Error())
		return nil
	}

	// Summarise asset information
	// Calculate supply
	for _, issuanceItem := range resultIssuances {
		assetBalances.Supply += issuanceItem.Quantity
	}

	if len(resultIssuances) > 0 {
		if resultIssuances[0].Divisible == 1 { // the first valid issuance always defines divisibility
			assetBalances.Divisible = true
			assetBalances.Divisibility = 100000000 // always divisible to 8 decimal places for counterparty divisible assets
		} else {
			assetBalances.Divisible = false
			assetBalances.Divisibility = 1
		}

		// If any issuances has locked the asset, then the supply of the asset is locked
		assetBalances.Locked = false
		for _, i := range resultIssuances {
			if i.Locked == 1 {
				assetBalances.Locked = true
			}
		}

		assetBalances.Description = resultIssuances[len(resultIssuances)-1].Description // get the last description on the asset
	}

	// Iterate and gather the balances to return
	assetBalances.Asset = asset
	//	assetBalances.Supply = supply
	//	assetBalances.Divisible = divisible
	//	assetBalances.Divisibility = divisibility
	//	assetBalances.Locked = locked
	//	assetBalances.Description = description
	for _, item := range result {
		var balance enulib.AddressAmount
		var percentage float64

		percentage = float64(item.Quantity) / float64(assetBalances.Supply) * 100

		balance.Address = item.Address
		balance.Quantity = item.Quantity
		balance.PercentageHolding = percentage

		assetBalances.Balances = append(assetBalances.Balances, balance)
	}

	if err = json.NewEncoder(w).Encode(assetBalances); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func GetDividend(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	var dividend enulib.Dividend
	requestId := c.Value(consts.RequestIdKey).(string)
	dividend.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	dividendId := vars["dividendId"]

	if dividendId == "" || len(dividendId) < 16 {
		log.FluentfContext(consts.LOGERROR, c, "Invalid dividendId")
		handlers.ReturnBadRequest(c, w, consts.GenericErrors.InvalidDividendId.Code, consts.GenericErrors.InvalidDividendId.Description)

		return nil

	}

	log.FluentfContext(consts.LOGINFO, c, "GetDividend called for '%s' by '%s'\n", dividendId, c.Value(consts.AccessKeyKey).(string))

	dividend, err := database.GetDividendByDividendId(c, c.Value(consts.AccessKeyKey).(string), dividendId)
	if err != nil {

	}
	dividend.RequestId = requestId

	// Add the blockchain status
	if dividend.BroadcastTxId == "" {
		dividend.BlockchainStatus = "unconfimed"
		dividend.BlockchainConfirmations = 0
	}
	if dividend.BroadcastTxId != "" {
		confirmations, err := bitcoinapi.GetConfirmations(dividend.BroadcastTxId)
		if err == nil || confirmations == 0 {
			dividend.BlockchainStatus = "unconfimed"
			dividend.BlockchainConfirmations = 0
		}

		dividend.BlockchainStatus = "confirmed"
		dividend.BlockchainConfirmations = confirmations
	}

	if err := json.NewEncoder(w).Encode(dividend); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
