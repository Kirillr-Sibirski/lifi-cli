package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/earn"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type chainsCommand struct{}
type protocolsCommand struct{}
type tokensCommand struct{}
type vaultsCommand struct{}
type inspectCommand struct{}
type recommendCommand struct{}
type portfolioCommand struct{}
type statusCommand struct{}

func newChainsCommand() Command { return chainsCommand{} }

func (chainsCommand) Name() string { return "chains" }

func (chainsCommand) Summary() string {
	return "List LI.FI chains relevant to Earn and Composer execution"
}

func (chainsCommand) Usage() string {
	return "lifi chains [--search <query>] [--evm-only] [--json]"
}

func (chainsCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("chains")
	var search string
	var evmOnly bool
	fs.StringVar(&search, "search", "", "Filter chains by name or identifier")
	fs.BoolVar(&evmOnly, "evm-only", false, "Only show EVM chains")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	chains, err := rt.loadChains(ctx)
	if err != nil {
		return err
	}

	needle := normalizeLookup(search)
	filtered := make([]lifiapi.Chain, 0, len(chains))
	for _, chain := range chains {
		if evmOnly && !strings.EqualFold(chain.ChainType, "EVM") {
			continue
		}
		if needle != "" &&
			!strings.Contains(normalizeLookup(chain.Name), needle) &&
			!strings.Contains(normalizeLookup(chain.Key), needle) &&
			strconv.Itoa(chain.ID) != strings.TrimSpace(search) {
			continue
		}
		filtered = append(filtered, chain)
	}

	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })
	if cfg.Global.JSON {
		return writeJSON(filtered)
	}

	rows := make([][]string, 0, len(filtered))
	for _, chain := range filtered {
		rows = append(rows, []string{
			strconv.Itoa(chain.ID),
			chain.Name,
			chain.Key,
			chain.ChainType,
			chain.NativeToken.Symbol,
			boolText(chain.RelayerSupported),
		})
	}
	printTable([]string{"id", "name", "key", "type", "native", "relayer"}, rows)
	return nil
}

func newProtocolsCommand() Command { return protocolsCommand{} }

func (protocolsCommand) Name() string { return "protocols" }

func (protocolsCommand) Summary() string {
	return "List supported Earn and Composer protocols"
}

func (protocolsCommand) Usage() string {
	return "lifi protocols [--search <query>] [--supports deposit|withdraw] [--json]"
}

