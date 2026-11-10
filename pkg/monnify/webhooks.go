package monnify

import (
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type (
	// WebhookEvent represents the generic webhook structure sent by Monnify
	WebhookEvent struct {
		EventType string          `json:"eventType"`
		EventData json.RawMessage `json:"eventData"`
	}

	// SuccessfulTransactionEvent contains details of a successful transaction
	SuccessfulTransactionEvent struct {
		TransactionReference string                 `json:"transactionReference"`
		PaymentReference     string                 `json:"paymentReference"`
		AmountPaid           float64                `json:"amountPaid"`
		TotalPayable         float64                `json:"totalPayable"`
		PaidOn               string                 `json:"paidOn"`
		PaymentDescription   string                 `json:"paymentDescription"`
		PaymentMethod        string                 `json:"paymentMethod"`
		CurrencyCode         string                 `json:"currencyCode"`
		PaymentStatus        string                 `json:"paymentStatus"`
		Customer             CustomerInfo           `json:"customer"`
		Product              ProductInfo            `json:"product"`
		MetaData             map[string]interface{} `json:"metaData"`
	}

	// CustomerInfo represents the customer information in transaction event
	CustomerInfo struct {
		Name         string `json:"name"`
		Email        string `json:"email"`
		MobileNumber string `json:"mobileNumber"`
	}

	// ProductInfo represents product/reference details in transaction event
	ProductInfo struct {
		Reference string `json:"reference"`
		Type      string `json:"type"`
	}

	// SuccessfulDisbursementEvent contains details of disbursement events
	SuccessfulDisbursementEvent struct {
		TransactionReference string  `json:"transactionReference"`
		Amount               float64 `json:"amount"`
		Status               string  `json:"status"`
		Reference            string  `json:"reference"`
		Currency             string  `json:"currency"`
		BankCode             string  `json:"bankCode"`
		AccountNumber        string  `json:"accountNumber"`
	}
)

// VerifySignature verifies the monnify signature header against the raw body using the client's secret key.
func (c *Client) VerifySignature(rawBody []byte, signature string) bool {
	h := hmac.New(sha512.New, []byte(c.secretKey))
	h.Write(rawBody)
	computedSignature := hex.EncodeToString(h.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(computedSignature), []byte(signature)) == 1
}

// DecodeWebhookEvent decodes the raw webhook payload into a WebhookEvent
func DecodeWebhookEvent(payload []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to decode webhook event: %w", err)
	}
	return &event, nil
}

// DecodeSuccessfulTransaction decodes event data for SUCCESSFUL_TRANSACTION
func (e *WebhookEvent) DecodeSuccessfulTransaction() (*SuccessfulTransactionEvent, error) {
	var data SuccessfulTransactionEvent
	if err := json.Unmarshal(e.EventData, &data); err != nil {
		return nil, fmt.Errorf("failed to decode successful transaction event data: %w", err)
	}
	return &data, nil
}

// DecodeSuccessfulDisbursement decodes event data for SUCCESSFUL_DISBURSEMENT / FAILED_DISBURSEMENT / REVERSED_DISBURSEMENT
func (e *WebhookEvent) DecodeSuccessfulDisbursement() (*SuccessfulDisbursementEvent, error) {
	var data SuccessfulDisbursementEvent
	if err := json.Unmarshal(e.EventData, &data); err != nil {
		return nil, fmt.Errorf("failed to decode successful disbursement event data: %w", err)
	}
	return &data, nil
}
