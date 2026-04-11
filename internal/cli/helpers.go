package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/earn"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/evm"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type runtime struct {
	cfg        *config.Config
	earnClient *earn.Client
	lifiClient *lifiapi.Client

	chains        []lifiapi.Chain
	earnChains    []earn.Chain
	protocols     []earn.Protocol
	allVaults     []earn.Vault
	tokensByChain map[int][]lifiapi.Token
}

func newRuntime(cfg *config.Config) *runtime {
	return &runtime{
		cfg:           cfg,
		earnClient:    earn.New(),
		lifiClient:    lifiapi.New(cfg.APIKey),
		tokensByChain: map[int][]lifiapi.Token{},
	}
}

func (rt *runtime) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func (rt *runtime) loadChains(ctx context.Context) ([]lifiapi.Chain, error) {
	if rt.chains != nil {
		return rt.chains, nil
	}
	chains, err := rt.lifiClient.GetChains(ctx)
	if err != nil {
		return nil, err
	}
	rt.chains = chains
	return chains, nil
}

func (rt *runtime) loadEarnChains(ctx context.Context) ([]earn.Chain, error) {
	if rt.earnChains != nil {
		return rt.earnChains, nil
	}
	chains, err := rt.earnClient.GetChains(ctx)
	if err != nil {
		return nil, err
	}
	rt.earnChains = chains
	return chains, nil
}

func (rt *runtime) loadProtocols(ctx context.Context) ([]earn.Protocol, error) {
	if rt.protocols != nil {
		return rt.protocols, nil
	}
	protocols, err := rt.earnClient.GetProtocols(ctx)
	if err != nil {
		return nil, err
	}
	rt.protocols = protocols
	return protocols, nil
}

func (rt *runtime) loadAllVaults(ctx context.Context) ([]earn.Vault, error) {
	if rt.allVaults != nil {
		return rt.allVaults, nil
	}
	vaults, err := rt.earnClient.GetAllVaults(ctx, earn.VaultQuery{Limit: 100})
	if err != nil {
		return nil, err
	}
	rt.allVaults = vaults
	return vaults, nil
}

func (rt *runtime) resolveChain(ctx context.Context, value string) (*lifiapi.Chain, error) {
	if strings.TrimSpace(value) == "" {
		value = rt.cfg.DefaultFromChain
	}
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("chain is required; pass --chain <name-or-id> or set LIFI_DEFAULT_FROM_CHAIN in .env\nRun `lifi chains` to see available chains")
	}

	chains, err := rt.loadChains(ctx)
	if err != nil {
		return nil, err
	}

	needle := normalizeLookup(value)

	// Exact matches first
	for _, chain := range chains {
		switch {
		case strconv.Itoa(chain.ID) == value:
			return &chain, nil
		case normalizeLookup(chain.Name) == needle:
			return &chain, nil
		case normalizeLookup(chain.Key) == needle:
			return &chain, nil
		case normalizeLookup(chain.Coin) == needle && len(value) <= 5:
			return &chain, nil
		}
	}

	// Fuzzy fallback: prefix or substring match on name/key
	var candidates []lifiapi.Chain
	for _, chain := range chains {
		normName := normalizeLookup(chain.Name)
		normKey := normalizeLookup(chain.Key)
		if strings.HasPrefix(normName, needle) || strings.HasPrefix(normKey, needle) ||
			strings.Contains(normName, needle) {
			candidates = append(candidates, chain)
		}
	}
	if len(candidates) == 1 {
		return &candidates[0], nil
	}
	if len(candidates) > 1 {
		limit := min(5, len(candidates))
		hints := make([]string, limit)
		for i := 0; i < limit; i++ {
			hints[i] = fmt.Sprintf("%s (key: %s)", candidates[i].Name, candidates[i].Key)
		}
		return nil, fmt.Errorf("chain %q is ambiguous; did you mean one of:\n  %s\nRun `lifi chains` to see all available chains", value, strings.Join(hints, "\n  "))
	}

	return nil, fmt.Errorf("unknown chain %q\nRun `lifi chains` to see available chains and their keys", value)
}

func (rt *runtime) resolveVault(ctx context.Context, value string) (*earn.Vault, error) {
	vaults, err := rt.loadAllVaults(ctx)
	if err != nil {
		return nil, err
	}

	needle := normalizeLookup(value)
	for _, vault := range vaults {
		switch {
		case strings.EqualFold(vault.Address, value):
			return &vault, nil
		case normalizeLookup(vault.Slug) == needle:
			return &vault, nil
		case normalizeLookup(vault.Name) == needle:
			return &vault, nil
		}
	}

	return nil, fmt.Errorf("vault %q not found", value)
}

