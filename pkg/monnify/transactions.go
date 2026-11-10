package monnify

import (
	"context"
	"fmt"
	"net/http"
)

type (
	// InitTransactionReq holds the request body parameters for initializing a transaction
	InitTransactionReq struct {
		Amount             float64                `json:"amount"`
		CustomerName       string                 `json:"customerName"`
		CustomerEmail      string                 `json:"customerEmail"`
		PaymentReference   string                 `json:"paymentReference"`
		PaymentDescription string                 `json:"paymentDescription"`
		CurrencyCode       string                 `json:"currencyCode"`
		ContractCode       string                 `json:"contractCode"`
		RedirectUrl        string                 `json:"redirectUrl"`
		PaymentMethods     []string               `json:"paymentMethods,omitempty"`
		Metadata           map[string]interface{} `json:"metadata,omitempty"`
	}

	// InitTransactionRes holds the response details from initializing a transaction
	InitTransactionRes struct {
		RequestSuccessful bool   `json:"requestSuccessful"`
		ResponseMessage   string `json:"responseMessage"`
		ResponseCode      string `json:"responseCode"`
		ResponseBody      struct {
			TransactionReference string `json:"transactionReference"`
			PaymentReference     string `json:"paymentReference"`
			CheckoutUrl          string `json:"checkoutUrl"`
		} `json:"responseBody"`
	}

	// TransactionStatusRes holds the details from querying transaction status
	TransactionStatusRes struct {
		RequestSuccessful bool   `json:"requestSuccessful"`
		ResponseMessage   string `json:"responseMessage"`
		ResponseCode      string `json:"responseCode"`
		ResponseBody      struct {
			TransactionReference string                 `json:"transactionReference"`
			PaymentReference     string                 `json:"paymentReference"`
			AmountPaid           JSONFloat64            `json:"amountPaid"`
			TotalPayable         JSONFloat64            `json:"totalPayable"`
			SettlementAmount     JSONFloat64            `json:"settlementAmount"`
			PaidOn               string                 `json:"paidOn"`
			CreatedOn            string                 `json:"createdOn"`
			CompletedOn          string                 `json:"completedOn"`
			PaymentStatus        string                 `json:"paymentStatus"`
			PaymentMethod        string                 `json:"paymentMethod"`
			CurrencyCode         string                 `json:"currencyCode"`
			Fee                  JSONFloat64            `json:"fee"`
			Customer             CustomerInfo           `json:"customerDTO"`
			MetaData             map[string]interface{} `json:"metaData"`
		} `json:"responseBody"`
	}
)

// InitTransaction initializes a payment transaction with Monnify
func (c *Client) InitTransaction(ctx context.Context, req InitTransactionReq) (*InitTransactionRes, error) {
	url := c.baseUrl + "/v1/merchant/transactions/init-transaction"

	var res InitTransactionRes
	if err := c.send(ctx, c, http.MethodPost, url, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// QueryTransaction retrieves the status of a transaction using either the paymentReference or transactionReference.
// If both are provided, transactionReference takes preference.
func (c *Client) QueryTransaction(ctx context.Context, paymentReference, transactionReference string) (*TransactionStatusRes, error) {
	var url string
	if transactionReference != "" {
		url = fmt.Sprintf("%s/v2/merchant/transactions/query?transactionReference=%s", c.baseUrl, transactionReference)
	} else if paymentReference != "" {
		url = fmt.Sprintf("%s/v2/merchant/transactions/query?paymentReference=%s", c.baseUrl, paymentReference)
	} else {
		return nil, fmt.Errorf("must provide either paymentReference or transactionReference")
	}

	var res TransactionStatusRes
	if err := c.send(ctx, c, http.MethodGet, url, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
