package ripplehandlers

import (
	"encoding/json"
	"net/http"
	//	"strconv"
	"strings"
	"time"

	"errors"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/internal/github.com/vennd/mneumonic"
	"github.com/vennd/enu/log"
	"github.com/vennd/enu/rippleapi"
	"github.com/vennd/enu/ripplecrypto"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

func AssetCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var assetStruct enulib.Asset
	var distributionAddress string
	var distributionPassphrase string

	requestId := c.Value(consts.RequestIdKey).(string)
	assetStruct.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// The issuing address
	sourceAddress := m["sourceAddress"].(string)
	passphrase := m["passphrase"].(string)

	// The address which will hold the asset once it is issued
	if m["distributionAddress"] != nil {
		distributionAddress = m["distributionAddress"].(string)
	}
	if m["distributionPassphrase"] != nil {
		distributionPassphrase = m["distributionPassphrase"].(string)
	}

	asset := m["asset"].(string)
	quantity := uint64(m["quantity"].(float64))

	log.FluentfContext(consts.LOGINFO, c, "AssetCreate: received request Address: %s, asset: %s, quantity: %d, distributionAddress: %s from accessKey: %s\n", sourceAddress, asset, quantity, distributionAddress, c.Value(consts.AccessKeyKey).(string))

	// Generate an assetId
	assetId := enulib.GenerateAssetId()
	log.Printf("Generated assetId: %s", assetId)
	assetStruct.AssetId = assetId
	rippleAsset, err := rippleapi.ToCurrency(asset)
	if err != nil {
		log.FluentfContext(consts.LOGINFO, c, "Error in call to rippleapi.ToCurrency(): %s", err.Error())
	}

	//   If a distribution address has been specified, the passphrase must also be specified
	if distributionAddress != "" && distributionPassphrase == "" {
		log.FluentfContext(consts.LOGERROR, c, "If a distribution address is specified, the passphrase for the distribution address must be given.")
		handlers.ReturnBadRequest(c, w, consts.RippleErrors.DistributionPassphraseMissing.Code, consts.RippleErrors.DistributionPassphraseMissing.Description)

		return nil
	}

	//	If no distribution wallet was specified, create one to return to the client
	if distributionAddress == "" {
		//  create the wallet
		//  activate the wallet specifying a trust line for the asset from the issuing address
		wallet, errorCode, err := rippleapi.CreateWallet(c)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.CreateWallet: %s", err.Error())

			handlers.ReturnServerErrorWithCustomError(c, w, errorCode, err.Error())
			return nil
		}

		mn := mneumonic.FromHexstring(wallet.MasterSeedHex)
		passphrase := strings.Join(mn.ToWords(), " ") // The hex seed for Ripple wallets can be translated to the same mneumonic that generates counterparty wallets

		// Return to the client a distribution address if they needed one generated
		distributionAddress = wallet.AccountId
		distributionPassphrase = passphrase
		assetStruct.DistributionAddress = distributionAddress
		assetStruct.DistributionPassphrase = distributionPassphrase
	}

	assetStruct.BlockchainId = consts.RippleBlockchainId
	assetStruct.Asset = rippleAsset
	assetStruct.Issuer = sourceAddress
	assetStruct.Description = asset
	assetStruct.Quantity = quantity
	assetStruct.SourceAddress = sourceAddress

	// Return to the client the assetId and unblock the client
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(assetStruct); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	// Start asset creation in async mode
	go delegatedAssetCreate(c, sourceAddress, passphrase, distributionAddress, distributionPassphrase, asset, asset, quantity, assetId)

	return nil
}

