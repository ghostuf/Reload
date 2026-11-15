package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrEsewaInvalidPayload is returned when the base64 "data" eSewa
	// appends to success_url does not decode to the JSON object it should.
	ErrEsewaInvalidPayload = errors.New("could not decode eSewa response")
	// ErrEsewaSignatureMismatch is returned when a response's own signature
	// does not match the one recomputed from its signed_field_names.
	ErrEsewaSignatureMismatch = errors.New("eSewa response signature does not match")
)

// esewaRequestSignedFields is the fixed field order eSewa's docs specify for
// signing an outgoing payment request: total_amount, transaction_uuid, and
// product_code, in that order.
var esewaRequestSignedFields = []string{"total_amount", "transaction_uuid", "product_code"}

// esewaAmountFormat renders an amount exactly the way it must appear in both
// the form field and the signature message -- the two have to be the
// identical string, or eSewa's own recomputed signature will never match
// the one sent alongside it.
func esewaAmountFormat(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

// generateTransactionUUID returns a value satisfying eSewa's one constraint
// on it -- letters, numbers, and hyphens only. The timestamp prefix makes it
// recognisable in eSewa's own merchant dashboard; the random suffix is what
// actually guarantees uniqueness, backed by the database's unique index on
// the column as well.
func generateTransactionUUID() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating transaction uuid: %w", err)
	}
	return fmt.Sprintf("%s-%s", time.Now().UTC().Format("20060102150405"), hex.EncodeToString(buf)), nil
}

// esewaSignatureMessage builds the comma-joined "name=value" string eSewa's
// HMAC is computed over: exactly the fields named in signedFieldNames, in
// that order. This is the one format both signing an outgoing request and
// verifying an incoming response share.
func esewaSignatureMessage(signedFieldNames []string, fields map[string]string) string {
	parts := make([]string, len(signedFieldNames))
	for i, name := range signedFieldNames {
		parts[i] = name + "=" + fields[name]
	}
	return strings.Join(parts, ",")
}

// esewaSign returns the base64-encoded HMAC-SHA256 of message under secret
// -- the exact algorithm eSewa's docs specify for both directions: signing
// a request, and verifying the signature on its response.
func esewaSign(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// EsewaFormFields is every field the frontend must render as hidden inputs
// and POST to eSewa's form URL.
type EsewaFormFields struct {
	Amount                string `json:"amount"`
	TaxAmount             string `json:"tax_amount"`
	TotalAmount           string `json:"total_amount"`
	TransactionUUID       string `json:"transaction_uuid"`
	ProductCode           string `json:"product_code"`
	ProductServiceCharge  string `json:"product_service_charge"`
	ProductDeliveryCharge string `json:"product_delivery_charge"`
	SuccessURL            string `json:"success_url"`
	FailureURL            string `json:"failure_url"`
	SignedFieldNames      string `json:"signed_field_names"`
	Signature             string `json:"signature"`
}

// buildEsewaFormFields signs amount for transactionUUID under productCode
// and secret, and assembles every field the frontend needs to submit to
// eSewa's form URL. Reload charges no tax, service, or delivery fee, so
// amount and total_amount are always the same value -- but the signature
// itself is computed over total_amount, transaction_uuid, and product_code
// specifically, per eSewa's spec, never over amount.
func buildEsewaFormFields(secret, productCode, transactionUUID string, amount float64, successURL, failureURL string) EsewaFormFields {
	amountStr := esewaAmountFormat(amount)
	zero := esewaAmountFormat(0)

	fields := map[string]string{
		"total_amount":     amountStr,
		"transaction_uuid": transactionUUID,
		"product_code":     productCode,
	}
	signature := esewaSign(secret, esewaSignatureMessage(esewaRequestSignedFields, fields))

	return EsewaFormFields{
		Amount:                amountStr,
		TaxAmount:             zero,
		TotalAmount:           amountStr,
		TransactionUUID:       transactionUUID,
		ProductCode:           productCode,
		ProductServiceCharge:  zero,
		ProductDeliveryCharge: zero,
		SuccessURL:            successURL,
		FailureURL:            failureURL,
		SignedFieldNames:      strings.Join(esewaRequestSignedFields, ","),
		Signature:             signature,
	}
}

// decodeEsewaCallback base64-decodes and JSON-unmarshals the "data" query
// parameter eSewa appends to success_url, into a flat string map. Every
// field in a real eSewa response -- including ones that look numeric, like
// total_amount -- is itself a JSON string, which is the one shape that can
// hold whichever fields signed_field_names happens to list without this
// code having to guess them up front.
func decodeEsewaCallback(data string) (map[string]string, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEsewaInvalidPayload, err)
	}

	var fields map[string]string
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEsewaInvalidPayload, err)
	}
	return fields, nil
}

// verifyEsewaSignature recomputes the signature over exactly the fields
// named in the response's own signed_field_names -- not the fixed three an
// outgoing request uses -- and compares it against the response's
// signature in constant time, so a tampered field and a forged signature
// are both rejected the same way: the comparison just does not match.
func verifyEsewaSignature(secret string, fields map[string]string) bool {
	signedFieldNames := strings.Split(fields["signed_field_names"], ",")
	message := esewaSignatureMessage(signedFieldNames, fields)
	expected := esewaSign(secret, message)

	return subtle.ConstantTimeCompare([]byte(expected), []byte(fields["signature"])) == 1
}