func (rt *runtime) resolveToken(ctx context.Context, chain *lifiapi.Chain, value string) (*lifiapi.Token, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("--token is required; pass a symbol (e.g. USDC, ETH) or a token address\nRun `lifi tokens --chain %s` to browse available tokens", chain.Key)
	}

	if evm.IsNativeToken(value) || normalizeLookup(value) == normalizeLookup(chain.Coin) || normalizeLookup(value) == normalizeLookup(chain.NativeToken.Symbol) {
		token := chain.NativeToken
		return &token, nil
	}

	tokens, err := rt.tokensForChain(ctx, chain.ID)
	if err != nil {
		return nil, err
	}

	needle := normalizeLookup(value)
	for _, token := range tokens {
		switch {
		case strings.EqualFold(token.Address, value):
			return &token, nil
		case normalizeLookup(token.Symbol) == needle:
			return &token, nil
		case normalizeLookup(token.CoinKey) == needle:
			return &token, nil
		case normalizeLookup(token.Name) == needle:
			return &token, nil
		}
	}

	return nil, fmt.Errorf("token %q not found on %s\nRun `lifi tokens --chain %s --token %s` to search for it", value, chain.Name, chain.Key, value)
}

func (rt *runtime) tokensForChain(ctx context.Context, chainID int) ([]lifiapi.Token, error) {
	if tokens, ok := rt.tokensByChain[chainID]; ok {
		return tokens, nil
	}

	response, err := rt.lifiClient.GetTokens(ctx, []int{chainID}, nil)
	if err != nil {
		return nil, err
	}

	tokens := response.Tokens[strconv.Itoa(chainID)]
	rt.tokensByChain[chainID] = tokens
	return tokens, nil
}

func (rt *runtime) walletAddress(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return explicit, nil
	}
	if strings.TrimSpace(rt.cfg.WalletAddress) != "" {
		return rt.cfg.WalletAddress, nil
	}
	return "", errors.New("wallet address is required; pass --from-address or set LIFI_WALLET_ADDRESS")
}

func (rt *runtime) wallet() (*evm.Wallet, error) {
	if strings.TrimSpace(rt.cfg.WalletPrivateKey) == "" {
		return nil, errors.New("wallet private key is required; set LIFI_WALLET_PRIVATE_KEY or provide it through .env")
	}
	return evm.WalletFromHex(rt.cfg.WalletPrivateKey)
}

func (rt *runtime) rpcURL(chain *lifiapi.Chain) (string, error) {
	keys := []string{
		normalizeLookup(chain.Name),
		normalizeLookup(chain.Key),
		strconv.Itoa(chain.ID),
	}
	for _, key := range keys {
		if url := rt.cfg.LookupRPC(key); url != "" {
			return url, nil
		}
	}

	if len(chain.Metamask.RPCURLs) > 0 && chain.Metamask.RPCURLs[0] != "" {
		return chain.Metamask.RPCURLs[0], nil
	}

	return "", fmt.Errorf("no RPC URL configured for %s", chain.Name)
}

