package flow

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/apperror"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/earn"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/evm"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type ChainExecutor interface {
	Close()
	TokenBalance(ctx context.Context, tokenAddress string, owner common.Address) (*big.Int, error)
	NativeBalance(ctx context.Context, owner common.Address) (*big.Int, error)
	Allowance(ctx context.Context, tokenAddress string, owner, spender common.Address) (*big.Int, error)
	SimulateQuote(ctx context.Context, request lifiapi.TransactionRequest, from common.Address) (uint64, error)
	EstimateQuoteFee(ctx context.Context, request lifiapi.TransactionRequest, gasPolicy string) (*evm.FeeEstimate, error)
	EstimateApprovalFee(ctx context.Context, wallet common.Address, chainID int, tokenAddress string, spender common.Address, amount *big.Int, gasPolicy string) (*evm.FeeEstimate, error)
	Approve(ctx context.Context, wallet *evm.Wallet, chainID int, tokenAddress string, spender common.Address, amount *big.Int, gasPolicy string) (common.Hash, error)
	SendQuote(ctx context.Context, wallet *evm.Wallet, request lifiapi.TransactionRequest, gasPolicy string) (common.Hash, error)
	WaitForReceipt(ctx context.Context, txHash common.Hash, interval time.Duration) (*types.Receipt, error)
}

type PortfolioClient interface {
	GetPortfolio(ctx context.Context, address string) (*earn.PortfolioResponse, error)
}

type StatusClient interface {
	GetStatus(ctx context.Context, request lifiapi.StatusRequest) (map[string]any, error)
}

type DepositRequest struct {
	Quote              *lifiapi.Quote
	Vault              *earn.Vault
	FromChain          *lifiapi.Chain
	FromToken          *lifiapi.Token
	WalletAddress      string
	Wallet             *evm.Wallet
	Executor           ChainExecutor
	PortfolioClient    PortfolioClient
	StatusClient       StatusClient
	DryRun             bool
	Simulate           bool
	Wait               bool
	VerifyPosition     bool
	ApproveMode        string
	ApprovalAmountMode string
	GasPolicy          string
	WaitTimeout        time.Duration
	PortfolioTimeout   time.Duration
	StatusTimeout      time.Duration
	PollInterval       time.Duration
	VerificationDelta  float64
}

type Preflight struct {
	WalletAddress          string `json:"wallet_address"`
	SourceChain            string `json:"source_chain"`
	SourceToken            string `json:"source_token"`
	SourceAmount           string `json:"source_amount"`
	DestinationVault       string `json:"destination_vault"`
	ExpectedReceived       string `json:"expected_received"`
	ApprovalAddress        string `json:"approval_address"`
	ApprovalNeeded         bool   `json:"approval_needed"`
	ApprovalAmount         string `json:"approval_amount"`
	GasPolicy              string `json:"gas_policy"`
	NativeBalance          string `json:"native_balance"`
	NativeBalanceFormatted string `json:"native_balance_formatted"`
	TokenBalance           string `json:"token_balance"`
	TokenBalanceFormatted  string `json:"token_balance_formatted"`
	EstimatedGasWei        string `json:"estimated_gas_wei"`
	EstimatedGasFormatted  string `json:"estimated_gas_formatted"`
	QuoteGasLimit          uint64 `json:"quote_gas_limit"`
	SimulatedGasLimit      uint64 `json:"simulated_gas_limit,omitempty"`
}

type Result struct {
	Stage                 string           `json:"stage"`
	Status                string           `json:"status"`
	Message               string           `json:"message,omitempty"`
	TxHash                string           `json:"tx_hash,omitempty"`
	ApprovalTxHash        string           `json:"approval_tx_hash,omitempty"`
	ReceiptStatus         uint64           `json:"receipt_status,omitempty"`
	ApprovalReceiptStatus uint64           `json:"approval_receipt_status,omitempty"`
	PositionDetected      bool             `json:"position_detected,omitempty"`
	Positions             []map[string]any `json:"positions,omitempty"`
	StatusPayload         map[string]any   `json:"status_payload,omitempty"`
	Preflight             *Preflight       `json:"preflight,omitempty"`
	Quote                 *lifiapi.Quote   `json:"quote,omitempty"`
}

