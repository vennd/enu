package bitcoinapi

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/log"

	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/btcjson"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcd/wire"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcrpcclient"
	"github.com/vennd/enu/internal/github.com/btcsuite/btcutil"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

// Globals
var config btcrpcclient.ConnConfig
var isInit bool = false // set to true only after the init sequence is complete

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
		os.Exit(-100)
	}

	err = json.Unmarshal(file, &configuration)

	if err != nil {
		log.Println("Unable to parse enuapi.json")
		os.Exit(-100)
	}

	m := configuration.(map[string]interface{})

	// Bitcoin API parameters
	config.Host = m["btchost"].(string)     // Hostname:port for Bitcoin Core or BTCD
	config.User = m["btcuser"].(string)     // Basic authentication user name
	config.Pass = m["btcpassword"].(string) // Basic authentication password
	config.HTTPPostMode = true              // Bitcoin core only supports HTTP POST mode
	config.DisableTLS = true                // Bitcoin core does not provide TLS by default

	isInit = true
}

// Thanks to https://raw.githubusercontent.com/btcsuite/btcrpcclient/master/examples/bitcoincorehttp/main.go
func GetBlockCount() (int64, error) {
	if isInit == false {
		Init()
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(&config, nil)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	defer client.Shutdown()

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}

	return blockCount, nil
}

func GetNewAddress() (string, error) {
	if isInit == false {
		Init()
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(&config, nil)
	if err != nil {
		return "", err
	}
	defer client.Shutdown()

	// Get a new BTC address.
	address, err := client.GetNewAddress("")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", address), nil
}

// Transmits to the bitcoin network the raw transaction as provided.
// The transaction should be encoded as a hex string, as per the original Bitcoin RPC JSON API.
// The TxId of the transaction is returned if successfully transmitted.
func SendRawTransaction(c context.Context, txHexString string) (string, error) {
	if isInit == false {
		Init()
	}

	// Copy same context values to local variables which are often accessed
	env := c.Value(consts.EnvKey).(string)

	if env == "dev" {
		return "success", nil
	}

	// Convert the hex string to a byte array
	txBytes, err := hex.DecodeString(txHexString)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	// Deserialise the transaction
	tx, err := btcutil.NewTxFromBytes(txBytes)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	msgTx := tx.MsgTx()

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(&config, nil)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	defer client.Shutdown()

	// Send the tx
	result, err := client.SendRawTransaction(msgTx, true)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	return fmt.Sprintf("%s", result.String()), nil
}

func GetBalance(c context.Context, address string) (uint64, error) {
	if isInit == false {
		Init()
	}

	result, status, err := httpGet(c, "http://btc.blockr.io/api/v1/address/balance/"+address)

	if status != 200 {
		log.FluentfContext(consts.LOGERROR, c, string(result))
	}

	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())

		return 0.0, err
	}

	var r interface{}

	if err := json.Unmarshal(result, &r); err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return 0.0, err
	}

	m := r.(map[string]interface{})

	if m["status"] == nil || m["status"] != "success" {
		return 0.0, errors.New("Blockr.io unavailable")
	}

	data := m["data"].(map[string]interface{})
	balanceFloat := data["balance"].(float64) * consts.Satoshi

	return uint64(balanceFloat), nil
}

func httpGet(c context.Context, url string) ([]byte, int64, error) {
	// Set headers
	req, err := http.NewRequest("GET", url, nil)

	clientPointer := &http.Client{}
	resp, err := clientPointer.Do(req)

	if err != nil {
		log.FluentfContext(consts.LOGDEBUG, c, "Request failed. %s", err.Error())
		return nil, 0, err
	}

	if resp.StatusCode != 200 {
		log.FluentfContext(consts.LOGDEBUG, c, "Request failed. Status code: %d\n", resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			return nil, 0, err
		}

		log.FluentfContext(consts.LOGDEBUG, c, "Reply: %s", string(body))

		return nil, -1000, errors.New(string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, 0, err
	}

	return body, 0, nil
}

func GetRawTransaction(txid string) (*btcjson.TxRawResult, error) {
	if isInit == false {
		Init()
	}

	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := btcrpcclient.New(&config, nil)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer client.Shutdown()

	txHash, err := wire.NewShaHashFromStr(txid)
	if err != nil {
		log.Fluentf(consts.LOGERROR, err.Error())
		return nil, err
	}

	txVerbose, err := client.GetRawTransactionVerbose(txHash)
	if err != nil {
		log.Fluentf(consts.LOGERROR, err.Error())
		return nil, err
	}

	return txVerbose, nil
}

func GetConfirmations(txid string) (uint64, error) {
	if isInit == false {
		Init()
	}

	// Testing
	if txid == "success" {
		return 777, nil
	}

	rawtx, err := GetRawTransaction(txid)

	if err != nil {
		return 0, err
	}

	return rawtx.Confirmations, nil
}
