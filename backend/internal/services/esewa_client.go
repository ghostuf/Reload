package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// EsewaStatusComplete is the one status value that counts as a paid
// transaction. Every other value -- PENDING, CANCELED, AMBIGUOUS, a refund,
// or anything else eSewa might return -- is treated as not yet paid.
const EsewaStatusComplete = "COMPLETE"

// EsewaStatusResult is what CheckStatus reports back: the provider's own
// status string, its reference for the transaction (nil unless COMPLETE or
// refunded), and the raw response body to keep for support and debugging.
type EsewaStatusResult struct {
	Status      string
	TotalAmount float64
	RefID       *string
	Raw         []byte
}

// EsewaStatusChecker confirms a payment independently of whatever the
// success callback claims, by asking eSewa directly. The production
// implementation is EsewaClient, calling eSewa's own status-check API over
// HTTP; tests substitute a fake that returns a canned result with no
// network call at all.
type EsewaStatusChecker interface {
	CheckStatus(productCode, transactionUUID string, totalAmount float64) (EsewaStatusResult, error)
}

// esewaStatusResponse is the shape of eSewa's status-check JSON response.
// Unlike the success-callback payload (where every field, including
// amounts, is a JSON string), this one comes straight from eSewa's REST API
// and uses real JSON types.
type esewaStatusResponse struct {
	ProductCode     string  `json:"product_code"`
	TransactionUUID string  `json:"transaction_uuid"`
	TotalAmount     float64 `json:"total_amount"`
	Status          string  `json:"status"`
	RefID           *string `json:"ref_id"`
}

// EsewaClient calls eSewa's transaction status-check API.
type EsewaClient struct {
	statusURL string
	http      *http.Client
}

func NewEsewaClient(statusURL string) *EsewaClient {
	return &EsewaClient{
		statusURL: statusURL,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

// CheckStatus asks eSewa directly whether transactionUUID actually
// completed, rather than trusting the success callback alone -- the
// callback travels through the user's own browser and could be replayed or
// forged, where this request goes straight from this server to eSewa's.
func (c *EsewaClient) CheckStatus(productCode, transactionUUID string, totalAmount float64) (EsewaStatusResult, error) {
	query := url.Values{
		"product_code":     {productCode},
		"total_amount":     {esewaAmountFormat(totalAmount)},
		"transaction_uuid": {transactionUUID},
	}

	sep := "?"
	if strings.Contains(c.statusURL, "?") {
		sep = "&"
	}

	resp, err := c.http.Get(c.statusURL + sep + query.Encode())
	if err != nil {
		return EsewaStatusResult{}, fmt.Errorf("calling eSewa status check: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return EsewaStatusResult{}, fmt.Errorf("reading eSewa status check response: %w", err)
	}

	var parsed esewaStatusResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return EsewaStatusResult{}, fmt.Errorf("decoding eSewa status check response: %w", err)
	}

	return EsewaStatusResult{
		Status:      parsed.Status,
		TotalAmount: parsed.TotalAmount,
		RefID:       parsed.RefID,
		Raw:         raw,
	}, nil
}
