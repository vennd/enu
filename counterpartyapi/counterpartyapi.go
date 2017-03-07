// Contains API to counterparty functions
// Regarding errorhandling, if a lower level function returns an errorCode, propagate the error back upwards
// If the function handling the error is not exposed directly to the HTTP handlers, it's better that the original error is propagated to preserve the error

package counterpartyapi

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mxk/go-sqlite/sqlite3"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/counterpartycrypto"
	"github.com/vennd/enu/log"

	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/btcec"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/chaincfg"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/txscript"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/wire"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcutil"
	"github.com/vennd/enu/internal/github.com/gorilla/securecookie"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var Counterparty_DefaultDustSize uint64 = 5430
var Counterparty_DefaultTxFee uint64 = 10000       // in satoshis
var Counterparty_DefaultTestingTxFee uint64 = 1500 // in satoshis
var numericAssetIdMinString = "95428956661682176"
var numericAssetIdMaxString = "18446744073709551616"

type payloadGetBalances struct {
	Method  string                   `json:"method"`
	Params  payloadGetBalancesParams `json:"params"`
	Jsonrpc string                   `json:"jsonrpc"`
	Id      uint32                   `json:"id"`
}

type payloadGetBalancesParams struct {
	Filters  filters `json:"filters"`
	FilterOp string  `json:"filterop"`
}

type filters []filter

type filter struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type ResultGetBalances struct {
	Jsonrpc string    `json:"jsonrpc"`
	Id      uint32    `json:"id"`
	Result  []Balance `json:"result"`
}

type Balance struct {
	Quantity uint64 `json:"quantity"`
	Asset    string `json:"asset"`
	Address  string `json:"address"`
}

// Struct definitions for creating a send Counterparty transaction
type payloadCreateSend_Counterparty struct {
	Method  string                               `json:"method"`
	Params  payloadCreateSendParams_Counterparty `json:"params"`
	Jsonrpc string                               `json:"jsonrpc"`
	Id      uint32                               `json:"id"`
}

//  myParams = ["source":sourceAddress,"destination":destinationAddress,"asset":asset,"quantity":amount,"allow_unconfirmed_inputs":true,"encoding":counterpartyTransactionEncoding,"pubkey":pubkey]
type payloadCreateSendParams_Counterparty struct {
	Source                 string `json:"source"`
	Destination            string `json:"destination"`
	Asset                  string `json:"asset"`
	Quantity               uint64 `json:"quantity"`
	AllowUnconfirmedInputs string `json:"allow_unconfirmed_inputs"`
	Encoding               string `json:"encoding"`
	PubKey                 string `json:"pubkey"`
	Fee                    uint64 `json:"fee"`
	DustSize               uint64 `json:"regular_dust_size"`
}

type ResultCreateSend_Counterparty struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
	Result  string `json:"result"`
}

type payloadCreateIssuance_Counterparty struct {
	Method  string                                   `json:"method"`
	Params  payloadCreateIssuanceParams_Counterparty `json:"params"`
	Jsonrpc string                                   `json:"jsonrpc"`
	Id      uint32                                   `json:"id"`
}

type payloadCreateIssuanceParams_Counterparty struct {
	Source      string `json:"source"`
	Quantity    uint64 `json:"quantity"`
	Asset       string `json:"asset"`
	Divisible   bool   `json:"divisible"`
	Description string `json:"description"`
	//	TransferDestination    string `json:"transfer_destination"`
	Encoding               string `json:"encoding"`
	PubKey                 string `json:"pubkey"`
	AllowUnconfirmedInputs string `json:"allow_unconfirmed_inputs"`
	Fee                    uint64 `json:"fee"`
	DustSize               uint64 `json:"regular_dust_size"`
}

type ResultCreateIssuance_Counterparty struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
	Result  string `json:"result"`
}

type payloadCreateDividend_Counterparty struct {
	Method  string                                   `json:"method"`
	Params  payloadCreateDividendParams_Counterparty `json:"params"`
	Jsonrpc string                                   `json:"jsonrpc"`
	Id      uint32                                   `json:"id"`
}

type payloadCreateDividendParams_Counterparty struct {
	Source                 string `json:"source"`
	Asset                  string `json:"asset"`
	DividendAsset          string `json:"dividend_asset"`
	QuantityPerUnit        uint64 `json:"quantity_per_unit"`
	Encoding               string `json:"encoding"`
	PubKey                 string `json:"pubkey"`
	AllowUnconfirmedInputs string `json:"allow_unconfirmed_inputs"`
	Fee                    uint64 `json:"fee"`
	DustSize               uint64 `json:"regular_dust_size"`
}

type ResultCreateDividend_Counterparty struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
	Result  string `json:"result"`
}

type ResultError_Counterparty struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
	Error   string `json:"error"`
}

type payloadGetIssuances struct {
	Method  string                    `json:"method"`
	Params  payloadGetIssuancesParams `json:"params"`
	Jsonrpc string                    `json:"jsonrpc"`
	Id      uint32                    `json:"id"`
}

type payloadGetIssuancesParams struct {
	OrderBy  string  `json:"order_by"`
	OrderDir string  `json:"order_dir"`
	Filters  filters `json:"filters"`
}

type ResultGetIssuances struct {
	Jsonrpc string     `json:"jsonrpc"`
	Id      uint32     `json:"id"`
	Result  []Issuance `json:"result"`
}

// Create wrapper for http response and error
type ApiResult struct {
	resp *http.Response
	err  error
}

