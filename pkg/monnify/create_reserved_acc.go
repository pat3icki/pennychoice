package monnify

import (
	"context"
	"net/http"
)

type (
	// CreateReservedAcctReq request object
	CreateReservedAcctReq struct {
		AccountName          string                     `json:"accountName"`
		AccountReference     string                     `json:"accountReference"`
		CurrencyCode         string                     `json:"currencyCode"`
		ContractCode         string                     `json:"contractCode"`
		CustomerName         string                     `json:"customerName"`
		CustomerEmail        string                     `json:"customerEmail"`
		GetAllAvailableBanks bool                       `json:"getAllAvailableBanks"`
		IncomeSplitConfig    []IncomeSplitConfigReqBody `json:"incomeSplitConfig,omitempty"`
	}

	// IncomeSplitConfigReqBody request object
	IncomeSplitConfigReqBody struct {
		SubAccountCode  string `json:"subAccountCode"`
		SplitPercentage int    `json:"splitPercentage"`
		FeePercentage   int    `json:"feePercentage"`
		FeeBearer       bool   `json:"feeBearer"`
	}

	// ReservedAccountBank holds details of a specific reserved account at a bank
	ReservedAccountBank struct {
		BankCode      string `json:"bankCode"`
		BankName      string `json:"bankName"`
		AccountNumber string `json:"accountNumber"`
		AccountName   string `json:"accountName"`
	}

	// CreateReservedAcctRes response object
	CreateReservedAcctRes struct {
		RequestSuccessful bool   `json:"requestSuccessful"`
		ResponseMessage   string `json:"responseMessage"`
		ResponseCode      string `json:"responseCode"`
		ResponseBody      struct {
			ContractCode      string                `json:"contractCode"`
			AccountReference  string                `json:"accountReference"`
			AccountName       string                `json:"accountName"`
			CurrencyCode      string                `json:"currencyCode"`
			CustomerEmail     string                `json:"customerEmail"`
			CustomerName      string                `json:"customerName"`
			AccountNumber     string                `json:"accountNumber"`
			BankName          string                `json:"bankName"`
			BankCode          string                `json:"bankCode"`
			Accounts          []ReservedAccountBank `json:"accounts"`
			Status            string                `json:"status"`
			CreatedOn         string                `json:"createdOn"`
			IncomeSplitConfig interface{}           `json:"incomeSplitConfig"`
		} `json:"responseBody"`
	}
)

// CreateReservedAcct creates a reserved account with Monnify
func (c *Client) CreateReservedAcct(ctx context.Context, req CreateReservedAcctReq) (*CreateReservedAcctRes, error) {
	url := c.baseUrl + "/v2/bank-transfer/reserved-accounts"

	var res CreateReservedAcctRes
	if err := c.send(ctx, c, http.MethodPost, url, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
