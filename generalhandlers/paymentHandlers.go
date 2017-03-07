package generalhandlers

import (
	"encoding/json"
	"net/http"

	"github.com/vennd/enu/bitcoinapi"
	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/log"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

func GetPayment(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var payment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	payment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	if paymentId == "" || len(paymentId) < 16 {
		log.FluentfContext(consts.LOGERROR, c, "Invalid paymentId")
		handlers.ReturnBadRequest(c, w, consts.GenericErrors.InvalidPaymentId.Code, consts.GenericErrors.InvalidPaymentId.Description)

		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "GetPayment called for '%s' by '%s'\n", paymentId, c.Value(consts.AccessKeyKey).(string))

	payment = database.GetPaymentByPaymentId(c, c.Value(consts.AccessKeyKey).(string), paymentId)
	// errorhandling here!!

	// Add the blockchain status
	if payment.BroadcastTxId != "" && payment.BlockchainId == consts.CounterpartyBlockchainId {
		confirmations, err := bitcoinapi.GetConfirmations(payment.BroadcastTxId)
		if err == nil || confirmations == 0 {
			payment.BlockchainStatus = "unconfimed"
			payment.BlockchainConfirmations = 0
		}

		payment.BlockchainStatus = "confirmed"
		payment.BlockchainConfirmations = confirmations
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payment); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	return nil
}

func GetPaymentsByAddress(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var payment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	payment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" {
		log.FluentfContext(consts.LOGERROR, c, "Invalid address")
		handlers.ReturnBadRequest(c, w, consts.GenericErrors.InvalidAddress.Code, consts.GenericErrors.InvalidAddress.Description)

		return nil
	}

	log.FluentfContext(consts.LOGINFO, c, "GetPaymentByAddress called for '%s' by '%s'\n", address, c.Value(consts.AccessKeyKey).(string))

	payments := database.GetPaymentsByAddress(c, c.Value(consts.AccessKeyKey).(string), address)
	// errorhandling here!!

	// Add the blockchain status

	for i, p := range payments {
		if p.BroadcastTxId != "" && payment.BlockchainId == consts.CounterpartyBlockchainId {
			confirmations, err := bitcoinapi.GetConfirmations(p.BroadcastTxId)
			if err == nil || confirmations == 0 {
				payments[i].BlockchainStatus = "unconfimed"
				payments[i].BlockchainConfirmations = 0
			}

			payments[i].BlockchainStatus = "confirmed"
			payments[i].BlockchainConfirmations = confirmations
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payments); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	return nil
}
