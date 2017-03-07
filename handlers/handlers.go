package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

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

var quotes = [...]string{"Here's to the crazy ones. The misfits. The rebels. The troublemakers. The round pegs in the square holes. The ones who see things differently. They're not fond of rules. And they have no respect for the status quo. You can quote them, disagree with them, glorify or vilify them. About the only thing you can't do is ignore them. Because they change things. They push the human race forward. And while some may see them as the crazy ones, we see genius. Because the people who are crazy enough to think they can change the world, are the ones who do. - Apple Inc.",
	"You miss 100% of the shots you don’t take. –Wayne Gretzky",
	"七転び八起き - Japanese proverb",
	"The problem is not the problem; the problem is your attitude about the probem. - Captain Jack Sparrow",
	"Only dead fish go with the flow.",
	"A friend is someone with whom you dare to be yourself. - Fran Crane",
	"Be yourself; everyone else is already taken. - Oscar Wilde",
	"Forty-two is a pronic number and an abundant number; its prime factorization 2 · 3 · 7 makes it the second sphenic number and also the second of the form { 2 · 3 · r }. As with all sphenic numbers of this form, the aliquot sum is abundant by 12. 42 is also the second sphenic number to be bracketed by twin primes; 30 is also a pronic number and also rests between two primes. 42 has a 14-member aliquot sequence 42, 54, 66, 78, 90, 144, 259, 45, 33, 15, 9, 4, 3, 1, 0 and is itself part of the aliquot sequence commencing with the first sphenic number 30. Further, 42 is the 10th member of the 3-aliquot tree.",
}

// Handles the '/serverinfo' path
func Serverinfo(w http.ResponseWriter, r *http.Request) {
	type version struct {
		Full       string `json:"full"`
		Major      uint32 `json:"major"`
		Minor      uint32 `json:"minor"`
		Patch      uint32 `json:"patch"`
		Prerelease string `json:"prerelease"`
		Tag        string `json:"tag"`
	}

	type ReleaseNote struct {
		IssueNumber           uint32 `json:"issueNumber"`
		InternalExternalIssue uint32 `json:"internalExternalIssue"`
		Description           string `json:"description"`
	}

	type serverinfo struct {
		Environment  string               `json:"env"`
		Version      version              `json:"version"`
		ReleaseNotes []enulib.ReleaseNote `json:"releaseNotes"`
	}

	var result = serverinfo{
		Version: version{Major: enulib.VersionMajor, Minor: enulib.VersionMinor, Patch: enulib.VersionPatch, Prerelease: enulib.VersionPrerelease, Tag: enulib.VersionTag},
	}

	// Populate a human readable string
	result.Version.Full = fmt.Sprintf("%d.%d.%d-%s+%s", enulib.VersionMajor, enulib.VersionMinor, enulib.VersionPatch, enulib.VersionPrerelease, enulib.VersionTag)

	// Populate env
	env := os.Getenv("ENV")
	if env == "" {
		env = "unspecified"
	}
	result.Environment = env

	// Populate release notes
	result.ReleaseNotes = enulib.ReleaseNotes

	// Return result as json
	j, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return
	}

	fmt.Fprintf(w, "%s\n", string(j))
}

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
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
	}
}

func ReturnBadRequest(c context.Context, w http.ResponseWriter, errorCode int64, errorString string) {

	returnCode := enulib.ReturnCode{Code: errorCode, Description: errorString, RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
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
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
	}
}

func ReturnCreated(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: 0, Description: "Success", RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
	}
}

func ReturnOK(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: 0, Description: "Success", RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
	}
}

func ReturnNotFound(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: consts.GenericErrors.NotFound.Code, Description: consts.GenericErrors.NotFound.Description, RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
	}
}

func ReturnNotFoundWithCustomError(c context.Context, w http.ResponseWriter, errorCode int64, errorString string) {
	returnCode := enulib.ReturnCode{Code: errorCode, Description: errorString, RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotFound)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
	}
}

func ReturnServerError(c context.Context, w http.ResponseWriter) {
	returnCode := enulib.ReturnCode{Code: consts.GenericErrors.GeneralError.Code, Description: consts.GenericErrors.GeneralError.Description, RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
	}
}

