package flow

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/earn"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/evm"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type fakeExecutor struct {
	tokenBalance  *big.Int
	nativeBalance *big.Int
	allowance     *big.Int
	simulatedGas  uint64
	simulateErr   error
	quoteFee      *evm.FeeEstimate
	approvalFee   *evm.FeeEstimate
	sendErr       error
	approveErr    error
}

func (f *fakeExecutor) Close() {}
func (f *fakeExecutor) TokenBalance(context.Context, string, common.Address) (*big.Int, error) {
	return new(big.Int).Set(f.tokenBalance), nil
}
func (f *fakeExecutor) NativeBalance(context.Context, common.Address) (*big.Int, error) {
	return new(big.Int).Set(f.nativeBalance), nil
}
func (f *fakeExecutor) Allowance(context.Context, string, common.Address, common.Address) (*big.Int, error) {
	return new(big.Int).Set(f.allowance), nil
}
func (f *fakeExecutor) SimulateQuote(context.Context, lifiapi.TransactionRequest, common.Address) (uint64, error) {
	if f.simulateErr != nil {
		return 0, f.simulateErr
	}
	return f.simulatedGas, nil
}
func (f *fakeExecutor) EstimateQuoteFee(context.Context, lifiapi.TransactionRequest, string) (*evm.FeeEstimate, error) {
	return f.quoteFee, nil
}
func (f *fakeExecutor) EstimateApprovalFee(context.Context, common.Address, int, string, common.Address, *big.Int, string) (*evm.FeeEstimate, error) {
	return f.approvalFee, nil
}
func (f *fakeExecutor) Approve(context.Context, *evm.Wallet, int, string, common.Address, *big.Int, string) (common.Hash, error) {
	if f.approveErr != nil {
		return common.Hash{}, f.approveErr
	}
	return common.HexToHash("0xabc"), nil
}
func (f *fakeExecutor) SendQuote(context.Context, *evm.Wallet, lifiapi.TransactionRequest, string) (common.Hash, error) {
	if f.sendErr != nil {
		return common.Hash{}, f.sendErr
	}
	return common.HexToHash("0xdef"), nil
}
func (f *fakeExecutor) WaitForReceipt(context.Context, common.Hash, time.Duration) (*types.Receipt, error) {
	return &types.Receipt{Status: types.ReceiptStatusSuccessful}, nil
}

type fakePortfolioClient struct {
	positions []map[string]any
}

func (f fakePortfolioClient) GetPortfolio(context.Context, string) (*earn.PortfolioResponse, error) {
	return &earn.PortfolioResponse{Positions: f.positions}, nil
}

func TestExecuteDepositDryRunShowsApprovalNeeded(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{
		tokenBalance:  big.NewInt(2_000_000),
		nativeBalance: big.NewInt(1_000_000_000_000_000),
		allowance:     big.NewInt(0),
		simulateErr:   errors.New("simulation should be skipped until approval exists"),
		quoteFee:      &evm.FeeEstimate{GasLimit: 200000, EstimatedCost: big.NewInt(1000)},
		approvalFee:   &evm.FeeEstimate{GasLimit: 65000, EstimatedCost: big.NewInt(1000)},
	}

	result, err := ExecuteDeposit(context.Background(), DepositRequest{
		Quote:              sampleQuote(),
		Vault:              sampleVault(),
		FromChain:          sampleChain(),
		FromToken:          sampleToken(),
		WalletAddress:      "0x1111111111111111111111111111111111111111",
		Executor:           executor,
		DryRun:             true,
		Simulate:           true,
		ApproveMode:        "auto",
		ApprovalAmountMode: "exact",
		GasPolicy:          "auto",
	})
	if err != nil {
		t.Fatalf("ExecuteDeposit returned error: %v", err)
	}
	if result.Preflight == nil || !result.Preflight.ApprovalNeeded {
		t.Fatalf("expected preflight approval_needed to be true")
	}
	if result.Preflight.SimulationStatus != "skipped" {
		t.Fatalf("expected simulation to be skipped, got %#v", result.Preflight)
	}
}

func TestExecuteDepositSimulatesWhenApprovalAlreadySatisfied(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{
		tokenBalance:  big.NewInt(2_000_000),
		nativeBalance: big.NewInt(1_000_000_000_000_000),
		allowance:     big.NewInt(2_000_000),
		simulatedGas:  210000,
		quoteFee:      &evm.FeeEstimate{GasLimit: 200000, EstimatedCost: big.NewInt(1000)},
		approvalFee:   &evm.FeeEstimate{GasLimit: 65000, EstimatedCost: big.NewInt(1000)},
	}

	result, err := ExecuteDeposit(context.Background(), DepositRequest{
		Quote:              sampleQuote(),
		Vault:              sampleVault(),
		FromChain:          sampleChain(),
		FromToken:          sampleToken(),
		WalletAddress:      "0x1111111111111111111111111111111111111111",
		Executor:           executor,
		DryRun:             true,
		Simulate:           true,
		ApproveMode:        "auto",
		ApprovalAmountMode: "exact",
		GasPolicy:          "auto",
	})
	if err != nil {
		t.Fatalf("ExecuteDeposit returned error: %v", err)
	}
	if result.Preflight == nil {
		t.Fatalf("expected preflight")
	}
	if result.Preflight.SimulationStatus != "ok" || result.Preflight.SimulatedGasLimit != 210000 {
		t.Fatalf("expected simulation to run, got %#v", result.Preflight)
	}
}