func (protocolsCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("protocols")
	var search string
	var supports string
	fs.StringVar(&search, "search", "", "Filter protocols by name")
	fs.StringVar(&supports, "supports", "", "Filter Earn protocols by deposit or withdraw capability")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	earnProtocols, err := rt.loadProtocols(ctx)
	if err != nil {
		return err
	}
	tools, err := rt.lifiClient.GetTools(ctx)
	if err != nil {
		return err
	}

	depositable := map[string]bool{}
	redeemable := map[string]bool{}
	if supports != "" {
		vaults, err := rt.loadAllVaults(ctx)
		if err != nil {
			return err
		}
		for _, vault := range vaults {
			if vault.IsTransactional {
				depositable[vault.Protocol.Name] = true
			}
			if vault.IsRedeemable {
				redeemable[vault.Protocol.Name] = true
			}
		}
	}

	needle := normalizeLookup(search)
	filteredEarn := make([]map[string]any, 0, len(earnProtocols))
	for _, protocol := range earnProtocols {
		if needle != "" && !strings.Contains(normalizeLookup(protocol.Name), needle) {
			continue
		}
		if supports == "deposit" && !depositable[protocol.Name] {
			continue
		}
		if supports == "withdraw" && !redeemable[protocol.Name] {
			continue
		}
		filteredEarn = append(filteredEarn, map[string]any{
			"name":             protocol.Name,
			"url":              protocol.URL,
			"supportsDeposit":  supports == "" || depositable[protocol.Name],
			"supportsWithdraw": supports == "" || redeemable[protocol.Name],
		})
	}

	filterTools := func(kind string, items []lifiapi.Tool) []map[string]any {
		out := make([]map[string]any, 0, len(items))
		for _, tool := range items {
			if needle != "" &&
				!strings.Contains(normalizeLookup(tool.Name), needle) &&
				!strings.Contains(normalizeLookup(tool.Key), needle) {
				continue
			}
			out = append(out, map[string]any{
				"kind": kind,
				"key":  tool.Key,
				"name": tool.Name,
			})
		}
		return out
	}

	payload := map[string]any{
		"earn_protocols": filteredEarn,
		"bridges":        filterTools("bridge", tools.Bridges),
		"exchanges":      filterTools("exchange", tools.Exchanges),
	}
	if cfg.Global.JSON {
		return writeJSON(payload)
	}

	printSectionHeader("Earn Protocols", cfg.Global.NoColor)
	earnRows := make([][]string, 0, len(filteredEarn))
	if supports != "" {
		// With a capability filter, show deposit/withdraw columns so the filter makes sense.
		for _, protocol := range filteredEarn {
			earnRows = append(earnRows, []string{
				fmt.Sprint(protocol["name"]),
				boolText(protocol["supportsDeposit"].(bool)),
				boolText(protocol["supportsWithdraw"].(bool)),
			})
		}
		printTable([]string{"name", "deposit", "withdraw"}, earnRows)
	} else {
		for _, protocol := range filteredEarn {
			row := []string{fmt.Sprint(protocol["name"])}
			if cfg.Global.Verbose {
				row = append(row, truncateStr(fmt.Sprint(protocol["url"]), 60))
			}
			earnRows = append(earnRows, row)
		}
		if cfg.Global.Verbose {
			printTable([]string{"name", "url"}, earnRows)
		} else {
			printTable([]string{"name"}, earnRows)
		}
	}
	fmt.Println()

	bridges := payload["bridges"].([]map[string]any)
	exchanges := payload["exchanges"].([]map[string]any)

	printSectionHeader("Bridges", cfg.Global.NoColor)
	bridgeRows := make([][]string, 0, len(bridges))
	for _, tool := range bridges {
		bridgeRows = append(bridgeRows, []string{fmt.Sprint(tool["key"]), fmt.Sprint(tool["name"])})
	}
	printTable([]string{"key", "name"}, bridgeRows)
	fmt.Println()

	printSectionHeader("Exchanges", cfg.Global.NoColor)
	exchangeRows := make([][]string, 0, len(exchanges))
	for _, tool := range exchanges {
		exchangeRows = append(exchangeRows, []string{fmt.Sprint(tool["key"]), fmt.Sprint(tool["name"])})
	}
	printTable([]string{"key", "name"}, exchangeRows)
	return nil
}

func newTokensCommand() Command { return tokensCommand{} }

func (tokensCommand) Name() string { return "tokens" }

func (tokensCommand) Summary() string { return "Resolve tokens by symbol or address" }

func (tokensCommand) Usage() string {
	return "lifi tokens [--chain <chain>] [--token <symbol-or-address>] [--tags <tag[,tag]>] [--limit <n>] [--json]"
}

