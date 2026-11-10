package monnify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	// ValidateReq is the interface representing a validation request body
	ValidateReq interface {
		Encode() ([]byte, error)
	}

	// ValiddateResult holds the outcome of a validation operation
	ValiddateResult struct {
		Err  error
		Mode string
		Data any
	}

	// ValiddateBankRequest represents a request to validate a bank account
	ValiddateBankRequest struct {
		AccountNumber string `json:"accountNumber"`
		BankCode      string `json:"bankCode"`
	}

	// ValidateBankAccountResponese represents response of bank account validation
	ValidateBankAccountResponese struct {
		RequestSuccessful bool   `json:"requestSuccessful"`
		Message           string `json:"responseMessage"`
		ResponseCode      string `json:"responseCode"`
		Data              struct {
			AccountName   string `json:"accountName"`
			AccountNumber string `json:"accountNumber"`
			BankCode      string `json:"bankCode"`
			BankName      string `json:"bankName"`
		} `json:"responseBody"`
	}

	// ValidateBvnRequest represents request to match BVN against bank details
	ValidateBvnRequest struct {
		Bvn           string `json:"bvn"`
		BankCode      string `json:"bankCode"`
		AccountNumber string `json:"accountNumber"`
	}

	// ValidateBvnResponse represents response of BVN account match
	ValidateBvnResponse struct {
		RequestSuccessful bool   `json:"requestSuccessful"`
		ResponseMessage   string `json:"responseMessage"`
		ResponseCode      string `json:"responseCode"`
		ResponseBody      struct {
			Bvn           string `json:"bvn"`
			AccountNumber string `json:"accountNumber"`
			AccountName   string `json:"accountName"`
			MatchStatus   string `json:"matchStatus"`
		} `json:"responseBody"`
	}
)

// Encode encodes ValiddateBankRequest to bytes
func (r *ValiddateBankRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// Encode encodes ValidateBvnRequest to bytes
func (r *ValidateBvnRequest) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// Validate performs validation operations based on the specified mode ("bank", "bvn").
func (c *Client) Validate(ctx context.Context, mode string, body ValidateReq) ValiddateResult {
	var valResult ValiddateResult
	valResult.Mode = mode

	switch mode {
	case "bank":
		req, ok := body.(*ValiddateBankRequest)
		if !ok {
			valResult.Err = fmt.Errorf("invalid request type for bank validation")
			return valResult
		}

		url := fmt.Sprintf("%s/v2/disbursements/account/validate?accountNumber=%s&bankCode=%s", c.baseUrl, req.AccountNumber, req.BankCode)
		var res ValidateBankAccountResponese
		if err := c.send(ctx, c, http.MethodGet, url, nil, &res); err != nil {
			valResult.Err = err
			return valResult
		}
		valResult.Data = &res

	case "bvn":
		req, ok := body.(*ValidateBvnRequest)
		if !ok {
			valResult.Err = fmt.Errorf("invalid request type for bvn validation")
			return valResult
		}

		url := c.baseUrl + "/v1/vas/bvn-account-match"
		var res ValidateBvnResponse
		if err := c.send(ctx, c, http.MethodPost, url, req, &res); err != nil {
			valResult.Err = err
			return valResult
		}
		valResult.Data = &res

	default:
		valResult.Err = fmt.Errorf("unknown validation mode: %s", mode)
	}

	return valResult
}