// Concurrency safe to create and send transactions from a single address.
func delegatedAssetCreate(c context.Context, issuingAddress string, issuingPassphrase string, distributionAddress string, distributionPassphrase string, asset string, assetDescription string, quantity uint64, assetId string) (int64, error) {
	//	var complete bool = false
	//	var numLinesRequired = 0
	//	var retries int = 0

	// Copy same context values to local variables which are often accessed
	accessKey := c.Value(consts.AccessKeyKey).(string)
	blockchainId := c.Value(consts.BlockchainIdKey).(string)

	// Convert asset name to ripple currency encoding
	rippleAsset, err := rippleapi.ToCurrency(asset)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.ToCurrency(): %s", err.Error())

		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
		return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	// Write the asset with the generated asset id to the database
	go database.InsertAsset(accessKey, blockchainId, assetId, issuingAddress, distributionAddress, rippleAsset, assetDescription, quantity, true, "valid")

	// Set issuer up as a gateway https://ripple.com/build/gateway-guide/
	// set DefaultRipple on the issuer https://ripple.com/build/gateway-guide/#defaultripple
	//
	// First check if the defaultRipple flag is already set
	accountInfo, _, err := rippleapi.GetAccountInfo(c, issuingAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.GetAccountInfo(): %s", err.Error())

		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
		return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	// If defaultRipple isn't set, set it
	defaultRipple := accountInfo.Flags & rippleapi.LsfDefaultRipple
	//	log.FluentfContext(consts.LOGINFO, c, "flags: %d", defaultRipple)
	//	log.FluentfContext(consts.LOGINFO, c, "defaultripple flag: %d", rippleapi.LsfDefaultRipple)

	if defaultRipple != rippleapi.LsfDefaultRipple {
		log.FluentfContext(consts.LOGINFO, c, "defaultRipple is NOT set for account %s. Setting the flag...", issuingAddress)
		txHash, _, err := rippleapi.AccountSetFlag(c, issuingAddress, 8, ripplecrypto.PassphraseToSecret(c, issuingPassphrase))
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.AccountSetFlag(): %s", err.Error())

			database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
			return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
		}

		log.FluentfContext(consts.LOGINFO, c, "defaultRipple has now been set on %s. TxId: %s", issuingAddress, txHash)
	} else {
		log.FluentfContext(consts.LOGINFO, c, "defaultRipple is already set on issuingAddress: %s", issuingAddress)
	}

	//	Activate the distribution wallet if necessary
	assets := []rippleapi.Amount{
		{Currency: asset, Issuer: issuingAddress},
	}
	_, err = delegatedActivateAddress(c, distributionAddress, distributionPassphrase, 1, assets, assetId)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in delegatedActivateAddress(): %s", err.Error())

		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
		return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	log.FluentfContext(consts.LOGINFO, c, "Waiting until address activation complete...")
	time.Sleep(time.Duration(10) * time.Second)
	log.FluentfContext(consts.LOGINFO, c, "Done")

	// The distribution wallet should be set up now at this stage

	// Check if a trust line already exists between the issuing account and the distribution account
	lines, _, err := rippleapi.GetAccountLines(c, distributionAddress)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in rippleapi.GetAccountLines: %s", err.Error())

		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
		return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	// A trust line should exist by this stage.
	if lines.Contains(issuingAddress, rippleAsset) == false {
		log.FluentfContext(consts.LOGERROR, c, "Trust line from distribution %s to issuer %s does not exist for %s", distributionAddress, issuingAddress, asset+"->"+rippleAsset)

		return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	// Pay from the issuer wallet to the distribution wallet the amount of custom currency specified
	payTxId, _, err := delegatedSend(c, accessKey, issuingPassphrase, issuingAddress, distributionAddress, asset, issuingAddress, quantity, assetId, "Asset creation")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in delegatedSend: %s", err.Error())

		database.UpdateAssetWithErrorByAssetId(c, accessKey, assetId, consts.RippleErrors.MiscError.Code, consts.RippleErrors.MiscError.Description)
		return consts.RippleErrors.MiscError.Code, errors.New(consts.RippleErrors.MiscError.Description)
	}

	database.UpdateAssetCompleteByAssetId(c, accessKey, assetId, payTxId)
	return 0, nil
}