func (tokensCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("tokens")
	var chainArg, tokenArg, tagsArg string
	var limit int
	fs.StringVar(&chainArg, "chain", "", "Filter tokens by chain")
	fs.StringVar(&tokenArg, "token", "", "Resolve a token by symbol or address")
	fs.StringVar(&tagsArg, "tags", "", "Filter by LI.FI token tags")
	fs.IntVar(&limit, "limit", 25, "Maximum number of results (0 = unlimited)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	var chainIDs []int
	if chainArg != "" {
		chain, err := rt.resolveChain(ctx, chainArg)
		if err != nil {
			return err
		}
		chainIDs = []int{chain.ID}
	}

	response, err := rt.lifiClient.GetTokens(ctx, chainIDs, splitCSV(tagsArg))
	if err != nil {
		return err
	}

	needle := normalizeLookup(tokenArg)
	filtered := make([]lifiapi.Token, 0)
	chainKeys := make([]string, 0, len(response.Tokens))
	for key := range response.Tokens {
		chainKeys = append(chainKeys, key)
	}
	sort.Strings(chainKeys)

	for _, key := range chainKeys {
		for _, token := range response.Tokens[key] {
			if needle != "" &&
				!strings.EqualFold(token.Address, tokenArg) &&
				!strings.Contains(normalizeLookup(token.Symbol), needle) &&
				!strings.Contains(normalizeLookup(token.Name), needle) &&
				!strings.Contains(normalizeLookup(token.CoinKey), needle) {
				continue
			}
			filtered = append(filtered, token)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].ChainID == filtered[j].ChainID {
			return filtered[i].Symbol < filtered[j].Symbol
		}
		return filtered[i].ChainID < filtered[j].ChainID
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	if cfg.Global.JSON {
		return writeJSON(filtered)
	}

	rows := make([][]string, 0, len(filtered))
	for _, token := range filtered {
		rows = append(rows, []string{
			strconv.Itoa(token.ChainID),
			token.Symbol,
			truncateStr(token.Name, 32),
			strconv.Itoa(token.Decimals),
			truncateAddr(token.Address),
			formatUSD(token.PriceUSD),
		})
	}
	printTable([]string{"chain", "symbol", "name", "decimals", "address", "price"}, rows)
	return nil
}

func newVaultsCommand() Command { return vaultsCommand{} }

func (vaultsCommand) Name() string { return "vaults" }

func (vaultsCommand) Summary() string { return "List depositable vaults" }

func (vaultsCommand) Usage() string {
	return "lifi vaults [--chain <chain>] [--asset <symbol-or-address>] [--protocol <name>] [--sort apy|apy30d|tvl|name] [--order asc|desc] [--min-tvl-usd <amount>] [--min-apy <percent>] [--transactional-only] [--limit <n>] [--json]"
}

func (vaultsCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("vaults")
	var chainArg, asset, protocol, sortBy, order, minTvl, minAPY string
	var transactionalOnly bool
	var limit int
	fs.StringVar(&chainArg, "chain", "", "Filter vaults by chain")
	fs.StringVar(&asset, "asset", "", "Filter vaults by asset")
	fs.StringVar(&protocol, "protocol", "", "Filter vaults by protocol")
	fs.StringVar(&sortBy, "sort", "apy", "Sort by apy, apy30d, tvl, or name")
	fs.StringVar(&order, "order", "desc", "Sort order: asc or desc")
	fs.StringVar(&minTvl, "min-tvl-usd", "", "Minimum TVL in USD")
	fs.StringVar(&minAPY, "min-apy", "", "Minimum APY percentage")
	fs.BoolVar(&transactionalOnly, "transactional-only", false, "Only include transactional vaults")
	fs.IntVar(&limit, "limit", 25, "Maximum number of results")
	if err := fs.Parse(args); err != nil {
		return err
	}

	vaults, err := loadAndFilterVaults(cfg, chainArg, asset, protocol, sortBy, order, minTvl, minAPY, transactionalOnly)
	if err != nil {
		return err
	}
	if limit > 0 && len(vaults) > limit {
		vaults = vaults[:limit]
	}
	if cfg.Global.JSON {
		return writeJSON(vaults)
	}

	rows := make([][]string, 0, len(vaults))
	for i, vault := range vaults {
		rows = append(rows, []string{
			strconv.Itoa(i + 1),
			vault.Name,
			vault.Protocol.Name,
			fmt.Sprintf("%s (%d)", vault.Network, vault.ChainID),
			underlyingSymbol(vault),
			formatPercent(vault.Analytics.APY.Total),
			formatPercent(derefFloat(vault.Analytics.APY30d)),
			formatUSD(vault.Analytics.TVL.USD),
			boolText(vault.IsTransactional),
			truncateAddr(vault.Address),
		})
	}
	printTable([]string{"#", "vault", "protocol", "chain", "asset", "apy", "apy30d", "tvl", "tx", "address"}, rows)
	return nil
}

func newInspectCommand() Command { return inspectCommand{} }

func (inspectCommand) Name() string { return "inspect" }

func (inspectCommand) Summary() string { return "Show full details for a vault" }

func (inspectCommand) Usage() string { return "lifi inspect <vault> [--json]" }

func (inspectCommand) Run(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("inspect requires a vault address, slug, or name")
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	vault, err := rt.resolveVault(ctx, args[0])
	if err != nil {
		return err
	}

	if cfg.Global.JSON {
		return writeJSON(vault)
	}

	rows := [][]string{
		{"name", vault.Name},
		{"slug", vault.Slug},
		{"protocol", vault.Protocol.Name},
		{"protocol url", vault.Protocol.URL},
		{"chain", fmt.Sprintf("%s (%d)", vault.Network, vault.ChainID)},
		{"address", vault.Address},
		{"asset", underlyingSymbol(*vault)},
		{"apy total", formatPercent(vault.Analytics.APY.Total)},
		{"apy base", formatPercent(vault.Analytics.APY.Base)},
		{"apy reward", formatPercent(vault.Analytics.APY.Reward)},
		{"apy 30d", formatPercent(derefFloat(vault.Analytics.APY30d))},
		{"tvl", formatUSD(vault.Analytics.TVL.USD)},
		{"transactional", boolText(vault.IsTransactional)},
		{"redeemable", boolText(vault.IsRedeemable)},
		{"deposit packs", strings.Join(packNames(vault.DepositPacks), ", ")},
		{"redeem packs", strings.Join(packNames(vault.RedeemPacks), ", ")},
		{"synced at", vault.SyncedAt},
	}
	if vault.Analytics.TVL.USD == "" {
		rows = append(rows, []string{"warning", "TVL data missing"})
	}
	if vault.Analytics.APY.Total == 0 {
		rows = append(rows, []string{"warning", "APY data may be incomplete"})
	}
	printSectionHeader("Vault Details", cfg.Global.NoColor)
	printTable([]string{"field", "value"}, rows)
	return nil
}

func newRecommendCommand() Command { return recommendCommand{} }

func (recommendCommand) Name() string { return "recommend" }

func (recommendCommand) Summary() string { return "Rank vaults for a target asset" }

func (recommendCommand) Usage() string {
	return "lifi recommend [--asset <symbol-or-address>] [--from-chain <chain>] [--to-chain <chain>] [--strategy highest-apy|safest|balanced] [--min-tvl-usd <amount>] [--limit <n>] [--json]"
}

func (recommendCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("recommend")
	var asset, fromChain, toChain, strategy, minTvl string
	var limit int
	fs.StringVar(&asset, "asset", "", "Target asset")
	fs.StringVar(&fromChain, "from-chain", "", "Source chain")
	fs.StringVar(&toChain, "to-chain", "", "Target vault chain")
	fs.StringVar(&strategy, "strategy", "balanced", "highest-apy, safest, or balanced")
	fs.StringVar(&minTvl, "min-tvl-usd", "", "Minimum TVL in USD")
	fs.IntVar(&limit, "limit", 5, "Maximum number of results")
	if err := fs.Parse(args); err != nil {
		return err
	}

	chainFilter := toChain
	if chainFilter == "" {
		chainFilter = fromChain
	}

	vaults, err := loadAndFilterVaults(cfg, chainFilter, asset, "", "apy", "desc", minTvl, "", true)
	if err != nil {
		return err
	}
	sort.Slice(vaults, func(i, j int) bool {
		return recommendationScore(vaults[i], strategy) > recommendationScore(vaults[j], strategy)
	})

	if limit > 0 && len(vaults) > limit {
		vaults = vaults[:limit]
	}
	if cfg.Global.JSON {
		return writeJSON(vaults)
	}

	rows := make([][]string, 0, len(vaults))
	for i, vault := range vaults {
		rows = append(rows, []string{
			strconv.Itoa(i + 1),
			vault.Name,
			vault.Protocol.Name,
			vault.Network,
			underlyingSymbol(vault),
			formatPercent(vault.Analytics.APY.Total),
			formatUSD(vault.Analytics.TVL.USD),
			fmt.Sprintf("%.2f", recommendationScore(vault, strategy)),
		})
	}
	printTable([]string{"#", "vault", "protocol", "chain", "asset", "apy", "tvl", "score"}, rows)
	return nil
}

func newPortfolioCommand() Command { return portfolioCommand{} }

func (portfolioCommand) Name() string { return "portfolio" }

func (portfolioCommand) Summary() string { return "Show Earn positions for an address" }

func (portfolioCommand) Usage() string {
	return "lifi portfolio <address> [--chain <chain>] [--protocol <name>] [--asset <symbol-or-address>] [--json]"
}

func (portfolioCommand) Run(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("portfolio requires an address\n\nUsage: %s\nExample: lifi portfolio 0xYourWallet --chain base", (portfolioCommand{}).Usage())
	}

	fs := newFlagSet("portfolio")
	var chainArg, protocol, asset string
	fs.StringVar(&chainArg, "chain", "", "Filter by chain")
	fs.StringVar(&protocol, "protocol", "", "Filter by protocol")
	fs.StringVar(&asset, "asset", "", "Filter by asset")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	// Load chains so we can resolve chain IDs to names in the table.
	chains, _ := rt.loadChains(ctx)
	chainNames := make(map[string]string, len(chains))
	for _, c := range chains {
		chainNames[strconv.Itoa(c.ID)] = c.Name
	}

	response, err := rt.earnClient.GetPortfolio(ctx, args[0])
	if err != nil {
		return err
	}

	filtered := response.Positions
	if chainArg != "" {
		chain, err := rt.resolveChain(ctx, chainArg)
		if err == nil {
			filtered = filterPortfolioPositions(filtered, []string{chainArg, chain.Name, chain.Key, strconv.Itoa(chain.ID)}, protocol, asset)
		} else {
			filtered = filterPortfolioPositions(filtered, []string{chainArg}, protocol, asset)
		}
	} else {
		filtered = filterPortfolioPositions(filtered, nil, protocol, asset)
	}
	if cfg.Global.JSON {
		return writeJSON(filtered)
	}

	rows := make([][]string, 0, len(filtered))
	for index, position := range filtered {
		rows = append(rows, portfolioSummaryRow(index+1, position, chainNames))
	}
	printSectionHeader("Portfolio", cfg.Global.NoColor)
	printTable([]string{"#", "chain", "protocol", "asset", "balance", "value"}, rows)
	if cfg.Global.Verbose {
		for index, position := range filtered {
			fmt.Printf("\n[%d]\n", index+1)
			blob, err := prettyJSON(position)
			if err != nil {
				return err
			}
			fmt.Println(blob)
		}
	}
	return nil
}

func newStatusCommand() Command { return statusCommand{} }

func (statusCommand) Name() string { return "status" }

func (statusCommand) Summary() string { return "Track LI.FI execution state for a transaction hash" }

func (statusCommand) Usage() string {
	return "lifi status --tx-hash <hash> [--from-chain <chain>] [--to-chain <chain>] [--bridge <name>] [--watch] [--interval <duration>] [--json]"
}

func (statusCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("status")
	var txHash, fromChainArg, toChainArg, bridge, interval string
	var watch bool
	fs.StringVar(&txHash, "tx-hash", "", "Transaction hash")
	fs.StringVar(&fromChainArg, "from-chain", "", "Source chain")
	fs.StringVar(&toChainArg, "to-chain", "", "Destination chain")
	fs.StringVar(&bridge, "bridge", "", "Bridge or tool key")
	fs.BoolVar(&watch, "watch", false, "Poll for updates continuously")
	fs.StringVar(&interval, "interval", "5s", "Polling interval")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if txHash == "" {
		return fmt.Errorf("--tx-hash is required\n\nUsage: %s\nExample: lifi status --tx-hash 0xabc123... --from-chain base", (statusCommand{}).Usage())
	}

	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	fromChainID := 0
	if fromChainArg != "" {
		chain, err := rt.resolveChain(ctx, fromChainArg)
		if err != nil {
			return err
		}
		fromChainID = chain.ID
	}

	toChainID := 0
	if toChainArg != "" {
		chain, err := rt.resolveChain(ctx, toChainArg)
		if err != nil {
			return err
		}
		toChainID = chain.ID
	}

	pollInterval, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}

	for {
		pollCtx, pollCancel := context.WithTimeout(context.Background(), 20*time.Second)
		payload, err := rt.lifiClient.GetStatus(pollCtx, lifiapi.StatusRequest{
			TxHash:    txHash,
			Bridge:    bridge,
			FromChain: fromChainID,
			ToChain:   toChainID,
		})
		pollCancel()
		if err != nil {
			return err
		}

		if cfg.Global.JSON {
			if watch {
				encoder := json.NewEncoder(osStdout{})
				if err := encoder.Encode(payload); err != nil {
					return err
				}
			} else {
				return writeJSON(payload)
			}
		} else {
			printSectionHeader("Status", cfg.Global.NoColor)
			printTable([]string{"field", "value"}, statusSummaryRows(payload))
			if cfg.Global.Verbose {
				fmt.Println()
				blob, err := prettyJSON(payload)
				if err != nil {
					return err
				}
				fmt.Println(blob)
			}
		}

		if !watch || isTerminalStatus(payload) {
			return nil
		}
		time.Sleep(pollInterval)
	}
}

