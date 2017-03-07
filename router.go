package main

import (
	//	"net/http"

	"github.com/vennd/enu/handlers"
	"github.com/vennd/enu/internal/github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", handlers.Index).Methods("GET")
	router.HandleFunc("/serverinfo", handlers.Serverinfo).Methods("GET")

	router.Handle("/payment", ctxHandler(PaymentCreate)).Methods("POST")
	router.Handle("/payment/address", ctxHandler(AddressCreate)).Methods("POST")
	router.Handle("/payment/address/{address}", ctxHandler(GetPaymentsByAddress)).Methods("GET")
	router.Handle("/payment/{paymentId}", ctxHandler(GetPayment)).Methods("GET")
	router.Handle("/payment/status/{paymentId}", ctxHandler(PaymentRetry)).Methods("POST")

	router.Handle("/asset", ctxHandler(AssetCreate)).Methods("POST")
	router.Handle("/asset/{assetId}", ctxHandler(GetAsset)).Methods("GET")
	router.Handle("/asset/dividend", ctxHandler(DividendCreate)).Methods("POST")
	router.Handle("/asset/dividend/{dividendId}", ctxHandler(GetDividend)).Methods("GET")
	router.Handle("/asset/issuances/{asset}", ctxHandler(AssetIssuances)).Methods("GET")
	router.Handle("/asset/ledger/{asset}", ctxHandler(AssetLedger)).Methods("GET")

	router.Handle("/wallet", ctxHandler(WalletCreate)).Methods("POST")
	router.Handle("/wallet/balances/{address}", ctxHandler(WalletBalance)).Methods("GET")
	router.Handle("/wallet/payment", ctxHandler(WalletSend)).Methods("POST")
	router.Handle("/wallet/payment/{paymentId}", ctxHandler(GetPayment)).Methods("GET")
	router.Handle("/wallet/activate/address/{address}", ctxHandler(ActivateAddress)).Methods("POST")

	// Direct access to Counterparty resources
	router.Handle("/counterparty/asset", ctxHandler(AssetCreate)).Methods("POST")
	router.Handle("/counterparty/asset/{assetId}", ctxHandler(GetAsset)).Methods("GET")
	router.Handle("/counterparty/asset/dividend", ctxHandler(DividendCreate)).Methods("POST")
	router.Handle("/counterparty/asset/dividend/{dividendId}", ctxHandler(GetDividend)).Methods("GET")
	router.Handle("/counterparty/asset/issuances/{asset}", ctxHandler(AssetIssuances)).Methods("GET")
	router.Handle("/counterparty/asset/ledger/{asset}", ctxHandler(AssetLedger)).Methods("GET")
	router.Handle("/counterparty/wallet", ctxHandler(WalletCreate)).Methods("POST")
	router.Handle("/counterparty/wallet/balances/{address}", ctxHandler(WalletBalance)).Methods("GET")
	router.Handle("/counterparty/wallet/payment", ctxHandler(WalletSend)).Methods("POST")
	router.Handle("/counterparty/wallet/payment/{paymentId}", ctxHandler(GetPayment)).Methods("GET")
	router.Handle("/counterparty/wallet/activate/address/{address}", ctxHandler(ActivateAddress)).Methods("POST")
	router.Handle("/counterparty/payment/address/{address}", ctxHandler(GetPaymentsByAddress)).Methods("GET")

	router.Handle("/blocks", ctxHandler(GetBlocks)).Methods("GET")

	return router
}
