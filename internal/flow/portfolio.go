package flow

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/earn"
)

func PortfolioShowsDeposit(positions []map[string]any, vault earn.Vault, baselineTotal, expectedDelta float64) bool {
	if findVaultInPositions(positions, vault) {
		return true
	}
	current := portfolioBalanceForVaultAsset(positions, vault)
	if current <= baselineTotal {
		return false
	}
	if expectedDelta <= 0 {
		return current > baselineTotal
	}
	minimumDelta := expectedDelta * 0.5
	return current-baselineTotal >= minimumDelta
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

func normalizeLookup(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func parseFloat(value string) float64 {
	return 0 + func() float64 {
		var parsed float64
		fmt.Sscan(strings.TrimSpace(value), &parsed)
		return parsed
	}()
}

func parseMaybeInt(value string) int {
	var parsed int
	fmt.Sscan(strings.TrimSpace(value), &parsed)
	return parsed
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
