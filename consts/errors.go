package consts

const SqlNotFound = "sql: no rows in result set"

const CounterpartylibInsufficientFunds = "insufficient funds"
const CounterpartylibMalformedAddress = "Odd-length string"
const CounterpartylibInsufficientBTC = "Insufficient BTC at address"
const CounterpartylibOnlyIssuerCanPayDividends = "only issuer can pay dividends"
const CountpartylibNoSuchAsset = "no such asset"
const CountpartylibMempoolIsNotReady = "Mempool is not yet ready"

type ErrCodes struct {
	Code        int64
	Description string
}

type CounterpartyStruct struct {
	MiscError                 ErrCodes
	Timeout                   ErrCodes
	ReparsingOrUnavailable    ErrCodes
	SigningError              ErrCodes
	BroadcastError            ErrCodes
	InsufficientFunds         ErrCodes
	InsufficientFees          ErrCodes
	InvalidPassphrase         ErrCodes
	DividendNotFound          ErrCodes
	ComposeError              ErrCodes
	MalformedAddress          ErrCodes
	OnlyIssuerCanPayDividends ErrCodes
	NoSuchAsset               ErrCodes
}

var CounterpartyErrors = CounterpartyStruct{
	MiscError:                 ErrCodes{1000, "Misc error when contacting Counterparty. Please contact Vennd.io support."},
	Timeout:                   ErrCodes{1001, "Timeout when contacting Counterparty. Please try again later."},
	ReparsingOrUnavailable:    ErrCodes{1002, "Counterparty Blockchain temporarily unavailable. Please try again later."},
	SigningError:              ErrCodes{1003, "Unable to sign transaction. Is your passphrase correct?"},
	BroadcastError:            ErrCodes{1004, "Unable to broadcast transaction to the blockchain. Please try the transaction again."},
	InsufficientFees:          ErrCodes{1005, "Insufficient BTC in address to perform transaction. Please use the Activate() call to add more BTC."},
	DividendNotFound:          ErrCodes{1007, "The dividend could not be found."},
	ComposeError:              ErrCodes{1008, "Unable to create the blockchain transaction."},
	InsufficientFunds:         ErrCodes{1009, "Insufficient asset in this address."},
	MalformedAddress:          ErrCodes{1010, "One of the addresses provided was not correct. Please check the addresses involved in the transaction."},
	OnlyIssuerCanPayDividends: ErrCodes{1011, "Only the issuer may pay dividends."},
	NoSuchAsset:               ErrCodes{1012, "The asset specified is incorrect or doesn't exist."},
}

type GenericStruct struct {
	InvalidDocument       ErrCodes
	InvalidDividendId     ErrCodes
	UnsupportedBlockchain ErrCodes
	HeadersIncorrect      ErrCodes
	UnknownAccessKey      ErrCodes
	InvalidSignature      ErrCodes
	InvalidNonce          ErrCodes
	NotFound              ErrCodes
	FunctionNotAvailable  ErrCodes
	InvalidPassphrase     ErrCodes
	InvalidAssetId        ErrCodes
	InvalidPaymentId      ErrCodes
	InvalidAddress        ErrCodes
	InvalidAsset          ErrCodes
	ApiKeyDisabled        ErrCodes

	GeneralError ErrCodes
}

var GenericErrors = GenericStruct{
	InvalidDocument:       ErrCodes{1, "There was a problem with the parameters in your JSON request. Please correct the request."},
	InvalidDividendId:     ErrCodes{2, "The specified dividend id is invalid."},
	UnsupportedBlockchain: ErrCodes{3, "The specified blockchain is not supported."},
	HeadersIncorrect:      ErrCodes{4, "Request headers were not set correctly ensure the following headers are set: accessKey and signature."},
	UnknownAccessKey:      ErrCodes{5, "Attempt to access API with unknown user key"},
	InvalidSignature:      ErrCodes{6, "Could not verify HMAC signature"},
	InvalidNonce:          ErrCodes{7, "Invalid nonce"},
	NotFound:              ErrCodes{8, "Not found"},
	FunctionNotAvailable:  ErrCodes{9, "The function is not available on the selected blockchain."},
	InvalidPassphrase:     ErrCodes{10, "The passphrase provided is not valid."},
	InvalidAssetId:        ErrCodes{11, "The specified asset id is invalid."},
	InvalidPaymentId:      ErrCodes{12, "The specified paymentId is invalid. Please correct the paymentId and resubmit."},
	GeneralError:          ErrCodes{13, "Misc error. Please contact Vennd.io support."},
	InvalidAddress:        ErrCodes{14, "The specified address is invalid. Please correct the address and resubmit."},
	InvalidAsset:          ErrCodes{15, "The specified asset is invalid. Please correct the asset and resubmit."},
	ApiKeyDisabled:        ErrCodes{16, "The specified API is valid. However it has been disabled by an administrator."},
}

type RippleStruct struct {
	MiscError                     ErrCodes
	Timeout                       ErrCodes
	InvalidAmount                 ErrCodes
	InvalidCurrency               ErrCodes
	SubmitError                   ErrCodes
	IssuerMustBeGiven             ErrCodes
	SigningError                  ErrCodes
	SubmitErrorFeeLost            ErrCodes
	InvalidCurrencyOrNoTrustline  ErrCodes
	InvalidSource                 ErrCodes
	InvalidDestination            ErrCodes
	DistributionPassphraseMissing ErrCodes
	DistributionInsufficientFunds ErrCodes
	InsufficientXRP               ErrCodes
}

var RippleErrors = RippleStruct{
	MiscError:                     ErrCodes{2000, "Misc error when contacting Ripple. Please contact Vennd.io support."},
	Timeout:                       ErrCodes{2001, "Timeout when contacting Ripple. Please try again later."},
	InvalidAmount:                 ErrCodes{2002, "The amount specified is not a valid amount."},
	InvalidCurrency:               ErrCodes{2003, "The currency is invalid. Ripple currencies must be 3 characters or longer."},
	SubmitError:                   ErrCodes{2004, "The Ripple node rejected the transaction submission. Please try again."},
	IssuerMustBeGiven:             ErrCodes{2005, "If the currency is not XRP the issuer must be provided."},
	SigningError:                  ErrCodes{2006, "Unable to sign transaction. Is your passphrase correct?"},
	SubmitErrorFeeLost:            ErrCodes{2007, "The transaction was submitted to the Ripple network but was invalid."},
	InvalidCurrencyOrNoTrustline:  ErrCodes{2008, "The specified asset is invalid or you must activate the destination wallet to accept the asset."},
	InvalidSource:                 ErrCodes{2009, "The specified source address is invalid."},
	InvalidDestination:            ErrCodes{2010, "The specified destination address is invalid."},
	DistributionPassphraseMissing: ErrCodes{2011, "If a distribution address is specified the passphrase for the distribution address must be given."},
	DistributionInsufficientFunds: ErrCodes{2012, "The specified distribution address does not contain sufficient funds. Please activate the address and try again."},
	InsufficientXRP:               ErrCodes{2013, "There was insufficient XRP in the address to perform the payment. Please activate the address and try again."},
}
