package evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

const erc20ABI = `[
	{"constant":true,"inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
	{"constant":false,"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},
	{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
]`

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

type FeeEstimate struct {
	GasLimit      uint64
	GasPrice      *big.Int
	GasTipCap     *big.Int
	GasFeeCap     *big.Int
	EstimatedCost *big.Int
	Dynamic       bool
}

type SendOptions struct {
	GasPolicy string
}

func WalletFromHex(privateKeyHex string) (*Wallet, error) {
	privateKeyHex = strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey: privateKey,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
	}, nil
}

func DialRPC(ctx context.Context, rpcURL string) (*ethclient.Client, error) {
	return ethclient.DialContext(ctx, rpcURL)
}

func IsNativeToken(address string) bool {
	address = strings.TrimSpace(strings.ToLower(address))
	return address == "" || address == "native" || address == "0x0000000000000000000000000000000000000000"
}

func Balance(ctx context.Context, client *ethclient.Client, tokenAddress string, owner common.Address) (*big.Int, error) {
	if IsNativeToken(tokenAddress) {
		return client.BalanceAt(ctx, owner, nil)
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, err
	}

	data, err := parsedABI.Pack("balanceOf", owner)
	if err != nil {
		return nil, err
	}

	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   addressPtr(tokenAddress),
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	values, err := parsedABI.Unpack("balanceOf", result)
	if err != nil {
		return nil, err
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("unexpected balanceOf result length")
	}

	balance, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected balanceOf result type")
	}

	return balance, nil
}

func Allowance(ctx context.Context, client *ethclient.Client, tokenAddress string, owner, spender common.Address) (*big.Int, error) {
	if IsNativeToken(tokenAddress) {
		return new(big.Int).SetUint64(^uint64(0)), nil
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, err
	}

	data, err := parsedABI.Pack("allowance", owner, spender)
	if err != nil {
		return nil, err
	}

	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   addressPtr(tokenAddress),
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	values, err := parsedABI.Unpack("allowance", result)
	if err != nil {
		return nil, err
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("unexpected allowance result length")
	}

	allowance, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected allowance result type")
	}

	return allowance, nil
}

func Approve(ctx context.Context, client *ethclient.Client, wallet *Wallet, chainID int, tokenAddress string, spender common.Address, amount *big.Int, gasPolicy string) (common.Hash, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return common.Hash{}, err
	}

	data, err := parsedABI.Pack("approve", spender, amount)
	if err != nil {
		return common.Hash{}, err
	}

	nonce, err := client.PendingNonceAt(ctx, wallet.Address)
	if err != nil {
		return common.Hash{}, err
	}

	token := common.HexToAddress(tokenAddress)
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: wallet.Address,
		To:   &token,
		Data: data,
	})
	if err != nil {
		gasLimit = 65000
	}

	fee, err := estimateFee(ctx, client, wallet.Address, &token, big.NewInt(0), data, gasLimit, nil, gasPolicy)
	if err != nil {
		return common.Hash{}, err
	}
	tx := buildTransaction(big.NewInt(int64(chainID)), nonce, &token, big.NewInt(0), data, fee)

	signed, err := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(int64(chainID))), wallet.PrivateKey)
	if err != nil {
		return common.Hash{}, err
	}

	if err := client.SendTransaction(ctx, signed); err != nil {
		return common.Hash{}, err
	}

	return signed.Hash(), nil
}

func SendQuoteTransaction(ctx context.Context, client *ethclient.Client, wallet *Wallet, request lifiapi.TransactionRequest, gasPolicy string) (common.Hash, error) {
	chainID := big.NewInt(int64(request.ChainID))
	nonce, err := client.PendingNonceAt(ctx, wallet.Address)
	if err != nil {
		return common.Hash{}, err
	}

	value, err := parseBigInt(request.Value)
	if err != nil {
		return common.Hash{}, fmt.Errorf("parse tx value: %w", err)
	}
	gasPrice, err := parseBigInt(request.GasPrice)
	if err != nil {
		return common.Hash{}, fmt.Errorf("parse gas price: %w", err)
	}
	gasLimit, err := parseUint64(request.GasLimit)
	if err != nil {
		return common.Hash{}, fmt.Errorf("parse gas limit: %w", err)
	}
	data := common.FromHex(request.Data)
	to := common.HexToAddress(request.To)

	fee, err := estimateFee(ctx, client, wallet.Address, &to, value, data, gasLimit, gasPrice, gasPolicy)
	if err != nil {
		return common.Hash{}, err
	}
	tx := buildTransaction(chainID, nonce, &to, value, data, fee)

	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), wallet.PrivateKey)
	if err != nil {
		return common.Hash{}, err
	}

	if err := client.SendTransaction(ctx, signed); err != nil {
		return common.Hash{}, err
	}

	return signed.Hash(), nil
}

func SimulateQuoteTransaction(ctx context.Context, client *ethclient.Client, request lifiapi.TransactionRequest, from common.Address) (uint64, error) {
	value, err := parseBigInt(request.Value)
	if err != nil {
		return 0, fmt.Errorf("parse tx value: %w", err)
	}
	to := common.HexToAddress(request.To)
	return client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: value,
		Data:  common.FromHex(request.Data),
	})
}

