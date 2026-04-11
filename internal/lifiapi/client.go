package lifiapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://li.quest/v1"

type Client struct {
	baseURL string
	http    *http.Client
	apiKey  string
}

type Chain struct {
	Key              string   `json:"key"`
	ChainType        string   `json:"chainType"`
	Name             string   `json:"name"`
	Coin             string   `json:"coin"`
	ID               int      `json:"id"`
	Mainnet          bool     `json:"mainnet"`
	LogoURI          string   `json:"logoURI"`
	RelayerSupported bool     `json:"relayerSupported"`
	Metamask         Metamask `json:"metamask"`
	NativeToken      Token    `json:"nativeToken"`
	DiamondAddress   string   `json:"diamondAddress"`
	Permit2          string   `json:"permit2"`
	Permit2Proxy     string   `json:"permit2Proxy"`
}

type Metamask struct {
	ChainID           string   `json:"chainId"`
	BlockExplorerURLs []string `json:"blockExplorerUrls"`
	ChainName         string   `json:"chainName"`
	RPCURLs           []string `json:"rpcUrls"`
}

type Token struct {
	ChainID                     int      `json:"chainId"`
	Address                     string   `json:"address"`
	Symbol                      string   `json:"symbol"`
	Name                        string   `json:"name"`
	Decimals                    int      `json:"decimals"`
	PriceUSD                    string   `json:"priceUSD"`
	CoinKey                     string   `json:"coinKey"`
	LogoURI                     string   `json:"logoURI"`
	Tags                        []string `json:"tags"`
	VerificationStatus          string   `json:"verificationStatus"`
	VerificationStatusBreakdown []any    `json:"verificationStatusBreakdown"`
}

type ToolsResponse struct {
	Bridges   []Tool `json:"bridges"`
	Exchanges []Tool `json:"exchanges"`
}

type Tool struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	LogoURI         string `json:"logoURI"`
	SupportedChains []any  `json:"supportedChains"`
}

type TokensResponse struct {
	Tokens   map[string][]Token `json:"tokens"`
	Extended bool               `json:"extended"`
}

type QuoteRequest struct {
	FromChain      int
	ToChain        int
	FromToken      string
	ToToken        string
	FromAddress    string
	ToAddress      string
	FromAmount     string
	Slippage       string
	Preset         string
	AllowBridges   []string
	DenyBridges    []string
	AllowExchanges []string
	DenyExchanges  []string
}

type Quote struct {
	Type               string             `json:"type"`
	ID                 string             `json:"id"`
	Tool               string             `json:"tool"`
	ToolDetails        ToolDetails        `json:"toolDetails"`
	Action             QuoteAction        `json:"action"`
	Estimate           QuoteEstimate      `json:"estimate"`
	IncludedSteps      []QuoteStep        `json:"includedSteps"`
	Integrator         string             `json:"integrator"`
	TransactionRequest TransactionRequest `json:"transactionRequest"`
	TransactionID      string             `json:"transactionId"`
}

type ToolDetails struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	LogoURI string `json:"logoURI"`
}

type QuoteAction struct {
	FromToken   Token   `json:"fromToken"`
	FromAmount  string  `json:"fromAmount"`
	ToToken     Token   `json:"toToken"`
	FromChainID int     `json:"fromChainId"`
	ToChainID   int     `json:"toChainId"`
	Slippage    float64 `json:"slippage"`
	FromAddress string  `json:"fromAddress"`
	ToAddress   string  `json:"toAddress"`
}

type QuoteEstimate struct {
	Tool              string    `json:"tool"`
	ApprovalAddress   string    `json:"approvalAddress"`
	ToAmountMin       string    `json:"toAmountMin"`
	ToAmount          string    `json:"toAmount"`
	FromAmount        string    `json:"fromAmount"`
	FeeCosts          []FeeCost `json:"feeCosts"`
	GasCosts          []GasCost `json:"gasCosts"`
	ExecutionDuration int       `json:"executionDuration"`
	FromAmountUSD     string    `json:"fromAmountUSD"`
	ToAmountUSD       string    `json:"toAmountUSD"`
}

type FeeCost struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Token       Token                  `json:"token"`
	Amount      string                 `json:"amount"`
	AmountUSD   string                 `json:"amountUSD"`
	Percentage  string                 `json:"percentage"`
	Included    bool                   `json:"included"`
	FeeSplit    map[string]string      `json:"feeSplit"`
	Raw         map[string]interface{} `json:"-"`
}

type GasCost struct {
	Type      string `json:"type"`
	Price     string `json:"price"`
	Estimate  string `json:"estimate"`
	Limit     string `json:"limit"`
	Amount    string `json:"amount"`
	AmountUSD string `json:"amountUSD"`
	Token     Token  `json:"token"`
}

