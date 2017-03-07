// database.go
package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/log"

	_ "github.com/vennd/enu/internal/github.com/go-sql-driver/mysql"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

var Db *sql.DB
var databaseString string
var isInit bool = false // set to true only after the init sequence is complete

// Initialises global variables and database connection for all handlers
func Init() {
	var configFilePath string

	if isInit == true {
		return
	}

	if _, err := os.Stat("./enuapi.json"); err == nil {
		log.Println("Found and using configuration file ./enuapi.json")
		configFilePath = "./enuapi.json"
	} else {
		if _, err := os.Stat(os.Getenv("GOPATH") + "/bin/enuapi.json"); err == nil {
			configFilePath = os.Getenv("GOPATH") + "/bin/enuapi.json"
			log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

		} else {
			if _, err := os.Stat(os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"); err == nil {
				configFilePath = os.Getenv("GOPATH") + "/src/github.com/vennd/enu/enuapi.json"
				log.Printf("Found and using configuration file from GOPATH: %s\n", configFilePath)

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
	log.Printf("Reading %s\n", configFilePath)
	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Println("Unable to read configuration file enuapi.json")
		log.Println(err.Error())
		os.Exit(-100)
	}

	err = json.Unmarshal(file, &configuration)

	m := configuration.(map[string]interface{})

	//	localMode := m["localMode"].(string) // True if running on local Windows development machine
	dbUrl := m["dburl"].(string)         // URL for MySQL
	schema := m["schema"].(string)       // Database schema name
	user := m["dbuser"].(string)         // User name for the DB
	password := m["dbpassword"].(string) // Password for the specified database

	stringsToConcatenate := []string{user, ":", password, "@", dbUrl, "/", schema}
	databaseString = strings.Join(stringsToConcatenate, "")

	log.Printf("Opening: %s\n", strings.Join([]string{dbUrl, "/", schema}, ""))
	Db, err = sql.Open("mysql", databaseString)
	if err != nil {
		panic(err.Error())
	}

	// Ping to check DB connection is okay
	err = Db.Ping()
	if err != nil {
		panic(err.Error())
	}

	log.Println("Opened DB successfully!")

	isInit = true
}

// Inserts an asset into the assets database
func InsertAsset(accessKey string, blockchainId string, assetId string, sourceAddressValue string, distributionAddressValue string, assetValue string, descriptionValue string, quantityValue uint64, divisibleValue bool, status string) error {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("insert into assets(accessKey, blockchainId, assetId, sourceAddress, distributionAddress, asset, description, quantity, divisible, status) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Perform the insert
	_, err = stmt.Exec(accessKey, blockchainId, assetId, sourceAddressValue, distributionAddressValue, assetValue, descriptionValue, quantityValue, divisibleValue, status)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return nil
}

func GetAssetByAssetId(c context.Context, accessKey string, assetId string) (enulib.Asset, error) {
	if isInit == false {
		Init()
	}

	// Set some initial values
	var assetStruct = enulib.Asset{}
	assetStruct.AssetId = assetId
	assetStruct.Status = consts.NotFound

	//	 Query DB
	log.FluentfContext(consts.LOGINFO, c, "select rowId, assetId, blockchainId, sourceAddress, distributionAddress, asset, description, quantity, divisible, status, errorDescription, broadcastTxId from assets where assetId=%s and accessKey=%s", assetId, accessKey)
	stmt, err := Db.Prepare("select rowId, assetId, blockchainId, sourceAddress, distributionAddress, asset, description, quantity, divisible, status, errorDescription, broadcastTxId from assets where assetId=? and accessKey=?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return assetStruct, err
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(assetId, accessKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to QueryRow. Reason: %s", err.Error())
		return assetStruct, err
	}

	var rowId string
	var blockchainId []byte
	var sourceAddress []byte
	var distributionAddress []byte
	var asset []byte
	var description []byte
	var quantity uint64
	var divisible bool
	var status []byte
	var errorMessage []byte
	var broadcastTxId []byte

	if err := row.Scan(&rowId, &assetId, &blockchainId, &sourceAddress, &distributionAddress, &asset, &description, &quantity, &divisible, &status, &errorMessage, &broadcastTxId); err == sql.ErrNoRows {
		if err.Error() == "sql: no rows in result set" {
		}
	} else if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		return assetStruct, err
	} else {
		assetStruct = enulib.Asset{BlockchainId: string(blockchainId), SourceAddress: string(sourceAddress), DistributionAddress: string(distributionAddress), Asset: string(asset), Description: string(description), Quantity: quantity, AssetId: assetId, Status: string(status), ErrorMessage: string(errorMessage)}
	}

	return assetStruct, nil
}

func UpdateAssetWithErrorByAssetId(c context.Context, accessKey string, assetId string, errorCode int64, errorDescription string) error {
	if isInit == false {
		Init()
	}

	asset, err := GetAssetByAssetId(c, accessKey, assetId)
	if err != nil {
		return err
	}

	if asset.AssetId == "" {
		errorString := fmt.Sprintf("Asset does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	//	log.FluentfContext(consts.LOGINFO, c, "update assets set status='error', errorCode=%d, errorDescription=%s where accessKey=%s and assetId = %s", errorCode, errorDescription, accessKey, assetId)
	stmt, err := Db.Prepare("update assets set status='error', errorCode=?, errorDescription=? where accessKey=? and assetId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(errorCode, errorDescription, accessKey, assetId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdateAssetStatusByAssetId(c context.Context, accessKey string, assetId string, status string) error {
	if isInit == false {
		Init()
	}

	asset, err := GetAssetByAssetId(c, accessKey, assetId)
	if err != nil {
		return err
	}

	if asset.AssetId == "" {
		errorString := fmt.Sprintf("Asset does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update assets set status=? where accessKey=? and assetId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(status, accessKey, assetId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdateAssetNameByAssetId(c context.Context, accessKey string, assetId string, assetName string) error {
	if isInit == false {
		Init()
	}

	asset, err := GetAssetByAssetId(c, accessKey, assetId)
	if err != nil {
		return err
	}

	if asset.AssetId == "" {
		errorString := fmt.Sprintf("Asset does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update assets set asset=? where accessKey=? and assetId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(assetName, accessKey, assetId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdateAssetCompleteByAssetId(c context.Context, accessKey string, assetId string, txId string) error {
	if isInit == false {
		Init()
	}

	asset, err := GetAssetByAssetId(c, accessKey, assetId)
	if err != nil {
		return err
	}

	if asset.AssetId == "" {
		errorString := fmt.Sprintf("Asset does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	log.Printf("update assets set status='complete', broadcastTxId=%s where accessKey=%s and assetId = %s\n", txId, accessKey, assetId)

	stmt, err := Db.Prepare("update assets set status='complete', broadcastTxId=? where accessKey=? and assetId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(txId, accessKey, assetId)
	if err2 != nil {
		return err2
	}

	return nil
}

// Inserts a dividend into the dividends database
func InsertDividend(accessKey string, dividendId string, sourceAddressValue string, assetValue string, dividendAssetValue string, quantityPerUnitValue uint64, status string) {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("insert into dividends(accessKey, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, status) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	// Perform the insert
	_, err = stmt.Exec(accessKey, dividendId, sourceAddressValue, assetValue, dividendAssetValue, quantityPerUnitValue, status)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
}

func GetDividendByDividendId(c context.Context, accessKey string, dividendId string) (enulib.Dividend, error) {
	if isInit == false {
		Init()
	}

	// Initialise some initial values
	var dividendStruct = enulib.Dividend{}
	dividendStruct.DividendId = dividendId
	dividendStruct.Status = consts.NotFound

	//	 Query DB
	//	log.FluentfContext(consts.LOGDEBUG, c, "select rowId, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, errorDescription, broadcastTxId from dividends where dividendId=%s and accessKey=%s", dividendId, accessKey)
	stmt, err := Db.Prepare("select rowId, dividendId, sourceAddress, asset, dividendAsset, quantityPerUnit, status, errorDescription, broadcastTxId from dividends where dividendId=? and accessKey=?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return dividendStruct, err
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(dividendId, accessKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to QueryRow. Reason: %s", err.Error())
		return dividendStruct, err
	}

	var rowId string
	var sourceAddress []byte
	var asset []byte
	var dividendAsset []byte
	var quantityPerUnit uint64
	var status []byte
	var errorMessage []byte
	var broadcastTxId []byte

	if err := row.Scan(&rowId, &dividendId, &sourceAddress, &asset, &dividendAsset, &quantityPerUnit, &status, &errorMessage, &broadcastTxId); err == sql.ErrNoRows {
		if err.Error() == consts.SqlNotFound {
			dividendStruct.Status = consts.NotFound
			return dividendStruct, err
		}
	} else if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
	} else {
		dividendStruct = enulib.Dividend{SourceAddress: string(sourceAddress), Asset: string(asset), DividendAsset: string(dividendAsset), QuantityPerUnit: quantityPerUnit, DividendId: dividendId, Status: string(status), ErrorMessage: string(errorMessage), BroadcastTxId: string(broadcastTxId)}
	}

	return dividendStruct, nil
}

func UpdateDividendWithErrorByDividendId(c context.Context, accessKey string, dividendId string, errorCode int64, errorDescription string) error {
	if isInit == false {
		Init()
	}

	dividend, err := GetDividendByDividendId(c, accessKey, dividendId)

	if dividend.Status == consts.NotFound || err != nil {
		return errors.New(consts.GenericErrors.NotFound.Description)
	}

	stmt, err := Db.Prepare("update dividends set status='error', errorCode=?, errorDescription=? where accessKey=? and dividendId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(errorCode, errorDescription, accessKey, dividendId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdateDividendCompleteByDividendId(c context.Context, accessKey string, dividendId string, txId string) error {
	if isInit == false {
		Init()
	}

	dividend, err := GetDividendByDividendId(c, accessKey, dividendId)

	if dividend.DividendId == "" || err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		errorString := fmt.Sprintf("Dividend does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	log.FluentfContext(consts.LOGINFO, c, "update dividends set status='complete', broadcastTxId=%s where accessKey=%s and dividendId = %s\n", txId, accessKey, dividendId)

	stmt, err := Db.Prepare("update dividends set status='complete', broadcastTxId=? where accessKey=? and dividendId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(txId, accessKey, dividendId)
	if err2 != nil {
		return err2
	}

	return nil
}

// Inserts a payment into the payment database
func InsertPayment(c context.Context, accessKey string, blockIdValue int64, blockchainIdValue string, sourceTxidValue string, sourceAddressValue string, destinationAddressValue string, outAssetValue string, issuerValue string, outAmountValue uint64, statusValue string, lastUpdatedBlockIdValue int64, txFeeValue uint64, paymentTag string) {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("insert into payments(accessKey, blockId, blockchainId, sourceTxid, sourceAddress, destinationAddress, outAsset, issuer, outAmount, status, lastUpdatedBlockId, txFee, broadcastTxId, paymentTag) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', ?)")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	// Perform the insert
	_, err = stmt.Exec(accessKey, blockIdValue, blockchainIdValue, sourceTxidValue, sourceAddressValue, destinationAddressValue, outAssetValue, issuerValue, outAmountValue, statusValue, lastUpdatedBlockIdValue, txFeeValue, paymentTag)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
}

func GetPaymentByPaymentId(c context.Context, accessKey string, paymentId string) enulib.SimplePayment {
	if isInit == false {
		Init()
	}

	//	 Query DB
	stmt, err := Db.Prepare("select rowId, blockId, blockchainId, sourceTxId, sourceAddress, destinationAddress, outAsset, issuer, outAmount, status, lastUpdatedBlockId, txFee, broadcastTxId, paymentTag, errorDescription from payments where sourceTxid=? and accessKey=?")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(paymentId, accessKey)
	if err != nil {
		panic(err.Error())
	}

	var rowId string
	var blockId []byte
	var blockchainId []byte
	var sourceAddress []byte
	var destinationAddress []byte
	var asset []byte
	var issuer []byte
	var amount uint64
	var txFee int64
	var broadcastTxId []byte
	var status []byte
	var sourceTxId []byte
	var lastUpdatedBlockId []byte
	var payment enulib.SimplePayment
	var paymentTag []byte
	var errorMessage []byte

	if err := row.Scan(&rowId, &blockId, &blockchainId, &sourceTxId, &sourceAddress, &destinationAddress, &asset, &issuer, &amount, &status, &lastUpdatedBlockId, &txFee, &broadcastTxId, &paymentTag, &errorMessage); err == sql.ErrNoRows {
		payment = enulib.SimplePayment{}
		if err.Error() == "sql: no rows in result set" {
			payment.PaymentId = paymentId
			payment.Status = consts.NotFound
		}
	} else if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
	}

	payment = enulib.SimplePayment{BlockchainId: string(blockchainId), SourceAddress: string(sourceAddress), DestinationAddress: string(destinationAddress), Asset: string(asset), Amount: amount, PaymentId: string(sourceTxId), Status: string(status), BroadcastTxId: string(broadcastTxId), TxFee: txFee, ErrorMessage: string(errorMessage)}

	return payment
}

func GetPaymentByPaymentTag(c context.Context, accessKey string, paymentTag string) enulib.SimplePayment {
	if isInit == false {
		Init()
	}

	//	 Query DB
	stmt, err := Db.Prepare("select rowId, blockId, blockchainId, sourceTxId, sourceAddress, destinationAddress, outAsset, issuer, outAmount, status, lastUpdatedBlockId, txFee, broadcastTxId, paymentTag, errorDescription from payments where paymentTag=? and accessKey=?")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(paymentTag, accessKey)
	if err != nil {
		panic(err.Error())
	}

	var rowId string
	var blockId []byte
	var blockchainId []byte
	var sourceAddress []byte
	var destinationAddress []byte
	var asset []byte
	var issuer []byte
	var amount uint64
	var txFee int64
	var broadcastTxId []byte
	var status []byte
	var sourceTxId []byte
	var lastUpdatedBlockId uint64
	var payment enulib.SimplePayment
	var errorMessage []byte

	if err := row.Scan(&rowId, &blockId, &blockchainId, &sourceTxId, &sourceAddress, &destinationAddress, &asset, &issuer, &amount, &status, &lastUpdatedBlockId, &txFee, &broadcastTxId, &errorMessage); err == sql.ErrNoRows {
		payment = enulib.SimplePayment{}
		if err.Error() == "sql: no rows in result set" {
			payment.PaymentTag = paymentTag
			payment.Status = consts.NotFound
		}
	} else {
		log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
	}

	payment = enulib.SimplePayment{BlockchainId: string(blockchainId), SourceAddress: string(sourceAddress), DestinationAddress: string(destinationAddress), Asset: string(asset), Issuer: string(issuer), Amount: amount, PaymentId: string(sourceTxId), Status: string(status), BroadcastTxId: string(broadcastTxId), TxFee: txFee, ErrorMessage: string(errorMessage), PaymentTag: string(paymentTag)}

	return payment
}

func GetPaymentsByAddress(c context.Context, accessKey string, address string) []enulib.SimplePayment {
	var result []enulib.SimplePayment

	if isInit == false {
		Init()
	}

	//	 Query DB
	//	log.Fluentf(consts.LOGDEBUG, "select rowId, blockId, blockchainId, sourceTxId, sourceAddress, destinationAddress, outAsset, issuer, outAmount, status, lastUpdatedBlockId, txFee, broadcastTxId, paymentTag, errorDescription from payments where accessKey = %s and (sourceAddress = %s or destinationAddress = %s)", accessKey, address, address)
	stmt, err := Db.Prepare("select rowId, blockId, blockchainId, sourceTxId, sourceAddress, destinationAddress, outAsset, outAmount, issuer, status, lastUpdatedBlockId, txFee, broadcastTxId, paymentTag, errorDescription from payments where accessKey = ? and (sourceAddress = ? or destinationAddress = ?)")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to prepare statement. Reason: %s", err.Error())
		return result
	}
	defer stmt.Close()

	//	 Get row
	rows, err := stmt.Query(accessKey, address, address)
	defer rows.Close()
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Failed to query. Reason: %s", err.Error())
		return result
	}

	for rows.Next() {
		var rowId string
		var blockId []byte
		var blockchainId []byte
		var sourceAddress []byte
		var destinationAddress []byte
		var asset []byte
		var issuer []byte
		var amount uint64
		var txFee int64
		var broadcastTxId []byte
		var status []byte
		var sourceTxId []byte
		var lastUpdatedBlockId uint64
		var payment enulib.SimplePayment
		var errorMessage []byte
		var paymentTag []byte

		if err := rows.Scan(&rowId, &blockId, &blockchainId, &sourceTxId, &sourceAddress, &destinationAddress, &asset, &amount, &issuer, &status, &lastUpdatedBlockId, &txFee, &broadcastTxId, &paymentTag, &errorMessage); err == sql.ErrNoRows {
			payment = enulib.SimplePayment{}
			if err.Error() == "sql: no rows in result set" {
				payment.Status = consts.NotFound
			}
		} else if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Failed to Scan. Reason: %s", err.Error())
		}

		payment = enulib.SimplePayment{BlockchainId: string(blockchainId), SourceAddress: string(sourceAddress), DestinationAddress: string(destinationAddress), Asset: string(asset), Issuer: string(issuer), Amount: amount, PaymentId: string(sourceTxId), Status: string(status), BroadcastTxId: string(broadcastTxId), TxFee: txFee, ErrorMessage: string(errorMessage), PaymentTag: string(paymentTag)}

		result = append(result, payment)
	}

	return result
}

func UpdatePaymentStatusByPaymentId(c context.Context, accessKey string, paymentId string, status string) error {
	if isInit == false {
		Init()
	}

	payment := GetPaymentByPaymentId(c, accessKey, paymentId)

	if payment.PaymentId == "" {
		errorString := fmt.Sprintf("Payment does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update payments set status=? where accessKey=? and sourceTxId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(status, accessKey, paymentId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdatePaymentWithErrorByPaymentId(c context.Context, accessKey string, paymentId string, errorCode int64, errorDescription string) error {
	if isInit == false {
		Init()
	}

	payment := GetPaymentByPaymentId(c, accessKey, paymentId)

	if payment.PaymentId == "" {
		errorString := fmt.Sprintf("Payment does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update payments set status='error', errorCode=?, errorDescription=? where accessKey=? and sourceTxId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(errorCode, errorDescription, accessKey, paymentId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdatePaymentCompleteByPaymentId(c context.Context, accessKey string, paymentId string, txId string) error {
	if isInit == false {
		Init()
	}

	payment := GetPaymentByPaymentId(c, accessKey, paymentId)

	if payment.PaymentId == "" {
		errorString := fmt.Sprintf("Payment does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update payments set status='complete', broadcastTxId=? where accessKey=? and sourceTxId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(txId, accessKey, paymentId)
	if err2 != nil {
		return err2
	}

	return nil
}

func UpdatePaymentSignedRawTxByPaymentId(c context.Context, accessKey string, paymentId string, signedRawTx string) error {
	if isInit == false {
		Init()
	}

	payment := GetPaymentByPaymentId(c, accessKey, paymentId)

	if payment.PaymentId == "" {
		errorString := fmt.Sprintf("Payment does not exist or cannot be accessed by %s\n", accessKey)

		return errors.New(errorString)
	}

	stmt, err := Db.Prepare("update payments set signedRawTx=? where accessKey=? and sourceTxId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(signedRawTx, accessKey, paymentId)
	if err2 != nil {
		return err2
	}

	return nil
}

// create table userKeys (userId BIGINT, accessKey varchar(64), secret varchar(64), nonce bigint, assetId varchar(100), blockchainId varchar(100), sourceAddress varchar(100))
// Used to verify if the current request has a nonce > the value stored in the DB
func GetNonceByAccessKey(accessKey string) int64 {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select nonce from userkeys where accessKey=?")

	if err != nil {
		return -1
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var nonce int64
	row.Scan(&nonce)

	return nonce
}

// Used to retrieve the secret to verify the HMAC signature
func GetSecretByAccessKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select secret from userkeys where accessKey=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var secret string
	row.Scan(&secret)

	return secret
}

// Returns newest address associated with the access key
func GetSourceAddressByAccessKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select a.sourceAddress as sourceAddress from userKeys u left outer join addresses a on u.accessKey = a.accessKey where a.accessKey=? order by a.rowId desc limit 1;")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var sourceAddress string
	row.Scan(&sourceAddress)

	return sourceAddress
}

func GetAssetByAccessKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select assetId from userKeys where accessKey=? and status=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey, consts.AccessKeyValidStatus)

	var assetId string
	row.Scan(&assetId)

	return assetId
}

// Used to update the value of the nonce after a successful API call
func UpdateNonce(accessKey string, nonce int64) error {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("update userkeys set nonce=? where accessKey=? and status=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(nonce, accessKey, consts.AccessKeyValidStatus)
	if err2 != nil {
		return err2
	}

	return nil

}

func CreateUserKey(userId int64, assetId string, blockchainId string, sourceAddress string, parentAccessKey string) (string, string, error) {
	if isInit == false {
		Init()
	}

	// blockchainId must be in the list of blockchains that we support
	supportedBlockchains := consts.SupportedBlockchains
	sort.Strings(supportedBlockchains)

	i := sort.SearchStrings(supportedBlockchains, blockchainId)
	blockchainValid := i < len(supportedBlockchains) && supportedBlockchains[i] == blockchainId

	if blockchainValid == false {
		e := fmt.Sprintf("Unsupported blockchain. Valid values: %s", strings.Join(supportedBlockchains, ", "))

		return "", "", errors.New(e)
	}

	key := enulib.GenerateKey()
	secret := enulib.GenerateKey()

	// Open a transaction to ensure consistency between userKeys and addresses table
	tx, beginErr := Db.Begin()
	if beginErr != nil {
		return "", "", beginErr
	}

	// Insert into userKeys table first
	stmt, err := Db.Prepare("insert into userkeys(userId, parentAccessKey, accessKey, secret, nonce, assetId, blockchainId, status) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		tx.Rollback()
		return "", "", err
	}
	_, err = stmt.Exec(userId, parentAccessKey, key, secret, 0, assetId, blockchainId, consts.AccessKeyValidStatus)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	// Insert into addresses second
	stmt2, err2 := Db.Prepare("insert into addresses(accessKey, sourceAddress) values(?, ?)")
	if err2 != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		tx.Rollback()
		return "", "", err
	}
	_, err2 = stmt2.Exec(key, sourceAddress)
	if err != nil {
		tx.Rollback()
		return "", "", err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return "", "", commitErr
	}

	defer stmt.Close()

	return key, secret, nil
}

func CreateSecondaryAddress(c context.Context, accessKey string, newAddress string) error {
	if isInit == false {
		Init()
	}

	// Check accessKey exists
	if UserKeyExists(accessKey) != true {
		log.Println("Call to CreateSecondaryAddress() with an invalid access key")

		return errors.New("Call to CreateSecondaryAddress() with an invalid access key")
	}

	stmt, err := Db.Prepare("insert into addresses(accessKey, sourceAddress) values(?, ?)")
	if err != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		return err
	}

	// Perform the insert
	_, err = stmt.Exec(accessKey, newAddress)
	if err != nil {
		return err
	}

	defer stmt.Close()

	return nil
}

// Only return true where an accessKey exists and also has a valid status
func UserKeyExists(accessKey string) bool {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select count(*) from userkeys where accesskey=? and status=?")

	if err != nil {
		return false
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey, consts.AccessKeyValidStatus)

	var count int64
	row.Scan(&count)

	if count == 0 {
		return false
	}

	return true
}

// Updates a given accessKey, ignores what the existing status is
func UpdateUserKeyStatus(accessKey string, status string) error {
	if isInit == false {
		Init()
	}

	// status must be a supported access key status
	statuses := consts.AccessKeyStatuses
	sort.Strings(statuses)

	x := sort.SearchStrings(statuses, status)
	statusValid := x < len(statuses) && statuses[x] == status
	if statusValid == false {
		e := fmt.Sprintf("Attempt to update status to an invalid value: %s. Valid values: %s", status, strings.Join(statuses, ", "))

		return errors.New(e)
	}

	stmt, err := Db.Prepare("update userkeys set status=? where accessKey=?")
	if err != nil {
		//		log.Println("Failed to prepare statement. Reason: ")
		return err
	}

	// Perform the update
	_, err = stmt.Exec(status, accessKey)
	if err != nil {
		return err
	}

	defer stmt.Close()

	return nil
}

func GetStatusByUserKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select status from userkeys where accesskey=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey)

	var status string
	row.Scan(&status)

	return status
}

func GetBlockchainIdByUserKey(accessKey string) string {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("select blockchainId from userkeys where accesskey=? and status=?")

	if err != nil {
		return ""
	}
	defer stmt.Close()

	row := stmt.QueryRow(accessKey, consts.AccessKeyValidStatus)

	var blockchainId string
	row.Scan(&blockchainId)

	return blockchainId
}

// Inserts an activation request into the database
func InsertActivation(c context.Context, accessKey string, activationId string, blockchainId string, addressToActivate string, amount uint64) {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("insert into activations(activationId, blockchainId, accessKey, addressToActivate, amount) values(?, ?, ?, ?, ?)")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return
	}
	defer stmt.Close()

	// Perform the insert
	_, err = stmt.Exec(activationId, blockchainId, accessKey, addressToActivate, amount)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return
	}
	defer stmt.Close()
}

func GetActivationByActivationId(c context.Context, accessKey string, activationId string) map[string]interface{} {
	if isInit == false {
		Init()
	}

	requestId := c.Value(consts.RequestIdKey).(string)

	//	 Query DB
	//	log.FluentfContext(consts.LOGDEBUG, c, "select blockchainId, addressToActivate, amount, a.rowId, sourceAddress, outAsset, outAmount, status, broadcastTxId, errorDescription from activations a, payments p where a.activationId = p.sourceTxid and activationId=%s and a.accessKey=%s", activationId, accessKey)
	stmt, err := Db.Prepare("select a.blockchainId, addressToActivate, amount, a.rowId, sourceAddress, outAsset, outAmount, status, broadcastTxId, errorDescription from activations a, payments p where a.activationId = p.sourceTxid and activationId=? and a.accessKey=?")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return map[string]interface{}{}
	}
	defer stmt.Close()

	//	 Get row
	row := stmt.QueryRow(activationId, accessKey)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return map[string]interface{}{}
	}

	var blockchainId []byte
	var addressToActivate []byte
	var amount uint64
	var rowId string
	var sourceAddress []byte
	var outAsset []byte
	var outAmount int64
	var status []byte
	var broadcastTxId []byte
	var errorMessage []byte

	if err := row.Scan(&blockchainId, &addressToActivate, &amount, &rowId, &sourceAddress, &outAsset, &outAmount, &status, &broadcastTxId, &errorMessage); err == sql.ErrNoRows {
		if err.Error() == "sql: no rows in result set" {
			var result = map[string]interface{}{
				"activationId": activationId,
				"status":       consts.NotFound,
			}

			return result
		}
	} else if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
	}

	// Return the values
	var result = map[string]interface{}{
		"address":       string(addressToActivate),
		"amount":        amount,
		"activationId":  string(activationId),
		"broadcastTxId": string(broadcastTxId),
		"status":        string(status),
		"errorMessage":  string(errorMessage),
		"requestId":     string(requestId),
	}

	return result
}

// Inserts an activation request into the database
func InsertTrustAsset(c context.Context, accessKey string, activationId string, blockchainId string, asset string, issuer string, amount uint64) {
	if isInit == false {
		Init()
	}

	stmt, err := Db.Prepare("insert into trustassets(activationId, blockchainId, accessKey, asset, issuer, trustAmount) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return
	}
	defer stmt.Close()

	// Perform the insert
	_, err = stmt.Exec(activationId, blockchainId, accessKey, asset, issuer, amount)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		return
	}
	defer stmt.Close()
}