func normalizeLookup(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseAmountToBaseUnits(value string, decimals int) (*big.Int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("amount is required")
	}

	if strings.HasPrefix(value, "-") {
		return nil, errors.New("amount must be positive")
	}

	parts := strings.SplitN(value, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	if fracPart != "" && len(fracPart) > decimals {
		return nil, fmt.Errorf("amount has too many decimal places for token with %d decimals", decimals)
	}

	if intPart == "" {
		intPart = "0"
	}

	fracPart = fracPart + strings.Repeat("0", decimals-len(fracPart))
	combined := strings.TrimLeft(intPart+fracPart, "0")
	if combined == "" {
		combined = "0"
	}

	amount := new(big.Int)
	if _, ok := amount.SetString(combined, 10); !ok {
		return nil, fmt.Errorf("invalid amount %q", value)
	}
	return amount, nil
}

func formatAmount(raw string, decimals int, precision int) string {
	if raw == "" {
		return "0"
	}

	value := new(big.Int)
	if _, ok := value.SetString(raw, 10); !ok {
		return raw
	}

	if decimals == 0 {
		return value.String()
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	intPart := new(big.Int).Div(value, divisor)
	fracPart := new(big.Int).Mod(value, divisor)

	fracString := fmt.Sprintf("%0*s", decimals, fracPart.String())
	fracString = strings.TrimRight(fracString, "0")
	if precision > 0 && len(fracString) > precision {
		fracString = fracString[:precision]
	}
	if fracString == "" {
		return intPart.String()
	}

	return intPart.String() + "." + fracString
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}

func formatUSD(value string) string {
	if value == "" {
		return "-"
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}
	if math.Abs(parsed) >= 1000 {
		return fmt.Sprintf("$%.0f", parsed)
	}
	if parsed >= 100 {
		return fmt.Sprintf("$%.2f", parsed)
	}
	return fmt.Sprintf("$%.4f", parsed)
}

func boolText(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

// tableNoColor is set by app.Run once the config is loaded.
var tableNoColor bool

func printTable(headers []string, rows [][]string) {
	cols := len(headers)
	if cols == 0 {
		return
	}
	widths := make([]int, cols)
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := 0; i < len(row) && i < cols; i++ {
			if w := len(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}

	sep := colorize("│", "90", tableNoColor)

	// Header row
	hParts := make([]string, cols)
	for i, h := range headers {
		cell := fmt.Sprintf(" %-*s ", widths[i], h)
		hParts[i] = colorize(cell, "1", tableNoColor)
	}
	fmt.Println(strings.Join(hParts, sep))

	// Divider
	divParts := make([]string, cols)
	for i, w := range widths {
		divParts[i] = strings.Repeat("─", w+2)
	}
	fmt.Println(colorize(strings.Join(divParts, "┼"), "90", tableNoColor))

	// Data rows
	for _, row := range rows {
		parts := make([]string, cols)
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			padded := fmt.Sprintf(" %-*s ", widths[i], cell)
			parts[i] = styledTableCell(padded, cell)
		}
		fmt.Println(strings.Join(parts, sep))
	}
}

func styledTableCell(padded, raw string) string {
	if tableNoColor {
		return padded
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "yes":
		return colorize(padded, "32", false)
	case "no":
		return colorize(padded, "90", false)
	}
	return padded
}

func promptConfirm(message string) (bool, error) {
	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

func parseMaybeInt(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return parsed
}

func quoteSummaryRows(quote *lifiapi.Quote) [][]string {
	rows := [][]string{
		{"tool", quote.ToolDetails.Name},
		{"from chain", strconv.Itoa(quote.Action.FromChainID)},
		{"to chain", strconv.Itoa(quote.Action.ToChainID)},
		{"from token", quote.Action.FromToken.Symbol},
		{"to token", quote.Action.ToToken.Symbol},
		{"from amount", formatAmount(quote.Action.FromAmount, quote.Action.FromToken.Decimals, 6)},
		{"to amount", formatAmount(quote.Estimate.ToAmount, quote.Action.ToToken.Decimals, 6)},
		{"min received", formatAmount(quote.Estimate.ToAmountMin, quote.Action.ToToken.Decimals, 6)},
		{"approval address", emptyFallback(quote.Estimate.ApprovalAddress)},
	}

	gasUSD := "0"
	for _, gasCost := range quote.Estimate.GasCosts {
		if gasCost.AmountUSD != "" {
			gasUSD = gasCost.AmountUSD
			break
		}
	}
	rows = append(rows, []string{"gas usd", formatUSD(gasUSD)})
	return rows
}

func prettyJSON(value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func uniqueStrings(values []string) []string {
	set := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			continue
		}
		set[value] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func findVaultInPositions(positions []map[string]any, vault earn.Vault) bool {
	needleAddress := strings.ToLower(vault.Address)
	needleProtocol := normalizeLookup(vault.Protocol.Name)
	assetNeedles := uniqueStrings(vaultAssetNeedles(vault))

	for _, position := range positions {
		if positionMatchesVault(position, needleProtocol, vault.ChainID, assetNeedles) {
			return true
		}

		blob, err := json.Marshal(position)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(string(blob)), needleAddress) {
			return true
		}
	}
	return false
}

func portfolioBalanceForVaultAsset(positions []map[string]any, vault earn.Vault) float64 {
	assetNeedles := uniqueStrings(vaultAssetNeedles(vault))
	total := 0.0
	for _, position := range positions {
		if !positionMatchesVault(position, "", vault.ChainID, assetNeedles) {
			continue
		}
		total += parseFloat(fmt.Sprint(position["balanceNative"]))
	}
	return total
}

func vaultAssetNeedles(vault earn.Vault) []string {
	values := make([]string, 0, len(vault.UnderlyingTokens)*2+len(vault.LPTokens)*2)
	for _, token := range vault.UnderlyingTokens {
		values = append(values, token.Address, token.Symbol)
	}
	for _, token := range vault.LPTokens {
		values = append(values, token.Address, token.Symbol)
	}
	return values
}

func positionMatchesVault(position map[string]any, protocol string, chainID int, assets []string) bool {
	if protocol != "" && normalizeLookup(fmt.Sprint(position["protocolName"])) != protocol {
		return false
	}

	switch value := position["chainId"].(type) {
	case float64:
		if int(value) != chainID {
			return false
		}
	case int:
		if value != chainID {
			return false
		}
	case string:
		if parseMaybeInt(value) != chainID {
			return false
		}
	}

	if len(assets) == 0 {
		return true
	}

	asset, ok := position["asset"].(map[string]any)
	if !ok {
		return false
	}
	address := normalizeLookup(fmt.Sprint(asset["address"]))
	symbol := normalizeLookup(fmt.Sprint(asset["symbol"]))
	for _, candidate := range assets {
		candidate = normalizeLookup(candidate)
		if candidate == "" {
			continue
		}
		if candidate == address || candidate == symbol {
			return true
		}
	}
	return false
}

func commonAddress(value string) common.Address {
	return common.HexToAddress(value)
}

func explorerTxURL(chain *lifiapi.Chain, txHash string) string {
	if chain == nil || txHash == "" {
		return ""
	}
	if len(chain.Metamask.BlockExplorerURLs) == 0 || chain.Metamask.BlockExplorerURLs[0] == "" {
		return ""
	}
	base := strings.TrimRight(chain.Metamask.BlockExplorerURLs[0], "/")
	return base + "/tx/" + txHash
}