type Issuance struct {
	TxIndex     uint64 `json:"tx_index"`
	TxHash      string `json:"tx_hash"`
	BlockIndex  uint64 `json:"block_index"`
	Asset       string `json:"asset"`
	Quantity    uint64 `json:"quantity"`
	Divisible   uint64 `json:"divisible"`
	Source      string `json:"source"`
	Issuer      string `json:"issuer"`
	Transfer    uint64 `json:"transfer"`
	Description string `json:"description"`
	FeePaid     uint64 `json:"fee_paid"`
	Locked      uint64 `json:"locked"`
	Status      string `json:"status"`
}

type payloadGetRunningInfo struct {
	Method  string `json:"method"`
	Jsonrpc string `json:"jsonrpc"`
	Id      uint32 `json:"id"`
}

type ResultGetRunningInfo struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      uint32      `json:"id"`
	Result  RunningInfo `json:"result"`
}

type LastBlock struct {
	BlockIndex uint64 `json:"block_index"`
	BlockHash  string `json:"block_hash"`
}

type RunningInfo struct {
	DbCaughtUp           bool      `json:"db_caught_up"`
	BitCoinBlockCount    uint64    `json:"bitcoin_block_count"`
	CounterpartydVersion string    `json:"counterpartyd_version"`
	LastMessageIndex     uint64    `json:"last_message_index"`
	RunningTestnet       bool      `json:"running_testnet"`
	LastBlock            LastBlock `json:"last_block"`
}

