package main

import (
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

func AddressCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "address")

	return handle(c, w, r)
}
