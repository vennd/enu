package counterpartyhandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/vennd/enu/internal/github.com/gorilla/mux"
	"github.com/vennd/enu/internal/golang.org/x/net/context"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/log"
)

func PaymentCreate(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var simplePayment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	simplePayment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	paymentId := m["paymentId"].(string)
	sourceAddress := m["sourceAddress"].(string)
	destinationAddress := m["destinationAddress"].(string)
	asset := m["asset"].(string)
	amount := uint64(m["amount"].(float64))
	txFee := uint64(m["txFee"].(float64))
	paymentTag := m["paymentTag"].(string)

	if m["paymentTag"] != nil {
		paymentTag = m["paymentTag"].(string)
	}

	// If a paymentId is not specified, generate one
	if paymentId == "" {
		paymentId = enulib.GeneratePaymentId()
		simplePayment.PaymentId = paymentId
		log.FluentfContext(consts.LOGINFO, c, "Generated paymentId: %s", simplePayment.PaymentId)
	}

	database.InsertPayment(c, c.Value(consts.AccessKeyKey).(string), 0, c.Value(consts.BlockchainIdKey).(string), paymentId, sourceAddress, destinationAddress, asset, "", amount, "Authorized", 0, txFee, paymentTag)
	// errorhandling here!!

	simplePayment.SourceAddress = sourceAddress
	simplePayment.DestinationAddress = destinationAddress
	simplePayment.Asset = asset
	simplePayment.Amount = amount
	simplePayment.TxFee = int64(txFee)
	simplePayment.PaymentTag = paymentTag

	// Return to the client the paymentId
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(simplePayment); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	return nil
}

func PaymentRetry(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {

	var payment enulib.SimplePayment
	requestId := c.Value(consts.RequestIdKey).(string)
	payment.RequestId = requestId
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	paymentId := vars["paymentId"]

	log.FluentfContext(consts.LOGINFO, c, "PaymentRetry called for paymentId %s\n", paymentId)
	payment = database.GetPaymentByPaymentId(c, c.Value(consts.AccessKeyKey).(string), paymentId)

	// Payment not found
	if payment.Status == "Not found" || payment.Status == "" {
		errorString := fmt.Sprintf("PaymentId: %s not found", paymentId)
		log.FluentfContext(consts.LOGERROR, c, errorString)
		handlers.ReturnNotFoundWithCustomError(c, w, consts.GenericErrors.NotFound.Code, errorString)
		return nil
	}

	// Payment isn't in an error state or manual state
	if payment.Status != "error" && payment.Status != "manual" {
		errorString := fmt.Sprintf("PaymentId: %s is not in an 'error' or 'manual' state. It is in '%s' state.", paymentId, payment.Status)
		log.FluentfContext(consts.LOGINFO, c, errorString)
		handlers.ReturnNotFoundWithCustomError(c, w, consts.GenericErrors.NotFound.Code, errorString)
		return nil
	}

	err := database.UpdatePaymentStatusByPaymentId(c, c.Value(consts.AccessKeyKey).(string), paymentId, "authorized")
	if err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in UpdatePaymentStatusByPaymentId(): %s", err.Error())
		handlers.ReturnUnprocessableEntity(c, w, consts.GenericErrors.GeneralError.Code, errors.New(consts.GenericErrors.GeneralError.Description))
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payment); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	return nil
}