//tx_index (integer): The transaction index
//tx_hash (string): The transaction hash
//block_index (integer): The block index (block number in the block chain)
//source (string): The source address of the send
//destination (string): The destination address of the send
//asset (string): The assets being sent
//quantity (integer): The quantities of the specified asset sent
//validity (string): Set to “valid” if a valid send. Any other setting signifies an invalid/improper send
type ResultGetSends struct {
	TxIndex     uint64 `json:"tx_index"`
	TxHash      string `json:"tx_hash"`
	BlockIndex  uint64 `json:"block_index"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Asset       string `json:"asset"`
	Quantity    uint64 `json:"quantity"`
	Validity    uint64 `json:"validity"`
	Status      string `json:"status"`
}

// Globals
var isInit bool = false // set to true only after the init sequence is complete
var counterpartyHost string
var counterpartyUser string
var counterpartyPassword string
var counterpartyTransactionEncoding string
var counterpartyDBLocation string

// Initialises global variables and database connection for all handlers
func Init() {
	var configFilePath string

	if isInit == true {
		return
	}

	if _, err := os.Stat("./enuapi.json"); err == nil {
		//		log.Println("Found and using configuration file ./enuapi.json")
		configFilePath = "./enuapi.json"
	} else {
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/enuapi.json"); err == nil {
			configFilePath = os.Getenv("GOPATH") + "/bin/enuapi.json"
			//			log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

		} else {
			if _, err := os.Stat(os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"); err == nil {
				configFilePath = os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"
				//				log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)
			} else {
				log.Println("Cannot find enuapi.json")
				os.Exit(-100)
			}
		}
	}

	InitWithConfigPath(configFilePath)
}

func InitWithConfigPath(configFilePath string) {
	var configuration interface{}

	if isInit == true {
		return
	}

	// Read configuration from file
	//	log.Printf("Reading %s\n", configFilePath)
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Println("Unable to read configuration file enuapi.json")
		log.Println(err.Error())
		os.Exit(-101)
	}

	err = json.Unmarshal(file, &configuration)

	if err != nil {
		log.Println("Unable to parse enuapi.json")
		log.Println(err.Error())
		os.Exit(-101)
	}

	m := configuration.(map[string]interface{})

	// Counterparty API parameters
	counterpartyHost = m["counterpartyhost"].(string)                               // End point for JSON RPC server
	counterpartyUser = m["counterpartyuser"].(string)                               // Basic authentication user name
	counterpartyPassword = m["counterpartypassword"].(string)                       // Basic authentication password
	counterpartyTransactionEncoding = m["counterpartytransactionencoding"].(string) // The encoding that should be used for Counterparty transactions "auto" will let Counterparty select, valid values "multisig", "opreturn"
	counterpartyDBLocation = m["counterpartydblocation"].(string)                   // Direct location of counterpartydb if we can't reach the API

	isInit = true
}

// Posts to the given counterparty JSON RPC call. Returns a map[string]interface{} which has already unmarshalled the JSON result
// Attempts to interpret the counterparty errors such that the caller doesn't need to work out what is going on
func postAPI(c context.Context, postData []byte) (map[string]interface{}, int64, error) {
	var result map[string]interface{}
	var apiResp ApiResult

	postDataJson := string(postData)
	//		log.FluentfContext(consts.LOGDEBUG, c, "counterpartyapi postAPI() posting: %s", postDataJson)

	// Set headers
	req, err := http.NewRequest("POST", counterpartyHost, bytes.NewBufferString(postDataJson))
	req.SetBasicAuth(counterpartyUser, counterpartyPassword)
	req.Header.Set("Content-Type", "application/json")

	clientPointer := &http.Client{}

	// Call counterparty JSON service with 10 second timeout
	c1 := make(chan ApiResult, 1)
	go func() {
		var result ApiResult // Wrap the response into a struct so we can return both the error and response

		resp, err := clientPointer.Do(req)
		result.resp = resp
		result.err = err

		c1 <- result
	}()

	select {
	case apiResp = <-c1:
	case <-time.After(time.Second * 10):
		return result, consts.CounterpartyErrors.Timeout.Code, errors.New(consts.CounterpartyErrors.Timeout.Description)
	}

	if apiResp.err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Do(req): %s", apiResp.err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Unsuccessful - ie didn't return HTTP status 200
	if apiResp.resp.StatusCode != 200 {
		log.FluentfContext(consts.LOGDEBUG, c, "Request didn't return a 200. Status code: %d\n", apiResp.resp.StatusCode)

		// Even though we got an error, counterparty often sends back errors inside the body
		body, err := ioutil.ReadAll(apiResp.resp.Body)
		defer apiResp.resp.Body.Close()
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, err.Error())
			return nil, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
		}

		//		log.FluentfContext(consts.LOGDEBUG, c, "Reply: %s", string(body))

		// Attempt to parse body if not empty in case it is something json-like from counterpartyd
		var errResult map[string]interface{}
		if unmarshallErr := json.Unmarshal(body, &errResult); unmarshallErr != nil {
			// If we couldn't parse the error properly, log error to fluent and return unhandled error
			log.FluentfContext(consts.LOGERROR, c, "Error in Unmarshal(): %s", unmarshallErr.Error())
			log.FluentfContext(consts.LOGDEBUG, c, "The body: %s", string(body))

			return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
		}

		if apiResp.resp.StatusCode == 503 {
			log.FluentfContext(consts.LOGDEBUG, c, "Reply: %s", string(body))
		}

		// Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errResult["code"].(float64) == -32000 || errResult["code"].(float64) == -10000 {
			return result, consts.CounterpartyErrors.ReparsingOrUnavailable.Code, errors.New(consts.CounterpartyErrors.ReparsingOrUnavailable.Description)
		}
	}

	// Success, read body and return
	body, err := ioutil.ReadAll(apiResp.resp.Body)
	defer apiResp.resp.Body.Close()

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in ReadAll(): %s", err.Error())
		return nil, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Unmarshall body
	if unmarshallErr := json.Unmarshal(body, &result); unmarshallErr != nil {
		// If we couldn't parse the error properly, log error to fluent and return unhandled error
		log.FluentfContext(consts.LOGERROR, c, "Error in Unmarshal(): %s", err.Error())

		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// If the body doesn't contain a result then the call must have failed. Attempt to read the payload to work out what happened
	if result["result"] == nil {
		// Uncomment this to work out what the hell counterparty sent back
		//		log.FluentfContext(consts.LOGDEBUG, c, "Result not returned from counterpartyd. Got: %s", fmt.Sprintf("%#v", result))

		// Got an error
		if result["error"] != nil {
			var dataMap map[string]interface{}
			errorMap := result["error"].(map[string]interface{})

			if errorMap["data"] != nil {
				dataMap = errorMap["data"].(map[string]interface{})
			}

			// Only issuer can pay dividends
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CounterpartylibOnlyIssuerCanPayDividends) {
				return result, consts.CounterpartyErrors.OnlyIssuerCanPayDividends.Code, errors.New(consts.CounterpartyErrors.OnlyIssuerCanPayDividends.Description)
			}

			// Insufficient asset in address
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CounterpartylibInsufficientFunds) {
				return result, consts.CounterpartyErrors.InsufficientFunds.Code, errors.New(consts.CounterpartyErrors.InsufficientFunds.Description)
			}

			// Bad address
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CounterpartylibMalformedAddress) {
				return result, consts.CounterpartyErrors.MalformedAddress.Code, errors.New(consts.CounterpartyErrors.MalformedAddress.Description)
			}

			// No such asset
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CountpartylibNoSuchAsset) {
				return result, consts.CounterpartyErrors.NoSuchAsset.Code, errors.New(consts.CounterpartyErrors.NoSuchAsset.Description)
			}

			//Insufficient BTC at address
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CounterpartylibInsufficientBTC) {
				return result, consts.CounterpartyErrors.InsufficientFees.Code, errors.New(consts.CounterpartyErrors.InsufficientFees.Description)
			}

			//Counterparty is just restarting now
			if dataMap["message"] != nil && strings.Contains(dataMap["message"].(string), consts.CountpartylibMempoolIsNotReady) {
				return result, consts.CounterpartyErrors.ReparsingOrUnavailable.Code, errors.New(consts.CounterpartyErrors.ReparsingOrUnavailable.Description)
			}
		}

		log.FluentfContext(consts.LOGDEBUG, c, "Counterparty returned an error in the body but returned a HTTP status of 200. Got: %s", fmt.Sprintf("%#v", result))

		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	return result, 0, nil
}

func generateId(c context.Context) uint32 {
	buf := securecookie.GenerateRandomKey(4)
	randomUint64, err := strconv.ParseUint(hex.EncodeToString(buf), 16, 32)

	if err != nil {
		panic(err)
	}

	randomUint32 := uint32(randomUint64)

	return randomUint32
}

func GetBalancesByAddress(c context.Context, address string) ([]Balance, int64, error) {
	var payload payloadGetBalances
	var result []Balance

	if isInit == false {
		Init()
	}

	filterCondition := filter{Field: "address", Op: "==", Value: address}

	payload.Method = "get_balances"
	payload.Params.Filters = append(payload.Params.Filters, filterCondition)
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		// Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errorCode == consts.CounterpartyErrors.ReparsingOrUnavailable.Code || errorCode == consts.CounterpartyErrors.Timeout.Code {
			return GetBalancesByAddressDB(c, address)
		}

		return result, errorCode, err
	}

	// Range over the result from api and create the reply
	if responseData["result"] != nil {
		for _, b := range responseData["result"].([]interface{}) {
			c := b.(map[string]interface{})
			result = append(result,
				Balance{Address: c["address"].(string),
					Asset:    c["asset"].(string),
					Quantity: uint64(c["quantity"].(float64))})
		}
	}

	return result, 0, nil
}

func GetBalancesByAddressDB(c context.Context, address string) ([]Balance, int64, error) {
	var result []Balance

	// sqlite drivers are not concurrency safe, so must create a connection each time
	db, err := sql.Open("sqlite3", counterpartyDBLocation)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to open DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	err = db.Ping()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to ping DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	//	 Query DB
	//	log.Fluentf(consts.LOGDEBUG, "select address, asset, quantity from balances where address = %s", address)
	stmt, err := db.Prepare("select address, asset, quantity from balances where address = ?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}
	defer stmt.Close()

	//	 Get row
	rows, err := stmt.Query(address)
	defer rows.Close()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to query. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	for rows.Next() {
		var balance = Balance{}
		var address []byte
		var asset []byte
		var quantity uint64

		if err := rows.Scan(&address, &asset, &quantity); err == sql.ErrNoRows {
			if err.Error() == "sql: no rows in result set" {
			}
		} else if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		} else {
			balance = Balance{Address: string(address), Asset: string(asset), Quantity: quantity}
		}

		result = append(result, balance)
	}

	return result, 0, nil
}

func GetBalancesByAsset(c context.Context, asset string) ([]Balance, int64, error) {
	var payload payloadGetBalances
	var result []Balance

	if isInit == false {
		Init()
	}

	filterCondition := filter{Field: "asset", Op: "==", Value: asset}

	payload.Method = "get_balances"
	payload.Params.Filters = append(payload.Params.Filters, filterCondition)
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		// Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errorCode == consts.CounterpartyErrors.ReparsingOrUnavailable.Code || errorCode == consts.CounterpartyErrors.Timeout.Code {
			return GetBalancesByAssetDB(c, asset)
		}

		return result, errorCode, err
	}

	// Range over the result from api and create the reply
	if responseData["result"] != nil {
		for _, b := range responseData["result"].([]interface{}) {
			c := b.(map[string]interface{})
			result = append(result,
				Balance{Address: c["address"].(string),
					Asset:    c["asset"].(string),
					Quantity: uint64(c["quantity"].(float64))})
		}
	}

	return result, 0, nil
}

func GetBalancesByAssetDB(c context.Context, asset string) ([]Balance, int64, error) {
	var result []Balance

	// sqlite drivers are not concurrency safe, so must create a connection each time
	db, err := sql.Open("sqlite3", counterpartyDBLocation)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to open DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	err = db.Ping()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to ping DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	//	 Query DB
	//	log.Fluentf(consts.LOGDEBUG, "select address, asset, quantity from balances where asset = %s", asset)
	stmt, err := db.Prepare("select address, asset, quantity from balances where asset = ?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}
	defer stmt.Close()

	//	 Get row
	rows, err := stmt.Query(asset)
	defer rows.Close()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to query. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	for rows.Next() {
		var balance = Balance{}
		var address []byte
		var asset []byte
		var quantity uint64

		if err := rows.Scan(&address, &asset, &quantity); err == sql.ErrNoRows {
			if err.Error() == "sql: no rows in result set" {
			}
		} else if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		} else {
			balance = Balance{Address: string(address), Asset: string(asset), Quantity: quantity}
		}

		result = append(result, balance)
	}

	return result, 0, nil
}

func GetSendsByAddress(c context.Context, address string) ([]ResultGetSends, int64, error) {
	var payload = make(map[string]interface{})
	var params = make(map[string]interface{})
	var result []ResultGetSends

	if isInit == false {
		Init()
	}

	var filterArray filters
	filterCondition := filter{Field: "destination", Op: "==", Value: address}
	filterArray = append(filterArray, filterCondition)

	params["filters"] = filterArray
	params["filterop"] = "OR"
	params["status"] = "valid"

	payload["method"] = "get_sends"
	payload["params"] = params
	payload["jsonrpc"] = "2.0"
	payload["id"] = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		//		 Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errorCode == consts.CounterpartyErrors.ReparsingOrUnavailable.Code || errorCode == consts.CounterpartyErrors.Timeout.Code {
			return GetSendsByAddressDB(c, address)
		}

		return result, errorCode, err
	}

	// Range over the result from api and create the reply
	if responseData["result"] != nil {
		for _, b := range responseData["result"].([]interface{}) {
			c := b.(map[string]interface{})
			result = append(result,
				ResultGetSends{Source: c["source"].(string),
					Destination: c["destination"].(string),
					Asset:       c["asset"].(string),
					Quantity:    uint64(c["quantity"].(float64)),
					BlockIndex:  uint64(c["block_index"].(float64)),
					TxHash:      string(c["tx_hash"].(string)),
					TxIndex:     uint64(c["tx_index"].(float64))})
		}
	}

	return result, 0, nil
}

func GetSendsByAddressDB(c context.Context, address string) ([]ResultGetSends, int64, error) {
	var result []ResultGetSends

	// sqlite drivers are not concurrency safe, so must create a connection each time
	db, err := sql.Open("sqlite3", counterpartyDBLocation)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to open DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	err = db.Ping()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to ping DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	//	 Query DB
	//	log.Fluentf(consts.LOGDEBUG, "select tx_index, tx_hash, block_index, source, destination, asset, quantity, status from sends where destination = %s", address)
	stmt, err := db.Prepare("select tx_index, tx_hash, block_index, source, destination, asset, quantity, status from sends where destination = ?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}
	defer stmt.Close()

	//	 Get row
	rows, err := stmt.Query(address)
	defer rows.Close()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to query. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	for rows.Next() {
		var send = ResultGetSends{}
		var txindex uint64
		var txhash []byte
		var blockindex uint64
		var source []byte
		var destination []byte
		var asset []byte
		var quantity uint64
		var status []byte

		if err := rows.Scan(&txindex, &txhash, &blockindex, &source, &destination, &asset, &quantity, &status); err == sql.ErrNoRows {
			if err.Error() == "sql: no rows in result set" {
			}
		} else if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		} else {
			send = ResultGetSends{TxIndex: txindex, TxHash: string(txhash), BlockIndex: blockindex, Source: string(source), Destination: string(destination), Asset: string(asset), Quantity: quantity, Status: string(status)}
		}

		result = append(result, send)
	}

	return result, 0, nil
}

func GetIssuances(c context.Context, asset string) ([]Issuance, int64, error) {
	var payload payloadGetIssuances
	var result []Issuance

	if isInit == false {
		Init()
	}
	filterCondition := filter{Field: "asset", Op: "==", Value: asset}
	filterCondition2 := filter{Field: "status", Op: "==", Value: "valid"}

	payload.Method = "get_issuances"
	payload.Params.OrderBy = "tx_index"
	payload.Params.OrderDir = "asc"
	payload.Params.Filters = append(payload.Params.Filters, filterCondition)
	payload.Params.Filters = append(payload.Params.Filters, filterCondition2)
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		// Counterparty DB is behind backend / reparsing or timed out, read directly from DB
		if errorCode == consts.CounterpartyErrors.ReparsingOrUnavailable.Code || errorCode == consts.CounterpartyErrors.Timeout.Code {
			return GetIssuancesDB(c, asset)
		}

		return result, errorCode, err
	}

	// Range over the result from api and create the reply
	if responseData["result"] != nil {
		for _, b := range responseData["result"].([]interface{}) {
			c := b.(map[string]interface{})
			result = append(result,
				Issuance{TxIndex: uint64(c["tx_index"].(float64)),
					TxHash:      c["tx_hash"].(string),
					BlockIndex:  uint64(c["block_index"].(float64)),
					Asset:       c["asset"].(string),
					Quantity:    uint64(c["quantity"].(float64)),
					Divisible:   uint64(c["divisible"].(float64)),
					Source:      c["source"].(string),
					Issuer:      c["issuer"].(string),
					Transfer:    uint64(c["transfer"].(float64)),
					Description: c["description"].(string),
					FeePaid:     uint64(c["fee_paid"].(float64)),
					Locked:      uint64(c["locked"].(float64)),
					Status:      c["status"].(string)})
		}
	}

	return result, 0, nil
}

func GetIssuancesDB(c context.Context, asset string) ([]Issuance, int64, error) {
	var result []Issuance

	// sqlite drivers are not concurrency safe, so must create a connection each time
	db, err := sql.Open("sqlite3", counterpartyDBLocation)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to open DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	err = db.Ping()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to ping DB. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	//	 Query DB
	//	log.Fluentf(consts.LOGDEBUG, "select tx_index, tx_hash, block_index, asset, quantity, divisible, source, issuer, transfer, description, fee_paid, locked, status from issuances where status='valid' and asset=%s", asset)
	stmt, err := db.Prepare("select tx_index, tx_hash, block_index, asset, quantity, divisible, source, issuer, transfer, description, fee_paid, locked, status from issuances where status='valid' and asset=?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}
	defer stmt.Close()

	//	 Get row
	rows, err := stmt.Query(asset)
	defer rows.Close()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to query. Reason: %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	for rows.Next() {
		var issuance = Issuance{}
		var tx_index uint64
		var tx_hash []byte
		var block_index uint64
		var asset []byte
		var quantity uint64
		var divisible []byte // returned as a string from the DB driver, we need to return as an int
		var source []byte
		var issuer []byte
		var transfer []byte // returned as a string from the DB driver, we need to return as an int
		var description []byte
		var fee_paid uint64
		var locked []byte // returned as a string from the DB driver, we need to return as an int
		var status []byte

		if err := rows.Scan(&tx_index, &tx_hash, &block_index, &asset, &quantity, &divisible, &source, &issuer, &transfer, &description, &fee_paid, &locked, &status); err == sql.ErrNoRows {
			if err.Error() == "sql: no rows in result set" {
			}
		} else if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		} else {
			var divisibleResult uint64
			if string(divisible) == "true" {
				divisibleResult = 1
			} else {
				divisibleResult = 0
			}

			var transferResult uint64
			if string(transfer) == "true" {
				transferResult = 1
			} else {
				transferResult = 0
			}

			var lockedResult uint64
			if string(locked) == "true" {
				lockedResult = 1
			} else {
				lockedResult = 0
			}

			issuance = Issuance{TxIndex: tx_index, TxHash: string(tx_hash), BlockIndex: block_index, Asset: string(asset), Quantity: quantity, Divisible: divisibleResult, Source: string(source), Issuer: string(issuer), Transfer: transferResult, Description: string(description), FeePaid: fee_paid, Locked: lockedResult, Status: string(status)}
		}

		result = append(result, issuance)
	}

	return result, 0, nil
}

// Generates a hex string serialed tx which contains the bitcoin transaction to send an asset from sourceAddress to destinationAddress
// Not exposed to the public
func CreateSend(c context.Context, sourceAddress string, destinationAddress string, asset string, quantity uint64, pubKeyHexString string) (string, int64, error) {
	var payload payloadCreateSend_Counterparty
	var result string

	if isInit == false {
		Init()
	}

	//	log.Println("In counterpartyapi.CreateSend()")

	// ["source":sourceAddress,"destination":destinationAddress,"asset":asset,"quantity":amount,"allow_unconfirmed_inputs":true,"encoding":counterpartyTransactionEncoding,"pubkey":pubkey]
	payload.Method = "create_send"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)
	payload.Params.Source = sourceAddress
	payload.Params.Destination = destinationAddress
	payload.Params.Asset = asset
	payload.Params.Quantity = quantity
	payload.Params.AllowUnconfirmedInputs = "true"
	payload.Params.Encoding = counterpartyTransactionEncoding
	payload.Params.PubKey = pubKeyHexString
	payload.Params.Fee = Counterparty_DefaultTxFee
	payload.Params.DustSize = Counterparty_DefaultDustSize

	// Marshal into json
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Post the request to counterpartyd
	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return "", errorCode, err
	}

	// Get the result
	if responseData["result"] != nil {
		result = responseData["result"].(string)
	}

	return result, 0, nil
}

// When given the 12 word passphrase:
// 1) Parses the raw TX to find the address being sent from
// 2) Derives the parent key and the child key for the address found in step 1)
// 3) Signs all the TX inputs
//
// Assumptions
// 1) This is a Counterparty transaction so all inputs need to be signed with the same pubkeyhash
func SignRawTransaction(c context.Context, passphrase string, rawTxHexString string) (string, error) {
	// Convert the hex string to a byte array
	txBytes, err := hex.DecodeString(rawTxHexString)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in DecodeString(): %s", err.Error())
		return "", err
	}

	//	log.Printf("Unsigned tx: %s", rawTxHexString)

	// Deserialise the transaction
	tx, err := btcutil.NewTxFromBytes(txBytes)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in NewTxFromBytes(): %s", err.Error())
		return "", err
	}
	//	log.Printf("Deserialised ok!: %+v", tx)

	msgTx := tx.MsgTx()
	redeemTx := wire.NewMsgTx() // Create a new transaction and copy the details from the tx that was serialised. For some reason BTCD can't sign in place transactions
	//	log.Printf("MsgTx: %+v", msgTx)
	//	log.Printf("Number of txes in: %d\n", len(msgTx.TxIn))
	for i := 0; i <= len(msgTx.TxIn)-1; i++ {
		//		log.Printf("MsgTx.TxIn[%d]:\n", i)
		//		log.Printf("TxIn[%d].PreviousOutPoint.Hash: %s\n", i, msgTx.TxIn[i].PreviousOutPoint.Hash)
		//		log.Printf("TxIn[%d].PreviousOutPoint.Index: %d\n", i, msgTx.TxIn[i].PreviousOutPoint.Index)
		//		log.Printf("TxIn[%d].SignatureScript: %s\n", i, hex.EncodeToString(msgTx.TxIn[i].SignatureScript))
		script := msgTx.TxIn[i].SignatureScript

		// Following block is for debugging only
		//		disasm, err := txscript.DisasmString(script)
		//		if err != nil {
		//			return "", err
		//		}
		//		log.Printf("TxIn[%d] Script Disassembly: %s", i, disasm)

		// Extract and print details from the script.
		// next line is for debugging only
		//		scriptClass, addresses, reqSigs, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
		scriptClass, _, _, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in ExtractPkScriptAddrs(): %s", err.Error())
			return "", err
		}

		// This function only supports pubkeyhash signing at this time (ie not multisig or P2SH)
		//		log.Printf("TxIn[%d] Script Class: %s\n", i, scriptClass)
		if scriptClass.String() != "pubkeyhash" {
			return "", errors.New("Counterparty_SignRawTransaction() currently only supports pubkeyhash script signing. However, the script type in the TX to sign was: " + scriptClass.String())
		}

		//		log.Printf("TxIn[%d] Addresses: %s\n", i, addresses)
		//		log.Printf("TxIn[%d] Required Signatures: %d\n", i, reqSigs)

		// Build txIn for new redeeming transaction
		prevOut := wire.NewOutPoint(&msgTx.TxIn[i].PreviousOutPoint.Hash, msgTx.TxIn[i].PreviousOutPoint.Index)
		txIn := wire.NewTxIn(prevOut, nil)
		redeemTx.AddTxIn(txIn)
	}

	// Copy TxOuts from serialised tx
	for _, txOut := range msgTx.TxOut {
		out := txOut
		redeemTx.AddTxOut(out)
	}

	// Callback to look up the signing key
	lookupKey := func(a btcutil.Address) (*btcec.PrivateKey, bool, error) {
		address := a.String()

		//		log.Printf("Looking up the private key for: %s\n", address)
		privateKeyString, err := counterpartycrypto.GetPrivateKey(passphrase, address)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in counterpartycrypto.GetPrivateKey(): %s", err.Error())
			return nil, false, nil
		}
		//		log.Printf("Private key retrieved!\n")

		privateKeyBytes, err := hex.DecodeString(privateKeyString)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in DecodeString(): %s", err.Error())
			return nil, false, nil
		}

		privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)

		return privKey, true, nil
	}

	// Range over TxIns and sign
	for i, _ := range redeemTx.TxIn {
		// Get the sigscript
		// Notice that the script database parameter is nil here since it isn't
		// used.  It must be specified when pay-to-script-hash transactions are
		// being signed.
		sigScript, err := txscript.SignTxOutput(&chaincfg.MainNetParams, redeemTx, i, msgTx.TxIn[i].SignatureScript, txscript.SigHashAll, txscript.KeyClosure(lookupKey), nil, nil)

		if err != nil {
			return "", err
		}

		// Copy the signed sigscript into the redeeming tx
		redeemTx.TxIn[i].SignatureScript = sigScript
		//		log.Println(hex.EncodeToString(sigScript))
	}

	// Prove that the transaction has been validly signed by executing the
	// script pair.
	//	log.Println("Checking signature(s)")
	flags := txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures | txscript.ScriptStrictMultiSig | txscript.ScriptDiscourageUpgradableNops | txscript.ScriptVerifyLowS | txscript.ScriptVerifyCleanStack | txscript.ScriptVerifyMinimalData | txscript.ScriptVerifySigPushOnly | txscript.ScriptVerifyStrictEncoding
	var buildError string
	for i, _ := range redeemTx.TxIn {
		vm, err := txscript.NewEngine(msgTx.TxIn[i].SignatureScript, redeemTx, i, flags)
		if err != nil {
			buildError += "NewEngine() error: " + err.Error() + ","
		}

		if err := vm.Execute(); err != nil {
			buildError += "TxIn[" + strconv.Itoa(i) + "]: " + err.Error() + ", "
		} else {
			// Signature verified
			//			log.Printf("TxIn[%d] ok!\n", i)
		}
	}
	if len(buildError) > 0 {
		return "", errors.New(buildError)
	}
	//	log.Println("Transaction successfully signed")

	// Encode the struct into BTC bytes wire format
	var byteBuffer bytes.Buffer
	encodeError := redeemTx.BtcEncode(&byteBuffer, wire.ProtocolVersion)
	if encodeError != nil {
		return "", err
	}

	// Encode bytes to hex string
	payloadBytes := byteBuffer.Bytes()
	payloadHexString := hex.EncodeToString(payloadBytes)
	//	log.Printf("Signed and encoded transaction: %s\n", payloadHexString)

	return payloadHexString, nil
}

// Reproduces counterwallet function to generate a random asset name
// Original JS:
//self.generateRandomId = function() {
//    var r = bigInt.randBetween(NUMERIC_ASSET_ID_MIN, NUMERIC_ASSET_ID_MAX);
//    self.name('A' + r);
//}
func generateRandomAssetName(c context.Context) (string, error) {
	numericAssetIdMin := new(big.Int)
	numericAssetIdMax := new(big.Int)
	//	var err error

	numericAssetIdMin.SetString(numericAssetIdMinString, 10)
	numericAssetIdMax.SetString(numericAssetIdMaxString, 10)

	//	log.Printf("numericAssetIdMax: %s", numericAssetIdMin.String())
	//	log.Printf("numericAssetIdMin: %s", numericAssetIdMax.String())

	numericAssetIdMax = numericAssetIdMax.Add(numericAssetIdMax, numericAssetIdMin)

	x, err := rand.Int(rand.Reader, numericAssetIdMax)
	xFinal := x.Sub(x, numericAssetIdMin)

	if err != nil {
		return "", err
	}

	return "A" + string(xFinal.String()), nil
}

// Generates unsigned hex encoded transaction to issue an asset on Counterparty
// This function MUST NOT be accessed by the client directly. The high level function Counterparty_CreateIssuanceAndSend() should be used instead.
func createIssuance(c context.Context, sourceAddress string, asset string, description string, quantity uint64, divisible bool, pubKeyHexString string) (string, int64, error) {
	var payload payloadCreateIssuance_Counterparty
	var result string

	if isInit == false {
		Init()
	}

	payload.Method = "create_issuance"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)
	payload.Params.Source = sourceAddress
	payload.Params.Asset = asset
	payload.Params.Description = description
	payload.Params.Quantity = quantity
	payload.Params.AllowUnconfirmedInputs = "true"
	payload.Params.Encoding = counterpartyTransactionEncoding
	payload.Params.PubKey = pubKeyHexString
	payload.Params.Fee = Counterparty_DefaultTxFee
	payload.Params.DustSize = Counterparty_DefaultDustSize

	// Marshal into json
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Post the request to counterpartyd
	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return "", errorCode, err
	}

	// Get the result
	if responseData["result"] != nil {
		result = responseData["result"].(string)
	}

	return result, 0, nil
}

func GenerateRandomAssetName(c context.Context) (string, int64, error) {
	if isInit == false {
		Init()
	}

	// Generate random asset name
	var err error
	var randomAssetName string
	randomAssetName, err = generateRandomAssetName(c)

	// If random asset name already exists, keep trying until we find a spare one
	for balance, errorCode, err := GetBalancesByAsset(c, randomAssetName); len(balance) != 0; {
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in GetBalancesByAsset(): %s, errorCode: %d", err.Error(), errorCode)
			return "", errorCode, err
		}
		randomAssetName, err = generateRandomAssetName(c)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in generateRandomAssetName(): %s", err.Error())
			return "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
		}

		balance, errorCode, err = GetBalancesByAsset(c, randomAssetName)
	}

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in after trying to check if asset exists: %s", err.Error())
		return "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	return randomAssetName, 0, nil
}

// Automatically generates a numeric asset name and generates unsigned hex encoded transaction to issue an asset on Counterparty
// Returns:
// randomAssetName that was generated
// hex string encoded transaction
// errorCode
// error
func CreateNumericIssuance(c context.Context, sourceAddress string, asset string, quantity uint64, divisible bool, pubKeyHexString string) (string, string, int64, error) {
	var description string

	if isInit == false {
		Init()
	}

	if len(asset) > 52 {
		description = asset[0:51]
	} else {
		description = asset
	}

	// Generate random asset name
	var err error
	var randomAssetName string
	randomAssetName, err = generateRandomAssetName(c)

	// If random asset name already exists, keep trying until we find a spare one
	for balance, errorCode, err := GetBalancesByAsset(c, randomAssetName); len(balance) != 0; {
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in GetBalancesByAsset(): %s, errorCode: %d", err.Error(), errorCode)
			return "", "", errorCode, err
		}
		randomAssetName, err = generateRandomAssetName(c)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in generateRandomAssetName(): %s", err.Error())
			return "", "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
		}

		balance, errorCode, err = GetBalancesByAsset(c, randomAssetName)
	}

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in after trying to check if asset exists: %s", err.Error())
		return "", "", consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Call counterparty to create the issuance
	result, errorCode, err := createIssuance(c, sourceAddress, randomAssetName, description, quantity, divisible, pubKeyHexString)
	if err != nil {
		return "", "", errorCode, err
	}

	return randomAssetName, result, 0, nil
}

func CreateIssuance(c context.Context, sourceAddress string, asset string, assetDescription string, quantity uint64, divisible bool, pubKeyHexString string) (string, int64, error) {
	if isInit == false {
		Init()
	}

	if len(assetDescription) > 52 {
		assetDescription = assetDescription[0:51]
	}

	// Call counterparty to create the issuance
	result, errorCode, err := createIssuance(c, sourceAddress, asset, assetDescription, quantity, divisible, pubKeyHexString)
	if err != nil {
		return "", errorCode, err
	}

	return result, 0, nil
}

// Generates unsigned hex encoded transaction to pay a dividend on an asset on Counterparty
func CreateDividend(c context.Context, sourceAddress string, asset string, dividendAsset string, quantityPerUnit uint64, pubKeyHexString string) (string, int64, error) {
	var payload payloadCreateDividend_Counterparty
	var result string

	if isInit == false {
		Init()
	}

	payload.Method = "create_dividend"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)
	payload.Params.Source = sourceAddress
	payload.Params.Asset = asset
	payload.Params.DividendAsset = dividendAsset
	payload.Params.QuantityPerUnit = quantityPerUnit
	payload.Params.Encoding = counterpartyTransactionEncoding
	payload.Params.PubKey = pubKeyHexString
	payload.Params.Fee = Counterparty_DefaultTxFee
	payload.Params.DustSize = Counterparty_DefaultDustSize

	// Marshal into json
	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	// Post the request to counterpartyd
	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		return "", errorCode, err
	}

	// Get the result
	if responseData["result"] != nil {
		result = responseData["result"].(string)
	}

	return result, 0, nil
}

// For internal use only - don't expose to customers
func GetRunningInfo(c context.Context) (RunningInfo, int64, error) {
	var payload payloadGetRunningInfo
	var result RunningInfo

	if isInit == false {
		Init()
	}

	payload.Method = "get_running_info"
	payload.Jsonrpc = "2.0"
	payload.Id = generateId(c)

	payloadJsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Marshal(): %s", err.Error())
		return result, consts.CounterpartyErrors.MiscError.Code, errors.New(consts.CounterpartyErrors.MiscError.Description)
	}

	responseData, errorCode, err := postAPI(c, payloadJsonBytes)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in postAPI(): %s", err.Error())
		return result, errorCode, err
	}

	// Get result from api and create the reply
	if responseData["result"] != nil {
		resultMap := responseData["result"].(map[string]interface{})
		lastBlockMap := resultMap["last_block"].(map[string]interface{})
		//		log.Printf("%#v\n", resultMap)
		//		log.Printf("%#v\n", lastBlockMap)
		result = RunningInfo{
			DbCaughtUp:           resultMap["db_caught_up"].(bool),
			BitCoinBlockCount:    uint64(resultMap["bitcoin_block_count"].(float64)),
			CounterpartydVersion: string(uint64(resultMap["version_major"].(float64))) + "." + string(uint64(resultMap["version_minor"].(float64))) + "." + string(uint64(resultMap["version_revision"].(float64))),
			LastMessageIndex:     uint64(resultMap["last_message_index"].(float64)),
			RunningTestnet:       resultMap["running_testnet"].(bool),
			LastBlock: LastBlock{
				BlockIndex: uint64(lastBlockMap["block_index"].(float64)),
				BlockHash:  lastBlockMap["block_hash"].(string),
			},
		}
	}

	return result, 0, nil
}

// Returns the total BTC that is required for the given number of transactions
func CalculateFeeAmount(c context.Context, amount uint64) (uint64, string, error) {
	// Get env and blockchain from context
	env := c.Value(consts.EnvKey).(string)
	blockchainId := c.Value(consts.BlockchainIdKey).(string)

	// Set some maximum and minimums
	var thisAmount = amount
	if thisAmount > 1000 {
		thisAmount = 1000
	}
	if thisAmount < 20 {
		thisAmount = 20
	}

	if blockchainId != consts.CounterpartyBlockchainId {
		errorString := fmt.Sprintf("Blockchain must be %s, got %s", consts.CounterpartyBlockchainId, blockchainId)
		log.FluentfContext(consts.LOGERROR, c, errorString)

		return 0, "", errors.New(errorString)
	}

	var quantity uint64
	var asset string

	if env == "dev" {
		quantity = (Counterparty_DefaultDustSize + Counterparty_DefaultTestingTxFee) * thisAmount
		asset = "BTC"
	} else {
		quantity = (Counterparty_DefaultDustSize + Counterparty_DefaultTxFee) * thisAmount
		asset = "BTC"
	}

	return quantity, asset, nil
}

// Returns the number of transactions that can be performed with the given amount of BTC
// If the env value is not found in the context, calculations are defaulted to production
func CalculateNumberOfTransactions(c context.Context, amount uint64) (uint64, error) {
	// Get env and blockchain from context
	env := c.Value(consts.EnvKey).(string)
	blockchainId := c.Value(consts.BlockchainIdKey).(string)

	if blockchainId != consts.CounterpartyBlockchainId {
		errorString := fmt.Sprintf("Blockchain must be %s, got %s", consts.CounterpartyBlockchainId, blockchainId)
		log.FluentfContext(consts.LOGERROR, c, errorString)

		return 0, errors.New(errorString)
	}

	if env == "dev" {
		return amount / (Counterparty_DefaultDustSize + Counterparty_DefaultTestingTxFee), nil
	} else {
		return amount / (Counterparty_DefaultDustSize + Counterparty_DefaultTxFee), nil
	}
}
