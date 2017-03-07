package main

import (
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"

	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

func WalletCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletCreate")

	return handle(c, w, r)
}

func WalletBalance(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletBalance")

	return handle(c, w, r)
}

func WalletSend(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "walletPayment")

	return handle(c, w, r)
}

func ActivateAddress(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "activateaddress")

	return handle(c, w, r)
}
