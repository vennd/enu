package ripplehandlers

import (
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
	"github.com/vennd/enu/log"
)

func Unhandled(c context.Context, w http.ResponseWriter, r *http.Request, m map[string]interface{}) *enulib.AppError {
	log.FluentfContext(consts.LOGINFO, c, "Unhandled function called: %s", c.Value(consts.RequestTypeKey).(string))
	handlers.ReturnNotFoundWithCustomError(c, w, consts.GenericErrors.FunctionNotAvailable.Code, consts.GenericErrors.FunctionNotAvailable.Description)

	return nil
}
