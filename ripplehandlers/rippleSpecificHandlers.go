package ripplehandlers

import (
	"encoding/json"
	"net/http"

	"github.com/vennd/enu/internal/golang.org/x/net/context"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/log"
	"github.com/vennd/enu/rippleapi"
)

func GetRippleLedgerStatus(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get ripple ledger status
	result, _, err := rippleapi.GetLatestValidatedLedger(c)
	if err != nil {
		handlers.ReturnServerError(c, w)
		return nil
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(result); err != nil {
		log.FluentfContext(consts.LOGERROR, c, "Error in Encode(): %s", err.Error())
		handlers.ReturnServerError(c, w)

		return nil
	}

	return nil
}