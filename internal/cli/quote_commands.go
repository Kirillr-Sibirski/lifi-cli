package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type quoteCommand struct{}

func newQuoteCommand() Command { return quoteCommand{} }

func (quoteCommand) Name() string { return "quote" }

func (quoteCommand) Summary() string {
	return "Generate a Composer quote for a vault deposit"
}

func (quoteCommand) Usage() string {
	return "lifi quote --vault <address> --from-chain <chain> --from-token <symbol-or-address> --amount <human> --from-address <address> [options]"
}

func (quoteCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("quote")
	var vaultArg, fromChainArg, toChainArg, fromTokenArg, amount, amountWei, fromAddress, toAddress string
	var slippageBps, preset, allowBridges, denyBridges, allowExchanges, denyExchanges string
	var raw bool
	fs.StringVar(&vaultArg, "vault", "", "Target vault address")
	fs.StringVar(&fromChainArg, "from-chain", "", "Source chain")
	fs.StringVar(&toChainArg, "to-chain", "", "Destination chain")
	fs.StringVar(&fromTokenArg, "from-token", "", "Source token")
	fs.StringVar(&amount, "amount", "", "Human-readable amount")
	fs.StringVar(&amountWei, "amount-wei", "", "Raw amount in base units")
	fs.StringVar(&fromAddress, "from-address", "", "Source wallet address")
	fs.StringVar(&toAddress, "to-address", "", "Destination wallet address")
	fs.StringVar(&slippageBps, "slippage-bps", "", "Allowed slippage in basis points")
	fs.StringVar(&preset, "preset", "", "Quote preset")
	fs.StringVar(&allowBridges, "allow-bridges", "", "Allowlisted bridges")
	fs.StringVar(&denyBridges, "deny-bridges", "", "Denylisted bridges")
	fs.StringVar(&allowExchanges, "allow-exchanges", "", "Allowlisted exchanges")
	fs.StringVar(&denyExchanges, "deny-exchanges", "", "Denylisted exchanges")
	fs.BoolVar(&raw, "raw", false, "Print raw transaction payload details")
	if err := fs.Parse(args); err != nil {
		return err
	}

	quote, _, _, _, err := prepareQuote(cfg, quoteInputs{
		vaultArg:       vaultArg,
		fromChainArg:   fromChainArg,
		toChainArg:     toChainArg,
		fromTokenArg:   fromTokenArg,
		amount:         amount,
		amountWei:      amountWei,
		fromAddress:    fromAddress,
		toAddress:      toAddress,
		slippageBps:    slippageBps,
		preset:         preset,
		allowBridges:   allowBridges,
		denyBridges:    denyBridges,
		allowExchanges: allowExchanges,
		denyExchanges:  denyExchanges,
	})
	if err != nil {
		return err
	}

	if cfg.Global.JSON {
		if raw {
			return writeJSON(map[string]any{
				"quote":              quote,
				"transactionRequest": quote.TransactionRequest,
			})
		}
		return writeJSON(quote)
	}

	printTable([]string{"field", "value"}, quoteSummaryRows(quote))
	if raw {
		fmt.Println()
		fmt.Println("transaction request")
		blob, err := prettyJSON(quote.TransactionRequest)
		if err != nil {
			return err
		}
		fmt.Println(blob)
	}
	return nil
}

type quoteInputs struct {
	vaultArg       string
	fromChainArg   string
	toChainArg     string
	fromTokenArg   string
	amount         string
	amountWei      string
	fromAddress    string
	toAddress      string
	slippageBps    string
	preset         string
	allowBridges   string
	denyBridges    string
	allowExchanges string
	denyExchanges  string
}

func prepareQuote(cfg *config.Config, in quoteInputs) (*lifiapi.Quote, *lifiapi.Chain, *lifiapi.Token, string, error) {
	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	if strings.TrimSpace(in.vaultArg) == "" {
		return nil, nil, nil, "", fmt.Errorf("--vault is required")
	}
	vault, err := rt.resolveVault(ctx, in.vaultArg)
	if err != nil {
		return nil, nil, nil, "", err
	}

	fromChain, err := rt.resolveChain(ctx, in.fromChainArg)
	if err != nil {
		return nil, nil, nil, "", err
	}

	toChain := fromChain
	if in.toChainArg != "" {
		toChain, err = rt.resolveChain(ctx, in.toChainArg)
		if err != nil {
			return nil, nil, nil, "", err
		}
	} else {
		toChain, err = rt.resolveChain(ctx, fmt.Sprintf("%d", vault.ChainID))
		if err != nil {
			return nil, nil, nil, "", err
		}
	}

	fromToken, err := rt.resolveToken(ctx, fromChain, in.fromTokenArg)
	if err != nil {
		return nil, nil, nil, "", err
	}

	resolvedFromAddress, err := rt.walletAddress(in.fromAddress)
	if err != nil {
		return nil, nil, nil, "", err
	}
	resolvedToAddress := in.toAddress
	if strings.TrimSpace(resolvedToAddress) == "" {
		resolvedToAddress = resolvedFromAddress
	}

	fromAmount := strings.TrimSpace(in.amountWei)
	if in.amount != "" && in.amountWei != "" {
		return nil, nil, nil, "", fmt.Errorf("--amount and --amount-wei are mutually exclusive")
	}
	if fromAmount == "" {
		if strings.TrimSpace(in.amount) == "" {
			return nil, nil, nil, "", fmt.Errorf("either --amount or --amount-wei is required")
		}
		parsed, err := parseAmountToBaseUnits(in.amount, fromToken.Decimals)
		if err != nil {
			return nil, nil, nil, "", err
		}
		fromAmount = parsed.String()
	}

	slippage := basisPointsToSlippage(firstNonEmpty(in.slippageBps, cfg.DefaultSlippageBPS))
	quote, err := rt.lifiClient.GetQuote(ctx, lifiapi.QuoteRequest{
		FromChain:      fromChain.ID,
		ToChain:        toChain.ID,
		FromToken:      fromToken.Address,
		ToToken:        vault.Address,
		FromAddress:    resolvedFromAddress,
		ToAddress:      resolvedToAddress,
		FromAmount:     fromAmount,
		Slippage:       slippage,
		Preset:         in.preset,
		AllowBridges:   splitCSV(in.allowBridges),
		DenyBridges:    splitCSV(in.denyBridges),
		AllowExchanges: splitCSV(in.allowExchanges),
		DenyExchanges:  splitCSV(in.denyExchanges),
	})
	if err != nil {
		return nil, nil, nil, "", err
	}

	return quote, fromChain, fromToken, resolvedFromAddress, nil
}

func basisPointsToSlippage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed := parseFloat(value)
	if parsed <= 0 {
		return ""
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", parsed/10000), "0"), ".")
}

func readQuoteFile(path string) (*lifiapi.Quote, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var quote lifiapi.Quote
	if err := json.Unmarshal(data, &quote); err != nil {
		return nil, err
	}
	return &quote, nil
}