func loadAndFilterVaults(cfg *config.Config, chainArg, asset, protocol, sortBy, order, minTvl, minAPY string, transactionalOnly bool) ([]earn.Vault, error) {
	rt := newRuntime(cfg)
	ctx, cancel := rt.context()
	defer cancel()

	query := earn.VaultQuery{Limit: 100}
	if chainArg != "" {
		chain, err := rt.resolveChain(ctx, chainArg)
		if err != nil {
			return nil, err
		}
		query.ChainID = chain.ID
	}
	if asset != "" && !strings.HasPrefix(strings.ToLower(asset), "0x") {
		query.Asset = asset
	}
	if minTvl != "" {
		query.MinTvlUSD = minTvl
	}

	vaults, err := rt.earnClient.GetAllVaults(ctx, query)
	if err != nil {
		return nil, err
	}

	needleProtocol := normalizeLookup(protocol)
	minAPYFloat := parseFloat(minAPY)
	for i := 0; i < len(vaults); {
		vault := vaults[i]
		if transactionalOnly && !vault.IsTransactional {
			vaults = append(vaults[:i], vaults[i+1:]...)
			continue
		}
		if needleProtocol != "" && normalizeLookup(vault.Protocol.Name) != needleProtocol {
			vaults = append(vaults[:i], vaults[i+1:]...)
			continue
		}
		if asset != "" && !vaultMatchesAsset(vault, asset) {
			vaults = append(vaults[:i], vaults[i+1:]...)
			continue
		}
		if minAPYFloat > 0 && vault.Analytics.APY.Total < minAPYFloat {
			vaults = append(vaults[:i], vaults[i+1:]...)
			continue
		}
		i++
	}

	sortVaults(vaults, sortBy, order)
	return vaults, nil
}

