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

func Approve(ctx context.Context, client *ethclient.Client, wallet *Wallet, chainID *big.Int, tokenAddress string, spender common.Address, amount *big.Int) (common.Hash, error) {
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return common.Hash{}, err
	}

	data, err := parsedABI.Pack("approve", spender, amount)
	if err != nil {
		return common.Hash{}, err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
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

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &token,
		Value:    big.NewInt(0),
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), wallet.PrivateKey)
	if err != nil {
		return common.Hash{}, err
	}

	if err := client.SendTransaction(ctx, signed); err != nil {
		return common.Hash{}, err
	}

	return signed.Hash(), nil
}

func SendQuoteTransaction(ctx context.Context, client *ethclient.Client, wallet *Wallet, request lifiapi.TransactionRequest) (common.Hash, error) {
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

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), wallet.PrivateKey)
	if err != nil {
		return common.Hash{}, err
	}

	if err := client.SendTransaction(ctx, signed); err != nil {
		return common.Hash{}, err
	}

	return signed.Hash(), nil
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