func ReturnServerErrorWithCustomError(c context.Context, w http.ResponseWriter, errorCode int64, errorString string) {
	returnCode := enulib.ReturnCode{Code: errorCode, Description: errorString, RequestId: c.Value(consts.RequestIdKey).(string)}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(returnCode); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
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

func CheckAndParseJsonCTX(c context.Context, w http.ResponseWriter, r *http.Request) (context.Context, map[string]interface{}, error) {
	//	var blockchainId string
	var payload interface{}
	var nonceDB int64

	signature := r.Header.Get("Signature")

	// Limit amount read to 512,000 bytes and parse body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 512000))
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		ReturnServerError(c, w)

		return c, nil, err
	}
	if err := r.Body.Close(); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Body.Close(): %s", err.Error())
		ReturnServerError(c, w)

		return c, nil, err
	}

	// If the body is an empty byte array then don't attempt to unmarshall the JSON and set a default
	if bytes.Compare(body, make([]byte, 0)) != 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			log.FluentfContext(consts.LOGINFO, c, "Malformed body: %s", string(body))
			returnErr := errors.New("The request did not contain a valid JSON object")
			log.FluentfContext(consts.LOGERROR, c, err.Error())                                   // Log the real error
			ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidDocument.Code, returnErr) // Send back sanitised error

			return c, nil, returnErr
		}
		log.FluentfContext(consts.LOGINFO, c, "Request received: %s", body)
	} else if bytes.Compare(body, make([]byte, 0)) == 0 {
		log.FluentfContext(consts.LOGINFO, c, "Empty body received")
		returnErr := errors.New("Empty body received. If you aren't sending a payload then send an empty JSON object: {}")
		log.FluentfContext(consts.LOGERROR, c, returnErr.Error())
		ReturnUnprocessableEntity(c, w, consts.GenericErrors.InvalidDocument.Code, returnErr)

		return c, nil, returnErr
	}

	// Then look up secret and calculate digest
	accessKey := c.Value(consts.AccessKeyKey).(string)
	calculatedSignature := enulib.ComputeHmac512(body, database.GetSecretByAccessKey(accessKey))

	// If we didn't receive the expected signature then raise a forbidden
	if calculatedSignature != signature {
		errorString := fmt.Sprintf("Could not verify HMAC signature. Expected: %s, received: %s", calculatedSignature, signature)
		err := errors.New(errorString)
		ReturnUnauthorised(c, w, consts.GenericErrors.InvalidSignature.Code, err)

		return c, nil, err
	}

	m := payload.(map[string]interface{})

	// nonce checking
	var nonceInt int64
	if m["nonce"] != nil {
		nonceInt := int64(m["nonce"].(float64))
		log.FluentfContext(consts.LOGINFO, c, "Nonce received: %d", nonceInt)
	} else {
		nonceInt = 0
	}

	if nonceInt > 0 {
		nonceDB = database.GetNonceByAccessKey(accessKey)
		if nonceInt <= nonceDB {
			//Nonce is not greater than the nonce in the DB
			log.FluentfContext(consts.LOGERROR, c, "Nonce for accessKey %s provided is <= nonce in db. %d <= %d\n", accessKey, nonceInt, nonceDB)
			ReturnUnauthorised(c, w, consts.GenericErrors.InvalidNonce.Code, errors.New(consts.GenericErrors.InvalidNonce.Description))

			return c, nil, err
		} else {
			log.FluentfContext(consts.LOGINFO, c, "Nonce for accessKey %s provided is ok. (%s > %d)\n", accessKey, nonceInt, nonceDB)
			database.UpdateNonce(accessKey, nonceInt)
			if err != nil {
				log.FluentfContext(consts.LOGERROR, c, "Nonce update failed, error: %s", err.Error())
				ReturnServerError(c, w)

				return c, nil, err
			}
		}
	}

	// Overwrite blockchain context if the blockchainId has been set as a parameter in the body
	var c2 context.Context
	if m["blockchainId"] != nil && m["blockchainId"] != "" {
		log.FluentfContext(consts.LOGINFO, c, "User specified in body requested blockchainId: %s", m["blockchainId"].(string))
		requestBlockchainId := m["blockchainId"].(string)

		// check if blockchainId is valid
		supportedBlockchains := consts.SupportedBlockchains
		sort.Strings(supportedBlockchains)

		i := sort.SearchStrings(supportedBlockchains, requestBlockchainId)
		blockchainValid := i < len(supportedBlockchains) && supportedBlockchains[i] == requestBlockchainId

		if blockchainValid {
			log.FluentfContext(consts.LOGINFO, c, "blockchainId specified as a body parameter. Overwriting blockchainId with: %s", m["blockchainId"].(string))
			c2 = context.WithValue(c, consts.BlockchainIdKey, requestBlockchainId)
		} else {
			log.FluentfContext(consts.LOGERROR, c, "Unsupported blockchainId: %s", m["blockchainId"].(string))
			ReturnBadRequest(c, w, consts.GenericErrors.UnsupportedBlockchain.Code, consts.GenericErrors.UnsupportedBlockchain.Description+" Given: "+m["blockchainId"].(string))
			return c, m, errors.New(consts.GenericErrors.UnsupportedBlockchain.Description)
		}
	} else {
		c2 = c
	}

	err = ValidateParameters(c2, payload)
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, err.Error())
		ReturnUnprocessableEntity(c2, w, consts.GenericErrors.InvalidDocument.Code, err)

		return c2, m, err
	}

	log.FluentfContext(consts.LOGINFO, c, "Parameters validated.")

	return c2, m, nil
}

func ValidateParameters(c context.Context, parameters interface{}) error {
	blockchainId := c.Value(consts.BlockchainIdKey).(string)
	u, ok := c.Value(consts.RequestTypeKey).(string)

	// Skip validation if a schema isn't found in the validations map
	if ok && consts.ParameterValidations[blockchainId][u] != "" {
		schemaLoader := gojsonschema.NewStringLoader(consts.ParameterValidations[blockchainId][u])
		documentLoader := gojsonschema.NewGoLoader(parameters)

		log.Printf("Validating against: %s\n", consts.ParameterValidations[blockchainId][u])

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			log.FluentfContext(consts.LOGERROR, c, "Error in gojsonschema.Validate(): %s", err.Error())
			return err
		}

		if result.Valid() {
			return nil
		} else {
			var errorList string
			for _, desc := range result.Errors() {
				errorList = errorList + fmt.Sprintf("%s. ", desc)

			}
			err := errors.New("There was a problem with the parameters in your JSON request. Please correct these errors : " + errorList)
			log.FluentfContext(consts.LOGERROR, c, err.Error())

			return err
		}
	}

	// shouldn't reach here...
	return nil
}
