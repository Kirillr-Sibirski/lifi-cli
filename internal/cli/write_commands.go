package cli

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/earn"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/evm"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type allowanceCommand struct{}
type approveCommand struct{}
type depositCommand struct{}

func newAllowanceCommand() Command { return allowanceCommand{} }

func (allowanceCommand) Name() string { return "allowance" }

func (allowanceCommand) Summary() string {
	return "Check token allowance for a wallet and spender"
}

func (allowanceCommand) Usage() string {
	return "lifi allowance [--chain <chain>] [--token <symbol-or-address>] [--owner <address>] [--spender <address>] [--amount <human>] [--quote-file <path>] [--json]"
}

func (allowanceCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("allowance")
	var chainArg, tokenArg, owner, spender, amount, quoteFile string
	fs.StringVar(&chainArg, "chain", "", "Chain name or ID")
	fs.StringVar(&tokenArg, "token", "", "Token symbol or address")
	fs.StringVar(&owner, "owner", "", "Owner address")
	fs.StringVar(&spender, "spender", "", "Spender address")
	fs.StringVar(&amount, "amount", "", "Required amount")
	fs.StringVar(&quoteFile, "quote-file", "", "Quote file path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	var chainID int
	var tokenAddress string
	var tokenDecimals int
	required := new(big.Int)

	if quoteFile != "" {
		quote, err := readQuoteFile(quoteFile)
		if err != nil {
			return err
		}
		chainID = quote.Action.FromChainID
		tokenAddress = quote.Action.FromToken.Address
		tokenDecimals = quote.Action.FromToken.Decimals
		owner = firstNonEmpty(owner, quote.Action.FromAddress)
		spender = firstNonEmpty(spender, quote.Estimate.ApprovalAddress)
		required.SetString(quote.Action.FromAmount, 10)
	} else {
		chain, err := rt.resolveChain(ctx, chainArg)
		if err != nil {
			return err
		}
		token, err := rt.resolveToken(ctx, chain, tokenArg)
		if err != nil {
			return err
		}
		chainID = chain.ID
		tokenAddress = token.Address
		tokenDecimals = token.Decimals
		if amount != "" {
			parsed, err := parseAmountToBaseUnits(amount, token.Decimals)
			if err != nil {
				return err
			}
			required = parsed
		}
	}

	if owner == "" || spender == "" {
		return fmt.Errorf("owner and spender are required")
	}

	chain, err := rt.resolveChain(ctx, fmt.Sprintf("%d", chainID))
	if err != nil {
		return err
	}
	rpcURL, err := rt.rpcURL(chain)
	if err != nil {
		return err
	}

	client, err := evm.DialRPC(ctx, rpcURL)
	if err != nil {
		return err
	}
	defer client.Close()

	allowance, err := evm.Allowance(ctx, client, tokenAddress, commonAddress(owner), commonAddress(spender))
	if err != nil {
		return err
	}

	sufficient := allowance.Cmp(required) >= 0
	payload := map[string]any{
		"chain_id":            chainID,
		"token":               tokenAddress,
		"owner":               owner,
		"spender":             spender,
		"allowance":           allowance.String(),
		"allowance_formatted": formatAmount(allowance.String(), tokenDecimals, 6),
		"required":            required.String(),
		"required_formatted":  formatAmount(required.String(), tokenDecimals, 6),
		"sufficient":          sufficient,
	}
	if cfg.Global.JSON {
		return writeJSON(payload)
	}

	printTable([]string{"field", "value"}, [][]string{
		{"chain", fmt.Sprintf("%s (%d)", chain.Name, chainID)},
		{"token", tokenAddress},
		{"owner", owner},
		{"spender", spender},
		{"allowance", formatAmount(allowance.String(), tokenDecimals, 6)},
		{"required", formatAmount(required.String(), tokenDecimals, 6)},
		{"sufficient", boolText(sufficient)},
	})
	return nil
}

func newApproveCommand() Command { return approveCommand{} }

func (approveCommand) Name() string { return "approve" }

func (approveCommand) Summary() string { return "Send an ERC-20 approval transaction" }

func (approveCommand) Usage() string {
	return "lifi approve --chain <chain> --token <symbol-or-address> --spender <address> --amount <human|max> [--yes] [--json]"
}

func (approveCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("approve")
	var chainArg, tokenArg, spender, amount string
	var yes bool
	fs.StringVar(&chainArg, "chain", "", "Chain name or ID")
	fs.StringVar(&tokenArg, "token", "", "Token symbol or address")
	fs.StringVar(&spender, "spender", "", "Spender address")
	fs.StringVar(&amount, "amount", "", "Approval amount or max")
	fs.BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	chain, err := rt.resolveChain(ctx, chainArg)
	if err != nil {
		return err
	}
	token, err := rt.resolveToken(ctx, chain, tokenArg)
	if err != nil {
		return err
	}
	if evm.IsNativeToken(token.Address) {
		return fmt.Errorf("native tokens do not require approval")
	}
	if spender == "" {
		return fmt.Errorf("--spender is required")
	}

	wallet, err := rt.wallet()
	if err != nil {
		return err
	}
	rpcURL, err := rt.rpcURL(chain)
	if err != nil {
		return err
	}
	client, err := evm.DialRPC(ctx, rpcURL)
	if err != nil {
		return err
	}
	defer client.Close()

	approvalAmount := evm.MaxApprovalAmount()
	if strings.ToLower(strings.TrimSpace(amount)) != "max" {
		approvalAmount, err = parseAmountToBaseUnits(amount, token.Decimals)
		if err != nil {
			return err
		}
	}

	if !yes {
		confirmed, err := promptConfirm(fmt.Sprintf("Approve %s %s for spender %s on %s?", formatAmount(approvalAmount.String(), token.Decimals, 6), token.Symbol, spender, chain.Name))
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	hash, err := evm.Approve(ctx, client, wallet, big.NewInt(int64(chain.ID)), token.Address, commonAddress(spender), approvalAmount)
	if err != nil {
		return err
	}
	receipt, err := evm.WaitForReceipt(ctx, client, hash, 3*time.Second)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"tx_hash": hash.Hex(),
		"status":  receipt.Status,
		"spender": spender,
		"amount":  approvalAmount.String(),
	}
	if cfg.Global.JSON {
		return writeJSON(payload)
	}

	printTable([]string{"field", "value"}, [][]string{
		{"tx hash", hash.Hex()},
		{"status", fmt.Sprintf("%d", receipt.Status)},
		{"spender", spender},
		{"amount", formatAmount(approvalAmount.String(), token.Decimals, 6)},
	})
	return nil
}

func newDepositCommand() Command { return depositCommand{} }

func (depositCommand) Name() string { return "deposit" }

func (depositCommand) Summary() string { return "Execute a full Earn deposit flow" }

func (depositCommand) Usage() string {
	return "lifi deposit --vault <address> --from-chain <chain> --from-token <symbol-or-address> --amount <human> [options]"
}

func (depositCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("deposit")
	var vaultArg, fromChainArg, toChainArg, fromTokenArg, amount, fromAddress, toAddress string
	var slippageBps, approveMode string
	var wait, verifyPosition, yes, dryRun bool
	fs.StringVar(&vaultArg, "vault", "", "Target vault address")
	fs.StringVar(&fromChainArg, "from-chain", "", "Source chain")
	fs.StringVar(&toChainArg, "to-chain", "", "Destination chain")
	fs.StringVar(&fromTokenArg, "from-token", "", "Source token")
	fs.StringVar(&amount, "amount", "", "Human-readable amount")
	fs.StringVar(&fromAddress, "from-address", "", "Source wallet address")
	fs.StringVar(&toAddress, "to-address", "", "Destination wallet address")
	fs.StringVar(&slippageBps, "slippage-bps", "", "Allowed slippage in basis points")
	fs.StringVar(&approveMode, "approve", "auto", "Approval mode: auto, always, or never")
	fs.BoolVar(&wait, "wait", false, "Wait for confirmation")
	fs.BoolVar(&verifyPosition, "verify-position", false, "Verify portfolio position after execution")
	fs.BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	fs.BoolVar(&dryRun, "dry-run", false, "Only prepare the quote and checks")
	if err := fs.Parse(args); err != nil {
		return err
	}

	quote, fromChain, fromToken, walletAddress, err := prepareQuote(cfg, quoteInputs{
		vaultArg:     vaultArg,
		fromChainArg: fromChainArg,
		toChainArg:   toChainArg,
		fromTokenArg: fromTokenArg,
		amount:       amount,
		fromAddress:  fromAddress,
		toAddress:    toAddress,
		slippageBps:  slippageBps,
	})
	if err != nil {
		return err
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	var baselinePositions []map[string]any
	if verifyPosition {
		verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 20*time.Second)
		portfolio, err := rt.earnClient.GetPortfolio(verifyCtx, walletAddress)
		verifyCancel()
		if err == nil {
			baselinePositions = portfolio.Positions
		}
	}

	rpcURL, err := rt.rpcURL(fromChain)
	if err != nil {
		return err
	}
	client, err := evm.DialRPC(ctx, rpcURL)
	if err != nil {
		return err
	}
	defer client.Close()

	balance, err := evm.Balance(ctx, client, fromToken.Address, commonAddress(walletAddress))
	if err != nil {
		return err
	}
	required := new(big.Int)
	required.SetString(quote.Action.FromAmount, 10)
	if balance.Cmp(required) < 0 {
		return fmt.Errorf(
			"insufficient balance: have %s %s, need %s %s",
			formatAmount(balance.String(), fromToken.Decimals, 6), fromToken.Symbol,
			formatAmount(required.String(), fromToken.Decimals, 6), fromToken.Symbol,
		)
	}

	approvalNeeded := !evm.IsNativeToken(fromToken.Address)
	if approvalNeeded {
		allowance, err := evm.Allowance(ctx, client, fromToken.Address, commonAddress(walletAddress), commonAddress(quote.Estimate.ApprovalAddress))
		if err != nil {
			return err
		}
		approvalNeeded = allowance.Cmp(required) < 0
	}

	mode := strings.ToLower(strings.TrimSpace(approveMode))
	switch mode {
	case "auto", "always", "never":
	default:
		return fmt.Errorf("--approve must be one of auto, always, or never")
	}
	if mode == "always" && !evm.IsNativeToken(fromToken.Address) {
		approvalNeeded = true
	}
	if mode == "never" && approvalNeeded && !dryRun {
		return fmt.Errorf("approval is required but --approve=never was set")
	}

	result := map[string]any{
		"quote":             quote,
		"wallet":            walletAddress,
		"balance":           balance.String(),
		"balance_formatted": formatAmount(balance.String(), fromToken.Decimals, 6),
		"approval_needed":   approvalNeeded,
		"dry_run":           dryRun,
	}

	if !cfg.Global.JSON {
		printTable([]string{"field", "value"}, quoteSummaryRows(quote))
		fmt.Println()
		printTable([]string{"field", "value"}, [][]string{
			{"wallet", walletAddress},
			{"balance", formatAmount(balance.String(), fromToken.Decimals, 6)},
			{"approval needed", boolText(approvalNeeded)},
			{"dry run", boolText(dryRun)},
		})
	}
	if dryRun {
		if cfg.Global.JSON {
			return writeJSON(result)
		}
		return nil
	}

	wallet, err := rt.wallet()
	if err != nil {
		return err
	}

	if !yes {
		confirmed, err := promptConfirm("Broadcast deposit transaction?")
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	if approvalNeeded && strings.ToLower(strings.TrimSpace(approveMode)) != "never" {
		hash, err := evm.Approve(ctx, client, wallet, big.NewInt(int64(fromChain.ID)), fromToken.Address, commonAddress(quote.Estimate.ApprovalAddress), required)
		if err != nil {
			return err
		}
		result["approval_tx_hash"] = hash.Hex()
		waitCtx, waitCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		receipt, err := evm.WaitForReceipt(waitCtx, client, hash, 3*time.Second)
		waitCancel()
		if err != nil {
			return err
		}
		result["approval_receipt_status"] = receipt.Status
	}

	hash, err := evm.SendQuoteTransaction(ctx, client, wallet, quote.TransactionRequest)
	if err != nil {
		return err
	}
	result["tx_hash"] = hash.Hex()
	if !cfg.Global.JSON {
		fmt.Printf("deposit transaction sent: %s\n", hash.Hex())
	}
	if !wait && !verifyPosition {
		if cfg.Global.JSON {
			return writeJSON(result)
		}
		return nil
	}

	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	receipt, err := evm.WaitForReceipt(waitCtx, client, hash, 3*time.Second)
	waitCancel()
	if err != nil {
		return err
	}
	result["receipt_status"] = receipt.Status
	if !cfg.Global.JSON {
		fmt.Printf("source receipt status: %d\n", receipt.Status)
	}

	if quote.Action.FromChainID != quote.Action.ToChainID {
		statusPayload, err := waitForExecutionStatus(rt, lifiapi.StatusRequest{
			TxHash:    hash.Hex(),
			Bridge:    quote.Tool,
			FromChain: quote.Action.FromChainID,
			ToChain:   quote.Action.ToChainID,
		}, 8*time.Second, 5*time.Minute)
		if err == nil {
			result["status"] = statusPayload
			if !cfg.Global.JSON {
				fmt.Println("cross-chain status")
				blob, _ := prettyJSON(statusPayload)
				fmt.Println(blob)
			}
		}
	}

	if verifyPosition {
		vault, err := rt.resolveVault(ctx, vaultArg)
		if err != nil {
			return err
		}
		expectedDelta := parseFloat(formatAmount(quote.Estimate.ToAmount, quote.Action.ToToken.Decimals, 9))
		found, positions, err := waitForPortfolioPosition(rt, walletAddress, vault, baselinePositions, expectedDelta, 5*time.Second, time.Minute)
		if err != nil {
			return err
		}
		result["position_detected"] = found
		result["positions"] = positions
		if !cfg.Global.JSON {
			fmt.Printf("position detected: %s\n", boolText(found))
		}
	}
	if cfg.Global.JSON {
		return writeJSON(result)
	}
	return nil
}

func waitForExecutionStatus(rt *runtime, request lifiapi.StatusRequest, interval, timeout time.Duration) (map[string]any, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	deadline := time.Now().Add(timeout)
	var last map[string]any
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		payload, err := rt.lifiClient.GetStatus(ctx, request)
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

func waitForPortfolioPosition(rt *runtime, walletAddress string, vault *earn.Vault, baseline []map[string]any, expectedDelta float64, interval, timeout time.Duration) (bool, []map[string]any, error) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if timeout <= 0 {
		timeout = time.Minute
	}

	deadline := time.Now().Add(timeout)
	var last []map[string]any
	baselineTotal := portfolioBalanceForVaultAsset(baseline, *vault)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		portfolio, err := rt.earnClient.GetPortfolio(ctx, walletAddress)
		cancel()
		if err == nil {
			last = portfolio.Positions
			if portfolioShowsDeposit(portfolio.Positions, *vault, baselineTotal, expectedDelta) {
				return true, portfolio.Positions, nil
			}
		}

		if time.Now().After(deadline) {
			if last != nil {
				return portfolioShowsDeposit(last, *vault, baselineTotal, expectedDelta), last, nil
			}
			if err != nil {
				return false, nil, err
			}
			return false, nil, fmt.Errorf("timed out waiting for portfolio update")
		}
		time.Sleep(interval)
	}
}

func portfolioShowsDeposit(positions []map[string]any, vault earn.Vault, baselineTotal, expectedDelta float64) bool {
	if findVaultInPositions(positions, vault) {
		return true
	}

	currentTotal := portfolioBalanceForVaultAsset(positions, vault)
	delta := currentTotal - baselineTotal
	if delta <= 0 {
		return false
	}
	if expectedDelta <= 0 {
		return true
	}
	return delta >= expectedDelta*0.5
}
