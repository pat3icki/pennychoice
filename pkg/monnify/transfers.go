package monnify

import (
	"context"
	"fmt"
	"net/http"
)

type (
	// SingleTransferReq request parameters for initiating a single transfer
	SingleTransferReq struct {
		Amount                   float64 `json:"amount"`
		Reference                string  `json:"reference"`
		Narration                string  `json:"narration"`
		DestinationBankCode      string  `json:"destinationBankCode"`
		DestinationAccountNumber string  `json:"destinationAccountNumber"`
		DestinationAccountName   string  `json:"destinationAccountName"`
		SourceAccountNumber      string  `json:"sourceAccountNumber"`
		Currency                 string  `json:"currency"`
		Async                    bool    `json:"async,omitempty"`
	}

	// SingleTransferRes response parameters for initiating a single transfer
	SingleTransferRes struct {
		RequestSuccessful bool   `json:"requestSuccessful"`
		ResponseMessage   string `json:"responseMessage"`
		ResponseCode      string `json:"responseCode"`
		ResponseBody      struct {
			Amount                   JSONFloat64 `json:"amount"`
			Reference                string      `json:"reference"`
			Status                   string      `json:"status"`
			TransactionReference     string      `json:"transactionReference"`
			DateCreated              string      `json:"dateCreated"`
			TotalFee                 JSONFloat64 `json:"totalFee"`
			DestinationAccountName   string      `json:"destinationAccountName"`
			DestinationAccountNumber string      `json:"destinationAccountNumber"`
			DestinationBankCode      string      `json:"destinationBankCode"`
			DestinationBankName      string      `json:"destinationBankName"`
		} `json:"responseBody"`
	}

	// ValidateOtpReq request body for validating disbursement OTP
	ValidateOtpReq struct {
		Reference         string `json:"reference"`
		AuthorizationCode string `json:"authorizationCode"`
	}

	// ValidateOtpRes response body of OTP validation
	ValidateOtpRes struct {
		RequestSuccessful bool   `json:"requestSuccessful"`
		ResponseMessage   string `json:"responseMessage"`
		ResponseCode      string `json:"responseCode"`
		ResponseBody      struct {
			Amount                   JSONFloat64 `json:"amount"`
			Reference                string      `json:"reference"`
			Status                   string      `json:"status"`
			TransactionReference     string      `json:"transactionReference"`
			DateCreated              string      `json:"dateCreated"`
			DestinationAccountNumber string      `json:"destinationAccountNumber"`
			DestinationBankName      string      `json:"destinationBankName"`
		} `json:"responseBody"`
	}
)

// InitiateSingleTransfer starts a single transfer operation
func (c *Client) InitiateSingleTransfer(ctx context.Context, req SingleTransferReq) (*SingleTransferRes, error) {
	url := c.baseUrl + "/v2/disbursements/single"

	var res SingleTransferRes
	if err := c.send(ctx, c, http.MethodPost, url, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetSingleTransferStatus retrieves the status summary of a single transfer operation using the merchant reference
func (c *Client) GetSingleTransferStatus(ctx context.Context, reference string) (*SingleTransferRes, error) {
	url := fmt.Sprintf("%s/v2/disbursements/single/summary?reference=%s", c.baseUrl, reference)

	var res SingleTransferRes
	if err := c.send(ctx, c, http.MethodGet, url, nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ValidateSingleTransferOtp authorizes a pending single transfer using multi-factor authorization (OTP) code
func (c *Client) ValidateSingleTransferOtp(ctx context.Context, req ValidateOtpReq) (*ValidateOtpRes, error) {
	url := c.baseUrl + "/v2/disbursements/single/validate-otp"

	var res ValidateOtpRes
	if err := c.send(ctx, c, http.MethodPost, url, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