func TestExecuteDepositReturnsSendFailure(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{
		tokenBalance:  big.NewInt(2_000_000),
		nativeBalance: big.NewInt(1_000_000_000_000_000),
		allowance:     big.NewInt(2_000_000),
		simulatedGas:  200000,
		quoteFee:      &evm.FeeEstimate{GasLimit: 200000, EstimatedCost: big.NewInt(1000)},
		approvalFee:   &evm.FeeEstimate{GasLimit: 65000, EstimatedCost: big.NewInt(1000)},
		sendErr:       errors.New("send failed"),
	}

	_, err := ExecuteDeposit(context.Background(), DepositRequest{
		Quote:              sampleQuote(),
		Vault:              sampleVault(),
		FromChain:          sampleChain(),
		FromToken:          sampleToken(),
		WalletAddress:      "0x1111111111111111111111111111111111111111",
		Wallet:             &evm.Wallet{Address: common.HexToAddress("0x1111111111111111111111111111111111111111")},
		Executor:           executor,
		Wait:               true,
		ApproveMode:        "auto",
		ApprovalAmountMode: "exact",
		GasPolicy:          "auto",
	})
	if err == nil {
		t.Fatalf("expected send failure")
	}
}

func TestExecuteDepositReturnsVerificationError(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{
		tokenBalance:  big.NewInt(2_000_000),
		nativeBalance: big.NewInt(1_000_000_000_000_000),
		allowance:     big.NewInt(2_000_000),
		simulatedGas:  200000,
		quoteFee:      &evm.FeeEstimate{GasLimit: 200000, EstimatedCost: big.NewInt(1000)},
		approvalFee:   &evm.FeeEstimate{GasLimit: 65000, EstimatedCost: big.NewInt(1000)},
	}

	_, err := ExecuteDeposit(context.Background(), DepositRequest{
		Quote:              sampleQuote(),
		Vault:              sampleVault(),
		FromChain:          sampleChain(),
		FromToken:          sampleToken(),
		WalletAddress:      "0x1111111111111111111111111111111111111111",
		Wallet:             &evm.Wallet{Address: common.HexToAddress("0x1111111111111111111111111111111111111111")},
		Executor:           executor,
		PortfolioClient:    fakePortfolioClient{},
		Wait:               true,
		VerifyPosition:     true,
		ApproveMode:        "auto",
		ApprovalAmountMode: "exact",
		GasPolicy:          "auto",
		PortfolioTimeout:   10 * time.Millisecond,
		PollInterval:       5 * time.Millisecond,
	})
	if err == nil {
		t.Fatalf("expected verification error")
	}
}

func sampleQuote() *lifiapi.Quote {
	return &lifiapi.Quote{
		Tool: "lifi",
		Action: lifiapi.QuoteAction{
			FromAmount: "1000000",
			FromToken: lifiapi.Token{
				Address:  "0xusdc",
				Symbol:   "USDC",
				Decimals: 6,
			},
			ToToken: lifiapi.Token{
				Address:  "0xvault",
				Symbol:   "aUSDC",
				Decimals: 6,
			},
			FromChainID: 10,
			ToChainID:   10,
		},
		Estimate: lifiapi.QuoteEstimate{
			ApprovalAddress: "0x2222222222222222222222222222222222222222",
			ToAmount:        "1000000",
		},
		TransactionRequest: lifiapi.TransactionRequest{
			To:       "0x3333333333333333333333333333333333333333",
			Value:    "0",
			Data:     "0x1234",
			ChainID:  10,
			GasPrice: "1000000000",
			GasLimit: "200000",
			From:     "0x1111111111111111111111111111111111111111",
		},
	}
}

func sampleVault() *earn.Vault {
	return &earn.Vault{
		Address:          "0xvault",
		ChainID:          10,
		IsTransactional:  true,
		Protocol:         earn.Protocol{Name: "morpho-v1"},
		UnderlyingTokens: []earn.UnderlyingToken{{Symbol: "USDC", Address: "0xusdc"}},
	}
}

func sampleChain() *lifiapi.Chain {
	return &lifiapi.Chain{
		ID:          10,
		Name:        "Optimism",
		NativeToken: lifiapi.Token{Symbol: "ETH", Decimals: 18},
	}
}

func sampleToken() *lifiapi.Token {
	return &lifiapi.Token{Address: "0xusdc", Symbol: "USDC", Decimals: 6}
}
