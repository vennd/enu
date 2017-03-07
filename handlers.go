package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"math/rand"
	"net/http"
	"time"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/github.com/xeipuuv/gojsonschema"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
)

func ReturnUnauthorised(c context.Context, w http.ResponseWriter, errorCode int64, e error) {
	var returnCode enulib.ReturnCode

	if e == nil {
		returnCode = enulib.ReturnCode{Code: errorCode, Description: "Forbidden", RequestId: c.Value(consts.RequestIdKey).(string)}
	} else {
		returnCode = enulib.ReturnCode{Code: errorCode, Description: e.Error(), RequestId: c.Value(consts.RequestIdKey).(string)}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusForbidden)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnBadRequest(c context.Context, w http.ResponseWriter, errorCode int64, e error) {
	var returnCode enulib.ReturnCode

	if e == nil {
		returnCode = enulib.ReturnCode{Code: errorCode, Description: "Bad Request", RequestId: c.Value(consts.RequestIdKey).(string)}
	} else {
		returnCode = enulib.ReturnCode{Code: errorCode, Description: e.Error(), RequestId: c.Value(consts.RequestIdKey).(string)}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnUnprocessableEntity(c context.Context, w http.ResponseWriter, errorCode int64, e error) {
	var returnCode enulib.ReturnCode

	if e == nil {
		returnCode = enulib.ReturnCode{Code: errorCode, Description: "Unprocessable entity", RequestId: c.Value(consts.RequestIdKey).(string)}
	} else {
		returnCode = enulib.ReturnCode{Code: errorCode, Description: e.Error(), RequestId: c.Value(consts.RequestIdKey).(string)}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(422)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnCreated(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: 0, Description: "Success", RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnOK(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: 0, Description: "Success", RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnNotFound(c context.Context, w http.ResponseWriter, errorCode int64, e error) {
	returnCode := enulib.ReturnCode{Code: -3, Description: "Not found", RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnNotFoundWithCustomError(c context.Context, w http.ResponseWriter, errorCode int64, errorString string) {
	returnCode := enulib.ReturnCode{Code: -3, Description: errorString, RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

func ReturnServerError(c context.Context, w http.ResponseWriter, errorCode int64, e error) {
	var returnCode enulib.ReturnCode

	if e == nil {
		log.FluentfContext(consts.LOGERROR, c, "Unspecified server error.\n")
		returnCode = enulib.ReturnCode{Code: -10000, Description: "Unspecified server error. Please contact Vennd.io support.", RequestId: c.Value(consts.RequestIdKey).(string)}
	} else {
		log.FluentfContext(consts.LOGERROR, c, "Server error: %s\n", e.Error())
		returnCode = enulib.ReturnCode{Code: -10000, Description: e.Error(), RequestId: c.Value(consts.RequestIdKey).(string)}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		panic(err)
	}
}

// Handles the '/' path and returns a random quote
func Index(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	number := rand.Intn(len(quotes))

	fmt.Fprintf(w, "%s\n", quotes[number])
}

func CheckHeaderGeneric(c context.Context, w http.ResponseWriter, r *http.Request) (string, error) {
	// Pull headers that are necessary
	accessKey := r.Header.Get("AccessKey")
	signature := r.Header.Get("Signature")
	var err error

	// Headers weren't set properly, return forbidden
	if accessKey == "" || signature == "" {
		log.FluentfContext(consts.LOGERROR, c, "Headers set incorrectly: accessKey=%s, signature=%s\n", accessKey, signature)
		ReturnUnauthorised(c, w, consts.GenericErrors.HeadersIncorrect.Code, errors.New(consts.GenericErrors.HeadersIncorrect.Description))

		return accessKey, err
	} else if database.UserKeyExists(accessKey) == false {
		// User key doesn't exist
		log.FluentfContext(consts.LOGERROR, c, "Attempt to access API with unknown user key: %s", accessKey)
		ReturnUnauthorised(c, w, consts.GenericErrors.UnknownAccessKey.Code, errors.New(consts.GenericErrors.UnknownAccessKey.Description))

		return accessKey, errors.New(consts.GenericErrors.UnknownAccessKey.Description)
	}

	return accessKey, nil
}

func CheckAndParseJsonCTX(c context.Context, w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	//	var blockchainId string
	var payload interface{}
	var nonceDB int64

	signature := r.Header.Get("Signature")

	// Limit amount read to 512,000 bytes and parse body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 512000))
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil, err
	}
	if err := r.Body.Close(); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Body.Close(): %s", err.Error())
		ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))

		return nil, err
	}

	// If the body is an empty byte array then don't attempt to unmarshall the JSON and set a default
	if bytes.Compare(body, make([]byte, 0)) != 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			log.FluentfContext(consts.LOGINFO, c, "Malformed body: %s", string(body))
			returnErr := errors.New("The request did not contain a valid JSON object")
			log.FluentfContext(consts.LOGERROR, c, err.Error())                                   // Log the real error
			ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidDocument.Code, returnErr) // Send back sanitised error

			return nil, returnErr
		}
		log.FluentfContext(consts.LOGINFO, c, "Request received: %s", body)
	} else if bytes.Compare(body, make([]byte, 0)) == 0 {
		log.FluentfContext(consts.LOGINFO, c, "Empty body received")
		returnErr := errors.New("Empty body received. If you aren't sending a payload then send an empty JSON object: {}")
		log.FluentfContext(consts.LOGERROR, c, returnErr.Error())
		ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidDocument.Code, returnErr)

		return nil, returnErr
	}

	// Then look up secret and calculate digest
	accessKey := c.Value(consts.AccessKeyKey).(string)
	calculatedSignature := enulib.ComputeHmac512(body, database.GetSecretByAccessKey(accessKey))

	// If we didn't receive the expected signature then raise a forbidden
	if calculatedSignature != signature {
		errorString := fmt.Sprintf("Could not verify HMAC signature. Expected: %s, received: %s", calculatedSignature, signature)
		err := errors.New(errorString)
		ReturnUnauthorised(c, w, consts.GenericErrors.InvalidSignature.Code, err)

		return nil, err
	}

	m := payload.(map[string]interface{})

	// nonce checking
	var nonceInt int64
	if m["nonce"] != nil {
		nonceInt := int64(m["nonce"].(float64))
		log.FluentfContext(consts.LOGINFO, c, "Nonce received: %s", nonceInt)
	} else {
		nonceInt = 0
	}

	if nonceInt > 0 {
		nonceDB = database.GetNonceByAccessKey(accessKey)
		if nonceInt <= nonceDB {
			//Nonce is not greater than the nonce in the DB
			log.FluentfContext(consts.LOGERROR, c, "Nonce for accessKey %s provided is <= nonce in db. %d <= %d\n", accessKey, nonceInt, nonceDB)
			ReturnUnauthorised(c, w, consts.GenericErrors.InvalidNonce.Code, errors.New(consts.GenericErrors.InvalidNonce.Description))

			return nil, err
		} else {
			log.FluentfContext(consts.LOGINFO, c, "Nonce for accessKey %s provided is ok. (%s > %d)\n", accessKey, nonceInt, nonceDB)
			database.UpdateNonce(accessKey, nonceInt)
			if err != nil {
				log.FluentfContext(consts.LOGERROR, c, "Nonce update failed, error: %s", err.Error())
				ReturnServerError(c, w, consts.GenericErrors.GeneralError.Code, errors.New("Nonce handling failed"))

				return nil, err
			}
		}
	}

	// Arg checking

	u, ok := c.Value(consts.RequestTypeKey).(string)
	if ok {

		check := make(map[string]string)
		check["asset"] =
			`
		{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"description":{"type":"string"},"asset":{"type":"string","minLength":4},"quantity":{"type":"integer"},"divisible":{"type":"boolean"}},"required":["sourceAddress","asset","quantity","divisible"]}
	`
		check["dividend"] =
			`
		{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"asset":{"type":"string","minLength":4},"dividendAsset":{"type":"string"},"quantityPerUnit":{"type":"integer"}},"required":["sourceAddress","asset","dividendAsset","quantityPerUnit"]}
	`
		check["walletCreate"] =
			`
		{"properties":{"numberOfAddresses":{"type":"integer"}}}
	`
		check["walletPayment"] =
			`
		{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"destinationAddress":{"type":"string", "maxLength":34, "minLength":34},"asset":{"type":"string","minLength":4},"quantity":{"type":"integer"}},"required":["sourceAddress","asset","quantity","destinationAddress"]}
	`
		check["simplePayment"] =
			`
		{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"destinationAddress":{"type":"string", "maxLength":34, "minLength":34},"asset":{"type":"string","minLength":4},"amount":{"type":"integer"},,"txFee":{"type":"integer"}},"required":["sourceAddress","destinationAddress","asset","amount"]}
	`
		check["activateaddress"] =
			`
		{"properties":{"address":{"type":"string","maxLength":34,"minLength":34},"amount":{"type":"integer"}},"required":["address","amount"]}
	`
		schemaLoader := gojsonschema.NewStringLoader(check[u])
		documentLoader := gojsonschema.NewGoLoader(payload)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			panic(err.Error())
		}

		if result.Valid() {
			log.FluentfContext(consts.LOGINFO, c, "The document is valid\n")
		} else {
			var errorList string
			for _, desc := range result.Errors() {
				errorList = errorList + fmt.Sprintf("%s. ", desc)

			}
			err := errors.New("There was a problem with the parameters in your JSON request. Please correct these errors : " + errorList)
			log.FluentfContext(consts.LOGERROR, c, err.Error())
			ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidDocument.Code, err)

			return m, err
		}
	}
	return m, nil
}