func ExecuteDeposit(ctx context.Context, in DepositRequest) (*Result, error) {
	if err := validateRequest(in); err != nil {
		return nil, err
	}
	if in.WaitTimeout <= 0 {
		in.WaitTimeout = 5 * time.Minute
	}
	if in.PortfolioTimeout <= 0 {
		in.PortfolioTimeout = time.Minute
	}
	if in.StatusTimeout <= 0 {
		in.StatusTimeout = 5 * time.Minute
	}
	if in.PollInterval <= 0 {
		in.PollInterval = 5 * time.Second
	}

	result := &Result{
		Stage:  "preflight",
		Status: "ready",
		Quote:  in.Quote,
	}

	var baselinePositions []map[string]any
	if in.VerifyPosition && in.PortfolioClient != nil {
		verifyCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		portfolio, err := in.PortfolioClient.GetPortfolio(verifyCtx, in.WalletAddress)
		cancel()
		if err == nil {
			baselinePositions = portfolio.Positions
		}
	}

	required := new(big.Int)
	if _, ok := required.SetString(in.Quote.Action.FromAmount, 10); !ok {
		return nil, apperror.Formatf("execution", apperror.ExitExecution, "invalid quote amount %q", in.Quote.Action.FromAmount)
	}

	tokenBalance, err := in.Executor.TokenBalance(ctx, in.FromToken.Address, common.HexToAddress(in.WalletAddress))
	if err != nil {
		return nil, apperror.Wrap("rpc", apperror.ExitRPC, err)
	}
	if tokenBalance.Cmp(required) < 0 {
		return nil, apperror.Formatf(
			"execution",
			apperror.ExitExecution,
			"insufficient balance: have %s %s, need %s %s",
			formatAmount(tokenBalance.String(), in.FromToken.Decimals, 6), in.FromToken.Symbol,
			formatAmount(required.String(), in.FromToken.Decimals, 6), in.FromToken.Symbol,
		)
	}

	nativeBalance, err := in.Executor.NativeBalance(ctx, common.HexToAddress(in.WalletAddress))
	if err != nil {
		return nil, apperror.Wrap("rpc", apperror.ExitRPC, err)
	}
	if nativeBalance.Sign() == 0 {
		return nil, apperror.Formatf("execution", apperror.ExitExecution, "insufficient gas token balance on %s", in.FromChain.Name)
	}

	approvalNeeded := !evm.IsNativeToken(in.FromToken.Address)
	approvalAmount := new(big.Int).Set(required)
	if strings.EqualFold(in.ApprovalAmountMode, "infinite") {
		approvalAmount = evm.MaxApprovalAmount()
	}
	if approvalNeeded {
		if strings.TrimSpace(in.Quote.Estimate.ApprovalAddress) == "" {
			return nil, apperror.New("execution", apperror.ExitExecution, "quote is missing approval address")
		}
		allowance, err := in.Executor.Allowance(ctx, in.FromToken.Address, common.HexToAddress(in.WalletAddress), common.HexToAddress(in.Quote.Estimate.ApprovalAddress))
		if err != nil {
			return nil, apperror.Wrap("rpc", apperror.ExitRPC, err)
		}
		approvalNeeded = allowance.Cmp(required) < 0
	}

	approveMode := strings.ToLower(strings.TrimSpace(in.ApproveMode))
	switch approveMode {
	case "", "auto":
		approveMode = "auto"
	case "always":
		approvalNeeded = !evm.IsNativeToken(in.FromToken.Address)
	case "never":
		if approvalNeeded && !in.DryRun {
			return nil, apperror.New("execution", apperror.ExitExecution, "approval is required but --approve=never was set")
		}
	default:
		return nil, apperror.New("input", apperror.ExitInput, "--approve must be one of auto, always, or never")
	}

	var simulatedGasLimit uint64
	if in.Simulate {
		simulatedGasLimit, err = in.Executor.SimulateQuote(ctx, in.Quote.TransactionRequest, common.HexToAddress(in.WalletAddress))
		if err != nil {
			return nil, apperror.Wrap("execution", apperror.ExitExecution, fmt.Errorf("simulation failed: %w", err))
		}
	}

	quoteFee, err := in.Executor.EstimateQuoteFee(ctx, in.Quote.TransactionRequest, in.GasPolicy)
	if err != nil {
		return nil, apperror.Wrap("rpc", apperror.ExitRPC, err)
	}

	totalGasCost := new(big.Int).Set(quoteFee.EstimatedCost)
	if approvalNeeded {
		approvalFee, err := in.Executor.EstimateApprovalFee(ctx, common.HexToAddress(in.WalletAddress), in.FromChain.ID, in.FromToken.Address, common.HexToAddress(in.Quote.Estimate.ApprovalAddress), approvalAmount, in.GasPolicy)
		if err != nil {
			return nil, apperror.Wrap("rpc", apperror.ExitRPC, err)
		}
		totalGasCost.Add(totalGasCost, approvalFee.EstimatedCost)
	}
	if nativeBalance.Cmp(totalGasCost) < 0 {
		return nil, apperror.Formatf(
			"execution",
			apperror.ExitExecution,
			"insufficient gas token balance: have %s %s, need about %s %s",
			formatAmount(nativeBalance.String(), in.FromChain.NativeToken.Decimals, 6), in.FromChain.NativeToken.Symbol,
			formatAmount(totalGasCost.String(), in.FromChain.NativeToken.Decimals, 6), in.FromChain.NativeToken.Symbol,
		)
	}

	result.Preflight = &Preflight{
		WalletAddress:          in.WalletAddress,
		SourceChain:            fmt.Sprintf("%s (%d)", in.FromChain.Name, in.FromChain.ID),
		SourceToken:            in.FromToken.Symbol,
		SourceAmount:           formatAmount(required.String(), in.FromToken.Decimals, 6),
		DestinationVault:       in.Vault.Address,
		ExpectedReceived:       formatAmount(in.Quote.Estimate.ToAmount, in.Quote.Action.ToToken.Decimals, 6),
		ApprovalAddress:        in.Quote.Estimate.ApprovalAddress,
		ApprovalNeeded:         approvalNeeded,
		ApprovalAmount:         formatAmount(approvalAmount.String(), in.FromToken.Decimals, 6),
		GasPolicy:              strings.ToLower(strings.TrimSpace(in.GasPolicy)),
		NativeBalance:          nativeBalance.String(),
		NativeBalanceFormatted: formatAmount(nativeBalance.String(), in.FromChain.NativeToken.Decimals, 6),
		TokenBalance:           tokenBalance.String(),
		TokenBalanceFormatted:  formatAmount(tokenBalance.String(), in.FromToken.Decimals, 6),
		EstimatedGasWei:        totalGasCost.String(),
		EstimatedGasFormatted:  formatAmount(totalGasCost.String(), in.FromChain.NativeToken.Decimals, 6),
		QuoteGasLimit:          quoteFee.GasLimit,
		SimulatedGasLimit:      simulatedGasLimit,
	}
	if in.DryRun {
		result.Stage = "dry-run"
		result.Status = "ok"
		result.Message = "deposit preflight completed"
		return result, nil
	}
	if in.Wallet == nil {
		return nil, apperror.New("config", apperror.ExitConfig, "wallet private key is required for broadcast")
	}

	if approvalNeeded && approveMode != "never" {
		hash, err := in.Executor.Approve(ctx, in.Wallet, in.FromChain.ID, in.FromToken.Address, common.HexToAddress(in.Quote.Estimate.ApprovalAddress), approvalAmount, in.GasPolicy)
		if err != nil {
			return nil, apperror.Wrap("execution", apperror.ExitExecution, err)
		}
		result.ApprovalTxHash = hash.Hex()
		waitCtx, cancel := context.WithTimeout(ctx, in.WaitTimeout)
		approvalReceipt, err := in.Executor.WaitForReceipt(waitCtx, hash, 3*time.Second)
		cancel()
		if err != nil {
			return nil, apperror.Wrap("rpc", apperror.ExitRPC, err)
		}
		result.ApprovalReceiptStatus = approvalReceipt.Status
		if approvalReceipt.Status != types.ReceiptStatusSuccessful {
			return nil, apperror.New("execution", apperror.ExitExecution, "approval transaction failed")
		}
	}

	hash, err := in.Executor.SendQuote(ctx, in.Wallet, in.Quote.TransactionRequest, in.GasPolicy)
	if err != nil {
		return nil, apperror.Wrap("execution", apperror.ExitExecution, err)
	}
	result.TxHash = hash.Hex()
	result.Stage = "broadcast"
	result.Status = "submitted"
	if !in.Wait && !in.VerifyPosition {
		return result, nil
	}

	waitCtx, cancel := context.WithTimeout(ctx, in.WaitTimeout)
	receipt, err := in.Executor.WaitForReceipt(waitCtx, hash, 3*time.Second)
	cancel()
	if err != nil {
		return nil, apperror.Wrap("rpc", apperror.ExitRPC, err)
	}
	result.ReceiptStatus = receipt.Status
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, apperror.New("execution", apperror.ExitExecution, "deposit transaction failed")
	}
	result.Stage = "confirmed"
	result.Status = "ok"

	if in.Quote.Action.FromChainID != in.Quote.Action.ToChainID && in.StatusClient != nil {
		statusPayload, err := waitForExecutionStatus(ctx, in.StatusClient, lifiapi.StatusRequest{
			TxHash:    hash.Hex(),
			Bridge:    in.Quote.Tool,
			FromChain: in.Quote.Action.FromChainID,
			ToChain:   in.Quote.Action.ToChainID,
		}, in.PollInterval, in.StatusTimeout)
		if err == nil {
			result.StatusPayload = statusPayload
		}
	}

	if in.VerifyPosition && in.PortfolioClient != nil {
		found, positions, err := waitForPortfolioPosition(ctx, in.PortfolioClient, in.WalletAddress, in.Vault, baselinePositions, in.VerificationDelta, in.PollInterval, in.PortfolioTimeout)
		if err != nil {
			return nil, apperror.Wrap("verification", apperror.ExitVerification, err)
		}
		result.PositionDetected = found
		result.Positions = positions
		if !found {
			return nil, apperror.New("verification", apperror.ExitVerification, "position detected: no")
		}
	}

	return result, nil
}