func sortVaults(vaults []earn.Vault, sortBy, order string) {
	desc := !strings.EqualFold(order, "asc")
	sort.Slice(vaults, func(i, j int) bool {
		left, right := vaults[i], vaults[j]
		var less bool
		switch strings.ToLower(strings.TrimSpace(sortBy)) {
		case "apy30d":
			less = derefFloat(left.Analytics.APY30d) < derefFloat(right.Analytics.APY30d)
		case "tvl":
			less = parseFloat(left.Analytics.TVL.USD) < parseFloat(right.Analytics.TVL.USD)
		case "name":
			less = left.Name < right.Name
		default:
			less = left.Analytics.APY.Total < right.Analytics.APY.Total
		}
		if desc {
			return !less
		}
		return less
	})
}

func vaultMatchesAsset(vault earn.Vault, asset string) bool {
	if asset == "" {
		return true
	}
	for _, token := range vault.UnderlyingTokens {
		if strings.EqualFold(token.Address, asset) || normalizeLookup(token.Symbol) == normalizeLookup(asset) {
			return true
		}
	}
	return false
}

func underlyingSymbol(vault earn.Vault) string {
	if len(vault.UnderlyingTokens) > 0 {
		return vault.UnderlyingTokens[0].Symbol
	}
	return "-"
}