type QuoteStep struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	Action      QuoteAction   `json:"action"`
	Estimate    QuoteEstimate `json:"estimate"`
	Tool        string        `json:"tool"`
	ToolDetails ToolDetails   `json:"toolDetails"`
}

type TransactionRequest struct {
	Value    string `json:"value"`
	To       string `json:"to"`
	Data     string `json:"data"`
	ChainID  int    `json:"chainId"`
	GasPrice string `json:"gasPrice"`
	GasLimit string `json:"gasLimit"`
	From     string `json:"from"`
}

type StatusRequest struct {
	TxHash    string
	Bridge    string
	FromChain int
	ToChain   int
}

func New(apiKey string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		apiKey:  strings.TrimSpace(apiKey),
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) GetChains(ctx context.Context) ([]Chain, error) {
	var response struct {
		Chains []Chain `json:"chains"`
	}
	if err := c.getJSON(ctx, c.baseURL+"/chains", &response); err != nil {
		return nil, err
	}
	return response.Chains, nil
}

func (c *Client) GetTools(ctx context.Context) (*ToolsResponse, error) {
	var response ToolsResponse
	if err := c.getJSON(ctx, c.baseURL+"/tools", &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetTokens(ctx context.Context, chainIDs []int, tags []string) (*TokensResponse, error) {
	params := url.Values{}
	if len(chainIDs) > 0 {
		parts := make([]string, 0, len(chainIDs))
		for _, chainID := range chainIDs {
			parts = append(parts, strconv.Itoa(chainID))
		}
		params.Set("chains", strings.Join(parts, ","))
	}
	if len(tags) > 0 {
		params.Set("tags", strings.Join(tags, ","))
	}

	endpoint := c.baseURL + "/tokens"
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	var response TokensResponse
	if err := c.getJSON(ctx, endpoint, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetQuote(ctx context.Context, request QuoteRequest) (*Quote, error) {
	params := url.Values{}
	params.Set("fromChain", strconv.Itoa(request.FromChain))
	params.Set("toChain", strconv.Itoa(request.ToChain))
	params.Set("fromToken", request.FromToken)
	params.Set("toToken", request.ToToken)
	params.Set("fromAddress", request.FromAddress)
	params.Set("toAddress", request.ToAddress)
	params.Set("fromAmount", request.FromAmount)
	if request.Slippage != "" {
		params.Set("slippage", request.Slippage)
	}
	if request.Preset != "" {
		params.Set("preset", request.Preset)
	}
	if len(request.AllowBridges) > 0 {
		params.Set("allowBridges", strings.Join(request.AllowBridges, ","))
	}
	if len(request.DenyBridges) > 0 {
		params.Set("denyBridges", strings.Join(request.DenyBridges, ","))
	}
	if len(request.AllowExchanges) > 0 {
		params.Set("allowExchanges", strings.Join(request.AllowExchanges, ","))
	}
	if len(request.DenyExchanges) > 0 {
		params.Set("denyExchanges", strings.Join(request.DenyExchanges, ","))
	}

	var quote Quote
	if err := c.getJSON(ctx, c.baseURL+"/quote?"+params.Encode(), &quote); err != nil {
		return nil, err
	}
	return &quote, nil
}

func (c *Client) GetStatus(ctx context.Context, request StatusRequest) (map[string]any, error) {
	params := url.Values{}
	params.Set("txHash", request.TxHash)
	if request.Bridge != "" {
		params.Set("bridge", request.Bridge)
	}
	if request.FromChain != 0 {
		params.Set("fromChain", strconv.Itoa(request.FromChain))
	}
	if request.ToChain != 0 {
		params.Set("toChain", strconv.Itoa(request.ToChain))
	}

	var payload map[string]any
	if err := c.getJSON(ctx, c.baseURL+"/status?"+params.Encode(), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, out any) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "lifi-cli/0.1")
		if c.apiKey != "" {
			req.Header.Set("x-lifi-api-key", c.apiKey)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			if shouldRetry(err, 0) && attempt < 2 {
				time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
				continue
			}
			return err
		}

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("li.fi api error: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
			if shouldRetry(lastErr, resp.StatusCode) && attempt < 2 {
				time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
				continue
			}
			return lastErr
		}

		decodeErr := json.NewDecoder(resp.Body).Decode(out)
		_ = resp.Body.Close()
		if decodeErr != nil {
			return decodeErr
		}
		return nil
	}
	return lastErr
}

func shouldRetry(err error, statusCode int) bool {
	if statusCode == http.StatusTooManyRequests || statusCode >= 500 {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}
