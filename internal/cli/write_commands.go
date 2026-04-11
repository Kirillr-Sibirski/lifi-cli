package cli

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/apperror"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/evm"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/flow"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type allowanceCommand struct{}
type approveCommand struct{}
type depositCommand struct{}
type withdrawCommand struct{}

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
	return "lifi approve --chain <chain> --token <symbol-or-address> --spender <address> --amount <human|max> [--gas-policy auto|rpc] [--yes] [--json]"
}

func (approveCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("approve")
	var chainArg, tokenArg, spender, amount, gasPolicy string
	var yes bool
	fs.StringVar(&chainArg, "chain", "", "Chain name or ID")
	fs.StringVar(&tokenArg, "token", "", "Token symbol or address")
	fs.StringVar(&spender, "spender", "", "Spender address")
	fs.StringVar(&amount, "amount", "", "Approval amount or max")
	fs.StringVar(&gasPolicy, "gas-policy", "auto", "Gas pricing policy: auto or rpc")
	fs.BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if tokenArg == "" {
		return fmt.Errorf("--token is required\n\nUsage: %s\nExample: lifi approve --chain base --token USDC --spender 0x1231... --amount 100", (approveCommand{}).Usage())
	}
	if spender == "" {
		return fmt.Errorf("--spender is required\n\nUsage: %s", (approveCommand{}).Usage())
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

	wallet, err := rt.wallet()
	if err != nil {
		return apperror.Wrap("config", apperror.ExitConfig, err)
	}
	rpcURL, err := rt.rpcURL(chain)
	if err != nil {
		return apperror.Wrap("rpc", apperror.ExitRPC, err)
	}
	client, err := evm.DialRPC(ctx, rpcURL)
	if err != nil {
		return apperror.Wrap("rpc", apperror.ExitRPC, err)
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

	hash, err := evm.Approve(ctx, client, wallet, chain.ID, token.Address, commonAddress(spender), approvalAmount, gasPolicy)
	if err != nil {
		return apperror.Wrap("execution", apperror.ExitExecution, err)
	}
	receipt, err := evm.WaitForReceipt(ctx, client, hash, 3*time.Second)
	if err != nil {
		return apperror.Wrap("rpc", apperror.ExitRPC, err)
	}

	payload := map[string]any{
		"stage":        "approval",
		"status":       "ok",
		"message":      "approval transaction confirmed",
		"tx_hash":      hash.Hex(),
		"receipt_code": receipt.Status,
		"spender":      spender,
		"amount":       approvalAmount.String(),
		"explorer_url": explorerTxURL(chain, hash.Hex()),
	}
	if cfg.Global.JSON {
		return writeJSON(payload)
	}

	printTable([]string{"field", "value"}, [][]string{
		{"tx hash", hash.Hex()},
		{"explorer", emptyFallback(explorerTxURL(chain, hash.Hex()))},
		{"status", fmt.Sprintf("%d", receipt.Status)},
		{"spender", spender},
		{"amount", formatAmount(approvalAmount.String(), token.Decimals, 6)},
		{"gas policy", gasPolicy},
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
	var slippageBps, approveMode, approvalAmountMode, gasPolicy, waitTimeoutArg, portfolioTimeoutArg string
	var wait, verifyPosition, yes, dryRun, simulate, skipSimulate bool
	fs.StringVar(&vaultArg, "vault", "", "Target vault address")
	fs.StringVar(&fromChainArg, "from-chain", "", "Source chain")
	fs.StringVar(&toChainArg, "to-chain", "", "Destination chain")
	fs.StringVar(&fromTokenArg, "from-token", "", "Source token")
	fs.StringVar(&amount, "amount", "", "Human-readable amount")
	fs.StringVar(&fromAddress, "from-address", "", "Source wallet address")
	fs.StringVar(&toAddress, "to-address", "", "Destination wallet address")
	fs.StringVar(&slippageBps, "slippage-bps", "", "Allowed slippage in basis points")
	fs.StringVar(&approveMode, "approve", "auto", "Approval mode: auto, always, or never")
	fs.StringVar(&approvalAmountMode, "approval-amount", "exact", "Approval amount: exact or infinite")
	fs.StringVar(&gasPolicy, "gas-policy", "auto", "Gas pricing policy: auto, quote, or rpc")
	fs.StringVar(&waitTimeoutArg, "wait-timeout", "5m", "Maximum time to wait for transaction confirmation")
	fs.StringVar(&portfolioTimeoutArg, "portfolio-timeout", "1m", "Maximum time to wait for portfolio verification")
	fs.BoolVar(&wait, "wait", false, "Wait for confirmation")
	fs.BoolVar(&verifyPosition, "verify-position", false, "Verify portfolio position after execution")
	fs.BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	fs.BoolVar(&dryRun, "dry-run", false, "Only prepare the quote and checks")
	fs.BoolVar(&simulate, "simulate", true, "Run RPC simulation before broadcast")
	fs.BoolVar(&skipSimulate, "skip-simulate", false, "Skip transaction simulation")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if skipSimulate {
		simulate = false
	}
	if approvalAmountMode != "exact" && approvalAmountMode != "infinite" {
		return apperror.New("input", apperror.ExitInput, "--approval-amount must be exact or infinite")
	}
	if gasPolicy != "auto" && gasPolicy != "quote" && gasPolicy != "rpc" {
		return apperror.New("input", apperror.ExitInput, "--gas-policy must be one of auto, quote, or rpc")
	}
	waitTimeout, err := time.ParseDuration(waitTimeoutArg)
	if err != nil {
		return apperror.Wrap("input", apperror.ExitInput, err)
	}
	portfolioTimeout, err := time.ParseDuration(portfolioTimeoutArg)
	if err != nil {
		return apperror.Wrap("input", apperror.ExitInput, err)
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
	vault, err := rt.resolveVault(ctx, vaultArg)
	if err != nil {
		return err
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
	executor := flow.NewRPCExecutor(client)
	expectedDelta := parseFloat(formatAmount(quote.Estimate.ToAmount, quote.Action.ToToken.Decimals, 9))

	preflightResult, err := flow.ExecuteDeposit(ctx, flow.DepositRequest{
		Quote:              quote,
		Vault:              vault,
		FromChain:          fromChain,
		FromToken:          fromToken,
		WalletAddress:      walletAddress,
		Executor:           executor,
		PortfolioClient:    rt.earnClient,
		StatusClient:       rt.lifiClient,
		DryRun:             true,
		Simulate:           simulate,
		Wait:               wait,
		VerifyPosition:     verifyPosition,
		ApproveMode:        approveMode,
		ApprovalAmountMode: approvalAmountMode,
		GasPolicy:          gasPolicy,
		WaitTimeout:        waitTimeout,
		PortfolioTimeout:   portfolioTimeout,
		VerificationDelta:  expectedDelta,
	})
	if err != nil {
		return err
	}
	result := map[string]any{
		"stage":     preflightResult.Stage,
		"status":    preflightResult.Status,
		"message":   preflightResult.Message,
		"quote":     quote,
		"preflight": preflightResult.Preflight,
	}

	if !cfg.Global.JSON {
		printDepositSummary(preflightResult.Preflight, cfg.Global.NoColor)
	}
	if dryRun {
		if cfg.Global.JSON {
			return writeJSON(result)
		}
		return nil
	}

	wallet, err := rt.wallet()
	if err != nil {
		return apperror.Wrap("config", apperror.ExitConfig, err)
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

	finalResult, err := flow.ExecuteDeposit(ctx, flow.DepositRequest{
		Quote:              quote,
		Vault:              vault,
		FromChain:          fromChain,
		FromToken:          fromToken,
		WalletAddress:      walletAddress,
		Wallet:             wallet,
		Executor:           executor,
		PortfolioClient:    rt.earnClient,
		StatusClient:       rt.lifiClient,
		DryRun:             false,
		Simulate:           simulate,
		Wait:               wait,
		VerifyPosition:     verifyPosition,
		ApproveMode:        approveMode,
		ApprovalAmountMode: approvalAmountMode,
		GasPolicy:          gasPolicy,
		WaitTimeout:        waitTimeout,
		PortfolioTimeout:   portfolioTimeout,
		VerificationDelta:  expectedDelta,
	})
	if err != nil {
		return err
	}
	result["stage"] = finalResult.Stage
	result["status"] = finalResult.Status
	result["message"] = finalResult.Message
	result["tx_hash"] = finalResult.TxHash
	result["approval_tx_hash"] = finalResult.ApprovalTxHash
	result["receipt_status"] = finalResult.ReceiptStatus
	result["approval_receipt_status"] = finalResult.ApprovalReceiptStatus
	result["position_detected"] = finalResult.PositionDetected
	result["positions"] = finalResult.Positions
	result["status_payload"] = finalResult.StatusPayload
	result["explorer_url"] = explorerTxURL(fromChain, finalResult.TxHash)
	if !cfg.Global.JSON {
		fmt.Printf("deposit transaction sent: %s\n", finalResult.TxHash)
		if url := explorerTxURL(fromChain, finalResult.TxHash); url != "" {
			fmt.Printf("explorer: %s\n", url)
		}
		if finalResult.ApprovalTxHash != "" {
			fmt.Printf("approval tx: %s\n", finalResult.ApprovalTxHash)
		}
		if wait || verifyPosition {
			fmt.Printf("source receipt status: %d\n", finalResult.ReceiptStatus)
		}
		if verifyPosition {
			fmt.Printf("position detected: %s\n", boolText(finalResult.PositionDetected))
		}
	}
	if cfg.Global.JSON {
		return writeJSON(result)
	}
	return nil
}

func newWithdrawCommand() Command { return withdrawCommand{} }

func (withdrawCommand) Name() string { return "withdraw" }

func (withdrawCommand) Summary() string { return "Redeem vault shares back to the underlying token" }

func (withdrawCommand) Usage() string {
	return "lifi withdraw --vault <address> --chain <chain> --amount <shares> [--to-token <symbol-or-address>] [--dry-run] [--yes] [options]"
}

func (withdrawCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("withdraw")
	var vaultArg, chainArg, toTokenArg, amount, amountWei, fromAddress string
	var slippageBps, gasPolicy string
	var yes, dryRun bool
	fs.StringVar(&vaultArg, "vault", "", "Vault address to redeem from")
	fs.StringVar(&chainArg, "chain", "", "Chain the vault lives on")
	fs.StringVar(&toTokenArg, "to-token", "", "Output token (default: vault's underlying token)")
	fs.StringVar(&amount, "amount", "", "Shares to redeem (human-readable)")
	fs.StringVar(&amountWei, "amount-wei", "", "Shares to redeem (raw base units)")
	fs.StringVar(&fromAddress, "from-address", "", "Wallet address (overrides config)")
	fs.StringVar(&slippageBps, "slippage-bps", "", "Slippage in basis points")
	fs.StringVar(&gasPolicy, "gas-policy", "auto", "Gas pricing policy: auto, quote, or rpc")
	fs.BoolVar(&dryRun, "dry-run", false, "Simulate only — never broadcast")
	fs.BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(vaultArg) == "" {
		return fmt.Errorf("--vault is required\n\nUsage: %s\nExample: lifi withdraw --vault 0xVaultAddress --chain opt --amount 0.049", (withdrawCommand{}).Usage())
	}
	if amount == "" && amountWei == "" {
		return fmt.Errorf("--amount or --amount-wei is required\n\nRun `lifi portfolio <address>` to see your share balances")
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	vault, err := rt.resolveVault(ctx, vaultArg)
	if err != nil {
		return err
	}
	if !vault.IsRedeemable {
		return fmt.Errorf("vault %q does not support withdrawals (isRedeemable=false)", vault.Name)
	}

	chain, err := rt.resolveChain(ctx, firstNonEmpty(chainArg, fmt.Sprintf("%d", vault.ChainID)))
	if err != nil {
		return err
	}

	// Determine output token: explicit flag > vault's first underlying token.
	var toTokenAddress string
	if strings.TrimSpace(toTokenArg) != "" {
		tok, err := rt.resolveToken(ctx, chain, toTokenArg)
		if err != nil {
			return err
		}
		toTokenAddress = tok.Address
	} else if len(vault.UnderlyingTokens) > 0 {
		toTokenAddress = vault.UnderlyingTokens[0].Address
	} else {
		return fmt.Errorf("could not determine output token; pass --to-token explicitly")
	}

	walletAddr, err := rt.walletAddress(fromAddress)
	if err != nil {
		return err
	}

	// Resolve the actual decimals of the vault share token (the vault address is
	// itself the ERC-20 share token). ERC-4626 vaults use 18 by default, but we
	// look it up from the LI.FI token list to be precise.
	shareDecimals := 18
	if len(vault.LPTokens) > 0 {
		shareDecimals = vault.LPTokens[0].Decimals
	} else {
		// Try to resolve the vault address as a token on the chain.
		if shareTok, err := rt.resolveToken(ctx, chain, vault.Address); err == nil {
			shareDecimals = shareTok.Decimals
		}
	}

	fromAmount := strings.TrimSpace(amountWei)
	if fromAmount == "" {
		parsed, err := parseAmountToBaseUnits(amount, shareDecimals)
		if err != nil {
			return err
		}
		fromAmount = parsed.String()
	}

	slippage := basisPointsToSlippage(firstNonEmpty(slippageBps, cfg.DefaultSlippageBPS))
	quote, err := rt.lifiClient.GetQuote(ctx, lifiapi.QuoteRequest{
		FromChain:   chain.ID,
		ToChain:     chain.ID,
		FromToken:   vault.Address,
		ToToken:     toTokenAddress,
		FromAddress: walletAddr,
		ToAddress:   walletAddr,
		FromAmount:  fromAmount,
		Slippage:    slippage,
	})
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

	// Preflight: share balance check.
	shareBalance, err := evm.Balance(ctx, client, vault.Address, commonAddress(walletAddr))
	if err != nil {
		return apperror.Wrap("rpc", apperror.ExitRPC, err)
	}
	required := new(big.Int)
	required.SetString(fromAmount, 10)
	if shareBalance.Cmp(required) < 0 {
		return apperror.Formatf("execution", apperror.ExitExecution,
			"insufficient share balance: have %s shares, need %s",
			formatAmount(shareBalance.String(), shareDecimals, 6),
			formatAmount(fromAmount, shareDecimals, 6),
		)
	}

	nativeBalance, err := evm.Balance(ctx, client, "", commonAddress(walletAddr))
	if err != nil {
		return apperror.Wrap("rpc", apperror.ExitRPC, err)
	}

	executor := flow.NewRPCExecutor(client)
	quoteFee, err := executor.EstimateQuoteFee(ctx, quote.TransactionRequest, gasPolicy)
	if err != nil {
		return apperror.Wrap("rpc", apperror.ExitRPC, err)
	}

	// Check approval for vault share token → spender.
	approvalNeeded := false
	approvalAddress := quote.Estimate.ApprovalAddress
	if approvalAddress != "" {
		allowance, err := evm.Allowance(ctx, client, vault.Address, commonAddress(walletAddr), commonAddress(approvalAddress))
		if err != nil {
			return apperror.Wrap("rpc", apperror.ExitRPC, err)
		}
		approvalNeeded = allowance.Cmp(required) < 0
	}

	outDecimals := shareDecimals
	if len(vault.UnderlyingTokens) > 0 {
		outDecimals = vault.UnderlyingTokens[0].Decimals
	}

	printSectionHeader("Withdrawal Plan", cfg.Global.NoColor)
	printTable([]string{"field", "value"}, [][]string{
		{"wallet", walletAddr},
		{"chain", fmt.Sprintf("%s (%d)", chain.Name, chain.ID)},
		{"vault", vault.Address},
		{"shares to redeem", formatAmount(fromAmount, shareDecimals, 6)},
		{"share balance", formatAmount(shareBalance.String(), shareDecimals, 6)},
		{"expected received", formatAmount(quote.Estimate.ToAmount, outDecimals, 6)},
		{"approval needed", boolText(approvalNeeded)},
		{"native balance", formatAmount(nativeBalance.String(), chain.NativeToken.Decimals, 6)},
		{"estimated gas cost", formatAmount(quoteFee.EstimatedCost.String(), chain.NativeToken.Decimals, 6)},
		{"gas policy", gasPolicy},
	})

	if dryRun {
		return nil
	}

	wallet, err := rt.wallet()
	if err != nil {
		return apperror.Wrap("config", apperror.ExitConfig, err)
	}

	if !yes {
		confirmed, err := promptConfirm("Broadcast withdrawal transaction?")
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	if approvalNeeded {
		hash, err := executor.Approve(ctx, wallet, chain.ID, vault.Address, commonAddress(approvalAddress), required, gasPolicy)
		if err != nil {
			return apperror.Wrap("execution", apperror.ExitExecution, err)
		}
		if _, err := executor.WaitForReceipt(ctx, hash, 3*time.Second); err != nil {
			return apperror.Wrap("rpc", apperror.ExitRPC, err)
		}
	}

	hash, err := executor.SendQuote(ctx, wallet, quote.TransactionRequest, gasPolicy)
	if err != nil {
		return apperror.Wrap("execution", apperror.ExitExecution, err)
	}

	receipt, err := executor.WaitForReceipt(ctx, hash, 3*time.Second)
	if err != nil {
		return apperror.Wrap("rpc", apperror.ExitRPC, err)
	}

	if cfg.Global.JSON {
		return writeJSON(map[string]any{
			"stage":          "confirmed",
			"status":         "ok",
			"tx_hash":        hash.Hex(),
			"receipt_status": receipt.Status,
			"explorer_url":   explorerTxURL(chain, hash.Hex()),
		})
	}

	fmt.Printf("withdrawal transaction sent: %s\n", hash.Hex())
	if url := explorerTxURL(chain, hash.Hex()); url != "" {
		fmt.Printf("explorer: %s\n", url)
	}
	fmt.Printf("receipt status: %d\n", receipt.Status)
	return nil
}

func printDepositSummary(preflight *flow.Preflight, noColor bool) {
	if preflight == nil {
		return
	}
	printSectionHeader("Execution Plan", noColor)
	printTable([]string{"field", "value"}, [][]string{
		{"wallet", preflight.WalletAddress},
		{"source chain", preflight.SourceChain},
		{"source token", preflight.SourceToken},
		{"source amount", preflight.SourceAmount},
		{"vault", preflight.DestinationVault},
		{"expected received", preflight.ExpectedReceived},
		{"approval address", emptyFallback(preflight.ApprovalAddress)},
		{"approval needed", boolText(preflight.ApprovalNeeded)},
		{"approval amount", preflight.ApprovalAmount},
		{"gas policy", preflight.GasPolicy},
		{"token balance", preflight.TokenBalanceFormatted},
		{"native balance", preflight.NativeBalanceFormatted},
		{"estimated gas cost", preflight.EstimatedGasFormatted},
	})
}
