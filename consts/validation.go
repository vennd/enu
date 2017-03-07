package consts

type Validations map[string]string

// cf http://spacetelescope.github.io/understanding-json-schema/
var ParameterValidations = map[string]Validations{
	"counterparty": {
		"asset":           `{"properties":{"blockchainId":{"type":"string"},"passphrase":{"type":"string"},"distributionAddress":{"type":"string","maxLength":34,"minLength":34},"distributionPassphrase":{"type":"string"},"description":{"type":"string"},"asset":{"type":"string","minLength":4},"quantity":{"type":"integer"},"divisible":{"type":"boolean"},"nonce":{"type":"integer"}},"required":["sourceAddress","passphrase","asset","quantity","divisible"]}`,
		"dividend":        `{"properties":{"blockchainId":{"type":"string"},"sourceAddress":{"type":"string","maxLength":34,"minLength":34},"asset":{"type":"string","minLength":4},"dividendAsset":{"type":"string"},"quantityPerUnit":{"type":"integer"},"nonce":{"type":"integer"}},"required":["sourceAddress","asset","dividendAsset","quantityPerUnit"]}`,
		"walletCreate":    `{"properties":{"blockchainId":{"type":"string"},"numberOfAddresses":{"type":"number","minimum":1,"maximum":100,"exclusiveMaximum":false},"nonce":{"type":"integer"}}}`,
		"walletPayment":   `{"properties":{"blockchainId":{"type":"string"},"sourceAddress":{"type":"string","maxLength":34,"minLength":34},"destinationAddress":{"type":"string","maxLength":34,"minLength":34},"asset":{"type":"string","minLength":4},"quantity":{"type":"integer"},"nonce":{"type":"integer"}},"required":["sourceAddress","asset","quantity","destinationAddress"]}`,
		"simplePayment":   `{"properties":{"sourceAddress":{"type":"string", "maxLength":34, "minLength":34},"destinationAddress":{"type":"string", "maxLength":34, "minLength":34},"asset":{"type":"string","minLength":4},"amount":{"type":"integer"},"txFee":{"type":"integer"}},"required":["sourceAddress","destinationAddress","asset","amount"]}`,
		"activateaddress": `{"properties":{"blockchainId":{"type":"string"},"address":{"type":"string","maxLength":34,"minLength":34},"amount":{"type":"integer"},"nonce":{"type":"integer"}},"required":["address","amount"]}`,
	},
	"ripple": {
		"asset":           `{"properties":{"blockchainId":{"type":"string"},"passphrase":{"type":"string"},"distributionAddress":{"type":"string"},"distributionPassphrase":{"type":"string"},"description":{"type":"string"},"asset":{"type":"string","minLength":4},"quantity":{"type":"integer"},"divisible":{"type":"boolean"},"nonce":{"type":"integer"}},"required":["sourceAddress","passphrase","asset","quantity","divisible"]}`,
		"walletCreate":    `{"properties":{"blockchainId":{"type":"string"},"nonce":{"type":"integer"}}}`,
		"walletPayment":   `{"properties":{"blockchainId":{"type":"string"},"sourceAddress":{"type":"string"},"destinationAddress":{"type":"string"},"asset":{"type":"string","minLength":3},"quantity":{"type":"integer"},"nonce":{"type":"integer"}},"required":["sourceAddress","asset","quantity","destinationAddress"]}`,
		"activateaddress": `{"properties":{"blockchainId":{"type":"string"},"address":{"type":"string"},"passphrase":{"type":"string"},"amount":{"type":"integer"},"assets":{"type":"array", "items": [{"type":"object","properties":{"currency":{"type":"string"},"issuer":{"type":"string"}}}]},"nonce":{"type":"integer"}},"required":["address","amount"]}`,
	},
}
