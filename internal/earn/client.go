package earn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://earn.li.fi/v1/earn"

type Client struct {
	baseURL string
	http    *http.Client
}

type Chain struct {
	ChainID     int    `json:"chainId"`
	Name        string `json:"name"`
	NetworkCAIP string `json:"networkCaip"`
}

type Protocol struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Vault struct {
	Name             string            `json:"name"`
	Slug             string            `json:"slug"`
	Tags             []string          `json:"tags"`
	Address          string            `json:"address"`
	ChainID          int               `json:"chainId"`
	Network          string            `json:"network"`
	Protocol         Protocol          `json:"protocol"`
	Provider         string            `json:"provider"`
	SyncedAt         string            `json:"syncedAt"`
	Analytics        VaultAnalytics    `json:"analytics"`
	Description      string            `json:"description"`
	RedeemPacks      []VaultPack       `json:"redeemPacks"`
	DepositPacks     []VaultPack       `json:"depositPacks"`
	IsRedeemable     bool              `json:"isRedeemable"`
	IsTransactional  bool              `json:"isTransactional"`
	UnderlyingTokens []UnderlyingToken `json:"underlyingTokens"`
	LPTokens         []UnderlyingToken `json:"lpTokens"`
}

type VaultAnalytics struct {
	APY       APYBreakdown `json:"apy"`
	TVL       TVLBreakdown `json:"tvl"`
	APY1d     *float64     `json:"apy1d"`
	APY7d     *float64     `json:"apy7d"`
	APY30d    *float64     `json:"apy30d"`
	UpdatedAt string       `json:"updatedAt"`
}

type APYBreakdown struct {
	Base   float64 `json:"base"`
	Reward float64 `json:"reward"`
	Total  float64 `json:"total"`
}

type TVLBreakdown struct {
	USD string `json:"usd"`
}

type VaultPack struct {
	Name      string `json:"name"`
	StepsType string `json:"stepsType"`
}

type UnderlyingToken struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}

type VaultQuery struct {
	ChainID   int
	Asset     string
	SortBy    string
	MinTvlUSD string
	Limit     int
	Cursor    string
}

type VaultPage struct {
	Data       []Vault `json:"data"`
	NextCursor string  `json:"nextCursor"`
	Total      int     `json:"total"`
}

type PortfolioResponse struct {
	Positions []map[string]any `json:"positions"`
}

func New() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) GetChains(ctx context.Context) ([]Chain, error) {
	var chains []Chain
	if err := c.getJSON(ctx, c.baseURL+"/chains", &chains); err != nil {
		return nil, err
	}
	return chains, nil
}

func (c *Client) GetProtocols(ctx context.Context) ([]Protocol, error) {
	var protocols []Protocol
	if err := c.getJSON(ctx, c.baseURL+"/protocols", &protocols); err != nil {
		return nil, err
	}
	return protocols, nil
}

func (c *Client) GetVaults(ctx context.Context, query VaultQuery) (*VaultPage, error) {
	params := url.Values{}
	if query.ChainID != 0 {
		params.Set("chainId", strconv.Itoa(query.ChainID))
	}
	if strings.TrimSpace(query.Asset) != "" {
		params.Set("asset", query.Asset)
	}
	if strings.TrimSpace(query.SortBy) != "" {
		params.Set("sortBy", query.SortBy)
	}
	if strings.TrimSpace(query.MinTvlUSD) != "" {
		params.Set("minTvlUsd", query.MinTvlUSD)
	}
	if query.Limit > 0 {
		params.Set("limit", strconv.Itoa(query.Limit))
	}
	if strings.TrimSpace(query.Cursor) != "" {
		params.Set("cursor", query.Cursor)
	}

	endpoint := c.baseURL + "/vaults"
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	var page VaultPage
	if err := c.getJSON(ctx, endpoint, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) GetAllVaults(ctx context.Context, query VaultQuery) ([]Vault, error) {
	pageSize := query.Limit
	if pageSize <= 0 {
		pageSize = 100
	}

	cursor := query.Cursor
	all := make([]Vault, 0, pageSize)
	for {
		page, err := c.GetVaults(ctx, VaultQuery{
			ChainID:   query.ChainID,
			Asset:     query.Asset,
			SortBy:    query.SortBy,
			MinTvlUSD: query.MinTvlUSD,
			Limit:     pageSize,
			Cursor:    cursor,
		})
		if err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}

	return all, nil
}

func (c *Client) GetPortfolio(ctx context.Context, address string) (*PortfolioResponse, error) {
	var response PortfolioResponse
	endpoint := fmt.Sprintf("%s/portfolio/%s/positions", c.baseURL, address)
	if err := c.getJSON(ctx, endpoint, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "lifi-cli/0.1")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("earn api error: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