func EstimateQuoteFee(ctx context.Context, client *ethclient.Client, from common.Address, request lifiapi.TransactionRequest, gasPolicy string) (*FeeEstimate, error) {
	value, err := parseBigInt(request.Value)
	if err != nil {
		return nil, fmt.Errorf("parse tx value: %w", err)
	}
	gasPrice, err := parseBigInt(request.GasPrice)
	if err != nil {
		return nil, fmt.Errorf("parse gas price: %w", err)
	}
	gasLimit, err := parseUint64(request.GasLimit)
	if err != nil {
		return nil, fmt.Errorf("parse gas limit: %w", err)
	}
	to := common.HexToAddress(request.To)
	return estimateFee(ctx, client, from, &to, value, common.FromHex(request.Data), gasLimit, gasPrice, gasPolicy)
}

func EstimateApprovalFee(ctx context.Context, client *ethclient.Client, wallet common.Address, tokenAddress string, spender common.Address, amount *big.Int, gasPolicy string) (*FeeEstimate, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, err
	}

	data, err := parsedABI.Pack("approve", spender, amount)
	if err != nil {
		return nil, err
	}
	token := common.HexToAddress(tokenAddress)
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: wallet,
		To:   &token,
		Data: data,
	})
	if err != nil {
		gasLimit = 65000
	}
	return estimateFee(ctx, client, wallet, &token, big.NewInt(0), data, gasLimit, nil, gasPolicy)
}

func WaitForReceipt(ctx context.Context, client *ethclient.Client, txHash common.Hash, interval time.Duration) (*types.Receipt, error) {
	if interval <= 0 {
		interval = 3 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil && receipt != nil {
			return receipt, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func MaxApprovalAmount() *big.Int {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(256), nil)
	return max.Sub(max, big.NewInt(1))
}

func addressPtr(hexAddress string) *common.Address {
	address := common.HexToAddress(hexAddress)
	return &address
}

func parseBigInt(value string) (*big.Int, error) {
	value = strings.TrimSpace(value)
	base := 10
	if strings.HasPrefix(value, "0x") {
		base = 0
	}

	result := new(big.Int)
	if _, ok := result.SetString(value, base); !ok {
		return nil, fmt.Errorf("invalid big integer %q", value)
	}
	return result, nil
}

func parseUint64(value string) (uint64, error) {
	bigValue, err := parseBigInt(value)
	if err != nil {
		return 0, err
	}
	return bigValue.Uint64(), nil
}

func estimateFee(ctx context.Context, client *ethclient.Client, from common.Address, to *common.Address, value *big.Int, data []byte, requestedGasLimit uint64, requestedGasPrice *big.Int, gasPolicy string) (*FeeEstimate, error) {
	policy := strings.ToLower(strings.TrimSpace(gasPolicy))
	if policy == "" {
		policy = "auto"
	}

	gasLimit := requestedGasLimit
	if policy == "rpc" || policy == "auto" || gasLimit == 0 {
		estimatedGas, err := client.EstimateGas(ctx, ethereum.CallMsg{
			From:  from,
			To:    to,
			Value: value,
			Data:  data,
		})
		if err == nil && estimatedGas > gasLimit {
			gasLimit = estimatedGas
		}
	}
	if gasLimit == 0 {
		gasLimit = 21000
	}

	if policy != "quote" {
		header, err := client.HeaderByNumber(ctx, nil)
		if err == nil && header.BaseFee != nil {
			tipCap, tipErr := client.SuggestGasTipCap(ctx)
			if tipErr != nil || tipCap == nil || tipCap.Sign() <= 0 {
				tipCap = big.NewInt(2_000_000_000)
			}
			feeCap := new(big.Int).Add(new(big.Int).Mul(header.BaseFee, big.NewInt(2)), tipCap)
			if requestedGasPrice != nil && requestedGasPrice.Sign() > 0 && feeCap.Cmp(requestedGasPrice) < 0 {
				feeCap = new(big.Int).Set(requestedGasPrice)
			}
			return &FeeEstimate{
				GasLimit:      gasLimit,
				GasTipCap:     tipCap,
				GasFeeCap:     feeCap,
				EstimatedCost: new(big.Int).Mul(feeCap, new(big.Int).SetUint64(gasLimit)),
				Dynamic:       true,
			}, nil
		}
	}

	gasPrice := requestedGasPrice
	suggestedGasPrice, err := client.SuggestGasPrice(ctx)
	if err == nil && suggestedGasPrice != nil {
		switch policy {
		case "rpc":
			gasPrice = suggestedGasPrice
		case "auto":
			if gasPrice == nil || gasPrice.Sign() <= 0 || suggestedGasPrice.Cmp(gasPrice) > 0 {
				gasPrice = suggestedGasPrice
			}
		}
	}
	if gasPrice == nil || gasPrice.Sign() <= 0 {
		gasPrice = big.NewInt(0)
	}

	return &FeeEstimate{
		GasLimit:      gasLimit,
		GasPrice:      gasPrice,
		EstimatedCost: new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasLimit)),
	}, nil
}

func buildTransaction(chainID *big.Int, nonce uint64, to *common.Address, value *big.Int, data []byte, fee *FeeEstimate) *types.Transaction {
	if fee != nil && fee.Dynamic {
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        to,
			Value:     value,
			Gas:       fee.GasLimit,
			GasTipCap: fee.GasTipCap,
			GasFeeCap: fee.GasFeeCap,
			Data:      data,
		})
	}
	gasPrice := big.NewInt(0)
	if fee != nil && fee.GasPrice != nil {
		gasPrice = fee.GasPrice
	}
	gasLimit := uint64(0)
	if fee != nil {
		gasLimit = fee.GasLimit
	}
	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       to,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})
}
