package x402

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"

	"rationalgo/internal/services/algorand"
)

const (
	minTxnFee        = 1000
	x402PaymentNote  = "x402-payment-v2"
	x402FeePayerNote = "x402-fee-payer"
)

// buildAlgorandPaymentSignature constructs a gasless x402 v2 PAYMENT-SIGNATURE payload
// per GoPlausible scheme_exact_algo (atomic group + facilitator fee payer).
func buildAlgorandPaymentSignature(
	ctx context.Context,
	client *algorand.Client,
	required *paymentRequired,
	accept paymentAccept,
) (string, error) {
	assetID, err := strconv.ParseUint(accept.Asset, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse asset id %q: %w", accept.Asset, err)
	}
	amount, err := strconv.ParseUint(accept.Amount, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse amount %q: %w", accept.Amount, err)
	}

	if err := client.EnsureAssetOptIn(ctx, assetID); err != nil {
		return "", fmt.Errorf("asa opt-in: %w", err)
	}

	params, err := client.SuggestedParams(ctx)
	if err != nil {
		return "", err
	}

	feePayer := feePayerFromExtra(accept.Extra)
	var txns []types.Transaction
	paymentIndex := 0

	if feePayer != "" {
		feeParams := params
		feeParams.FlatFee = true
		feeParams.Fee = minTxnFee * 2

		feeTxn, err := transaction.MakePaymentTxn(
			feePayer, feePayer, 0,
			[]byte(x402FeePayerNote), "",
			feeParams,
		)
		if err != nil {
			return "", fmt.Errorf("make fee payer txn: %w", err)
		}

		axParams := params
		axParams.FlatFee = true
		axParams.Fee = 0

		axfer, err := transaction.MakeAssetTransferTxn(
			client.Address(), accept.PayTo, amount,
			[]byte(x402PaymentNote), axParams, "", assetID,
		)
		if err != nil {
			return "", fmt.Errorf("make axfer txn: %w", err)
		}

		txns = []types.Transaction{feeTxn, axfer}
		paymentIndex = 1
	} else {
		axParams := params
		axParams.FlatFee = true
		axParams.Fee = minTxnFee

		axfer, err := transaction.MakeAssetTransferTxn(
			client.Address(), accept.PayTo, amount,
			[]byte(x402PaymentNote), axParams, "", assetID,
		)
		if err != nil {
			return "", fmt.Errorf("make axfer txn: %w", err)
		}
		txns = []types.Transaction{axfer}
	}

	gid, err := crypto.ComputeGroupID(txns)
	if err != nil {
		return "", fmt.Errorf("compute group id: %w", err)
	}
	for i := range txns {
		txns[i].Group = gid
	}

	group := make([]string, len(txns))
	for i, txn := range txns {
		if feePayer != "" && i == 0 {
			group[i] = base64.StdEncoding.EncodeToString(
				msgpack.Encode(types.SignedTxn{Txn: txn}),
			)
			continue
		}
		stxn, err := client.SignTxn(txn)
		if err != nil {
			return "", fmt.Errorf("sign txn %d: %w", i, err)
		}
		group[i] = base64.StdEncoding.EncodeToString(stxn)
	}

	sig := paymentSignature{
		X402Version: 2,
		Scheme:      accept.Scheme,
		Network:     accept.Network,
		Resource:    required.Resource,
		Accepted:    accept,
		Extensions:  required.Extensions,
		Payload: avmPayload{
			PaymentGroup: group,
			PaymentIndex: paymentIndex,
		},
	}

	raw, err := json.Marshal(sig)
	if err != nil {
		return "", fmt.Errorf("marshal payment signature: %w", err)
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}