func packNames(packs []earn.VaultPack) []string {
	names := make([]string, 0, len(packs))
	for _, pack := range packs {
		names = append(names, pack.Name)
	}
	return names
}

func derefFloat(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func recommendationScore(vault earn.Vault, strategy string) float64 {
	apy := vault.Analytics.APY.Total
	tvl := parseFloat(vault.Analytics.TVL.USD)
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "highest-apy":
		return apy
	case "safest":
		return math.Log10(tvl+1)*10 + vault.Analytics.APY.Base
	default:
		return apy + math.Log10(tvl+1)*3
	}
}

func filterPortfolioPositions(positions []map[string]any, chainNeedles []string, protocol, asset string) []map[string]any {
	normalizedChains := make([]string, 0, len(chainNeedles))
	for _, chain := range chainNeedles {
		needle := normalizeLookup(chain)
		if needle != "" {
			normalizedChains = append(normalizedChains, needle)
		}
	}
	needleProtocol := normalizeLookup(protocol)
	needleAsset := normalizeLookup(asset)
	filtered := make([]map[string]any, 0, len(positions))
	for _, position := range positions {
		blob, err := json.Marshal(position)
		if err != nil {
			continue
		}
		text := normalizeLookup(string(blob))
		if len(normalizedChains) > 0 {
			matched := false
			for _, needleChain := range normalizedChains {
				if strings.Contains(text, needleChain) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if needleProtocol != "" && !strings.Contains(text, needleProtocol) {
			continue
		}
		if needleAsset != "" && !strings.Contains(text, needleAsset) {
			continue
		}
		filtered = append(filtered, position)
	}
	return filtered
}

func isTerminalStatus(payload map[string]any) bool {
	status := strings.ToUpper(fmt.Sprint(payload["status"]))
	switch status {
	case "DONE", "FAILED", "NOT_FOUND", "INVALID":
		return true
	default:
		return false
	}
}

func portfolioSummaryRow(index int, position map[string]any, chainNames map[string]string) []string {
	assetSymbol := "-"
	if asset, ok := position["asset"].(map[string]any); ok {
		if symbol := strings.TrimSpace(fmt.Sprint(asset["symbol"])); symbol != "" {
			assetSymbol = symbol
		}
	}

	valueStr := nilSafe(position["valueUsd"])
	if valueStr != "" && valueStr != "-" {
		if v, err := strconv.ParseFloat(valueStr, 64); err == nil {
			valueStr = fmt.Sprintf("$%.2f", v)
		}
	}

	return []string{
		strconv.Itoa(index),
		chainLabelForPosition(position, chainNames),
		nilSafe(position["protocolName"]),
		assetSymbol,
		nilSafe(position["balanceNative"]),
		valueStr,
	}
}

// nilSafe converts any value to string, returning "-" for nil or "<nil>".
func nilSafe(v any) string {
	if v == nil {
		return "-"
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "" || s == "<nil>" {
		return "-"
	}
	return s
}

func chainLabelForPosition(position map[string]any, chainNames map[string]string) string {
	chainID := strings.TrimSpace(fmt.Sprint(position["chainId"]))
	if chainID == "" || chainID == "<nil>" {
		return "-"
	}
	if name, ok := chainNames[chainID]; ok {
		return name
	}
	return chainID
}

func statusSummaryRows(payload map[string]any) [][]string {
	rows := [][]string{
		{"status", fmt.Sprint(payload["status"])},
	}
	if value := strings.TrimSpace(fmt.Sprint(payload["substatus"])); value != "" && value != "<nil>" {
		rows = append(rows, []string{"substatus", value})
	}
	if value := strings.TrimSpace(fmt.Sprint(payload["bridge"])); value != "" && value != "<nil>" {
		rows = append(rows, []string{"bridge", value})
	}
	if sending, ok := payload["sending"].(map[string]any); ok {
		if txHash := strings.TrimSpace(fmt.Sprint(sending["txHash"])); txHash != "" && txHash != "<nil>" {
			rows = append(rows, []string{"sending tx", txHash})
		}
	}
	if receiving, ok := payload["receiving"].(map[string]any); ok {
		if txHash := strings.TrimSpace(fmt.Sprint(receiving["txHash"])); txHash != "" && txHash != "<nil>" {
			rows = append(rows, []string{"receiving tx", txHash})
		}
	}
	if tool := strings.TrimSpace(fmt.Sprint(payload["tool"])); tool != "" && tool != "<nil>" {
		rows = append(rows, []string{"tool", tool})
	}
	return rows
}
