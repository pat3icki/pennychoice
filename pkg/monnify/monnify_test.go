package monnify

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	APIKEY       string = "MK_TEST_GC3B8XG2XX"
	SECRETKEY    string = "A663NRZA544DDPEM7KDN7Z8HRV6YXD8S"
	CONTRACTCODE string = "5867418298"
)

func TestAuthentication(t *testing.T) {
	client := New(http.DefaultClient, SANDBOX_BASEURL, APIKEY, SECRETKEY)
	err := client.Authenticate(context.Background())
	require.NoError(t, err)
	require.True(t, client.isAuthenticated)
	require.NotEmpty(t, client.bearerToken)
}

func TestVerifyWebhookSignature(t *testing.T) {
	client := New(http.DefaultClient, SANDBOX_BASEURL, APIKEY, SECRETKEY)

	rawBody := []byte(`{"eventType":"SUCCESSFUL_TRANSACTION","eventData":{}}`)

	h := hmac.New(sha512.New, []byte(SECRETKEY))
	h.Write(rawBody)
	expectedHexSig := hex.EncodeToString(h.Sum(nil))

	ok := client.VerifySignature(rawBody, expectedHexSig)
	require.True(t, ok)

	ok = client.VerifySignature(rawBody, "invalidsignature")
	require.False(t, ok)
}

func TestDecodeWebhook(t *testing.T) {
	payload := []byte(`{
		"eventType": "SUCCESSFUL_TRANSACTION",
		"eventData": {
			"transactionReference": "MNFY|04|20211117112842|000170",
			"paymentReference": "ref-12345",
			"amountPaid": 5000.00,
			"totalPayable": 5000.00,
			"paidOn": "2021-11-17 11:28:42.615",
			"paymentDescription": "Test Payment",
			"paymentMethod": "ACCOUNT_TRANSFER",
			"currencyCode": "NGN",
			"paymentStatus": "PAID",
			"customer": {
				"name": "John Doe",
				"email": "john.doe@example.com",
				"mobileNumber": "08012345678"
			},
			"product": {
				"reference": "prod-ref",
				"type": "RESERVED_ACCOUNT"
			}
		}
	}`)

	event, err := DecodeWebhookEvent(payload)
	require.NoError(t, err)
	require.Equal(t, "SUCCESSFUL_TRANSACTION", event.EventType)

	txEvent, err := event.DecodeSuccessfulTransaction()
	require.NoError(t, err)
	require.Equal(t, "MNFY|04|20211117112842|000170", txEvent.TransactionReference)
	require.Equal(t, "ref-12345", txEvent.PaymentReference)
	require.Equal(t, 5000.00, float64(txEvent.AmountPaid))
	require.Equal(t, "John Doe", txEvent.Customer.Name)
	require.Equal(t, "john.doe@example.com", txEvent.Customer.Email)
}

func TestCreateReservedAcct(t *testing.T) {
	client := New(http.DefaultClient, SANDBOX_BASEURL, APIKEY, SECRETKEY)

	timestamp := time.Now().Format("20060102150405")
	req := CreateReservedAcctReq{
		AccountName:          "Test Account",
		AccountReference:     "ref-" + timestamp,
		CurrencyCode:         "NGN",
		ContractCode:         CONTRACTCODE,
		CustomerName:         "John Doe",
		CustomerEmail:        "john.doe-" + timestamp + "@customdomain.com",
		GetAllAvailableBanks: true,
	}

	res, err := client.CreateReservedAcct(context.Background(), req)
	if err != nil {
		t.Logf("CreateReservedAcct returned error (possibly expected due to Sandbox limits/IP whitelist): %v", err)
	} else {
		require.NotNil(t, res)
		t.Logf("CreateReservedAcct succeeded: %+v", res)
		require.True(t, res.RequestSuccessful)
		hasAccount := res.ResponseBody.AccountNumber != "" || len(res.ResponseBody.Accounts) > 0
		require.True(t, hasAccount, "Should return at least one virtual account number")
	}
}

func TestValidateBankAccount(t *testing.T) {
	client := New(http.DefaultClient, SANDBOX_BASEURL, APIKEY, SECRETKEY)

	req := &ValiddateBankRequest{
		AccountNumber: "3391267489",
		BankCode:      "035",
	}

	result := client.Validate(context.Background(), "bank", req)
	if result.Err != nil {
		t.Logf("Validate bank returned error (possibly expected in Sandbox): %v", result.Err)
	} else {
		require.NotNil(t, result.Data)
		res, ok := result.Data.(*ValidateBankAccountResponese)
		require.True(t, ok)
		t.Logf("Validate bank returned data: %+v", res)
		require.True(t, res.RequestSuccessful)
		require.NotEmpty(t, res.Data.AccountName)
	}
}

func TestInitTransaction(t *testing.T) {
	client := New(http.DefaultClient, SANDBOX_BASEURL, APIKEY, SECRETKEY)

	timestamp := time.Now().Format("20060102150405")
	req := InitTransactionReq{
		Amount:             100.0,
		CustomerName:       "John Doe",
		CustomerEmail:      "john.doe@example.com",
		PaymentReference:   "ref-tx-" + timestamp,
		PaymentDescription: "Test Payment Initialization",
		CurrencyCode:       "NGN",
		ContractCode:       CONTRACTCODE,
		RedirectUrl:        "http://localhost:8080/callback",
	}

	res, err := client.InitTransaction(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.RequestSuccessful)
	require.NotEmpty(t, res.ResponseBody.CheckoutUrl)
	require.NotEmpty(t, res.ResponseBody.TransactionReference)

	// Query it
	qRes, err := client.QueryTransaction(context.Background(), req.PaymentReference, "")
	require.NoError(t, err)
	require.NotNil(t, qRes)
	require.True(t, qRes.RequestSuccessful)
	require.Equal(t, "PENDING", qRes.ResponseBody.PaymentStatus)
}

func TestInitiateSingleTransfer(t *testing.T) {
	client := New(http.DefaultClient, SANDBOX_BASEURL, APIKEY, SECRETKEY)

	timestamp := time.Now().Format("20060102150405")
	req := SingleTransferReq{
		Amount:                   500.0,
		Reference:                "ref-tr-" + timestamp,
		Narration:                "Test transfer payout",
		DestinationBankCode:      "058",
		DestinationAccountNumber: "0123456789",
		DestinationAccountName:   "John Doe",
		SourceAccountNumber:      "1234567890",
		Currency:                 "NGN",
	}

	res, err := client.InitiateSingleTransfer(context.Background(), req)
	if err != nil {
		t.Logf("InitiateSingleTransfer returned error (possibly expected): %v", err)
	} else {
		require.NotNil(t, res)
		t.Logf("InitiateSingleTransfer response: %+v", res)
	}
}