func validateRequest(in DepositRequest) error {
	switch {
	case in.Quote == nil:
		return apperror.New("input", apperror.ExitInput, "quote is required")
	case in.Vault == nil:
		return apperror.New("input", apperror.ExitInput, "vault is required")
	case in.FromChain == nil:
		return apperror.New("input", apperror.ExitInput, "from chain is required")
	case in.FromToken == nil:
		return apperror.New("input", apperror.ExitInput, "from token is required")
	case strings.TrimSpace(in.WalletAddress) == "":
		return apperror.New("input", apperror.ExitInput, "wallet address is required")
	case in.Executor == nil:
		return apperror.New("input", apperror.ExitInput, "executor is required")
	case !in.Vault.IsTransactional:
		return apperror.New("execution", apperror.ExitExecution, "selected vault is not transactional")
	}
	return nil
}

func waitForExecutionStatus(ctx context.Context, client StatusClient, request lifiapi.StatusRequest, interval, timeout time.Duration) (map[string]any, error) {
	deadline := time.Now().Add(timeout)
	var last map[string]any
	for {
		pollCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		payload, err := client.GetStatus(pollCtx, request)
		cancel()
		if err == nil {
			last = payload
			if isTerminalStatus(payload) {
				return payload, nil
			}
		}
		if time.Now().After(deadline) {
			if last != nil {
				return last, nil
			}
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("timed out waiting for LI.FI execution status")
		}
		time.Sleep(interval)
	}
}

func waitForPortfolioPosition(ctx context.Context, portfolioClient PortfolioClient, walletAddress string, vault *earn.Vault, baseline []map[string]any, expectedDelta float64, interval, timeout time.Duration) (bool, []map[string]any, error) {
	deadline := time.Now().Add(timeout)
	var last []map[string]any
	baselineTotal := portfolioBalanceForVaultAsset(baseline, *vault)
	for {
		pollCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		portfolio, err := portfolioClient.GetPortfolio(pollCtx, walletAddress)
		cancel()
		if err == nil {
			last = portfolio.Positions
			if PortfolioShowsDeposit(portfolio.Positions, *vault, baselineTotal, expectedDelta) {
				return true, portfolio.Positions, nil
			}
		}

		if time.Now().After(deadline) {
			if last != nil {
				return PortfolioShowsDeposit(last, *vault, baselineTotal, expectedDelta), last, nil
			}
			if err != nil {
				return false, nil, err
			}
			return false, nil, fmt.Errorf("timed out waiting for portfolio update")
		}
		time.Sleep(interval)
	}
}
