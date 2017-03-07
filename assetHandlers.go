package main

import (
	"net/http"

	"github.com/vennd/enu/consts"
	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/internal/golang.org/x/net/context"
)

func AssetCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "asset")

	return handle(c, w, r)
}

func DividendCreate(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "dividend")

	return handle(c, w, r)
}

func AssetIssuances(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "issuances") //new

	return handle(c, w, r)
}

// Recommended call which summarises the ledger for a particular asset
func AssetLedger(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "ledger")

	return handle(c, w, r)
}

func GetDividend(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "getdividend") //new

	return handle(c, w, r)
}

func GetAsset(c context.Context, w http.ResponseWriter, r *http.Request) *enulib.AppError {
	// Add to the context the RequestType
	c = context.WithValue(c, consts.RequestTypeKey, "getasset") //new

	return handle(c, w, r)
}
