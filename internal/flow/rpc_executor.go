package flow

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/evm"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type RPCExecutor struct {
	client *ethclient.Client
}

func NewRPCExecutor(client *ethclient.Client) *RPCExecutor {
	return &RPCExecutor{client: client}
}

func (r *RPCExecutor) Close() {
	r.client.Close()
}

func (r *RPCExecutor) TokenBalance(ctx context.Context, tokenAddress string, owner common.Address) (*big.Int, error) {
	return evm.Balance(ctx, r.client, tokenAddress, owner)
}

func (r *RPCExecutor) NativeBalance(ctx context.Context, owner common.Address) (*big.Int, error) {
	return evm.Balance(ctx, r.client, "native", owner)
}

func (r *RPCExecutor) Allowance(ctx context.Context, tokenAddress string, owner, spender common.Address) (*big.Int, error) {
	return evm.Allowance(ctx, r.client, tokenAddress, owner, spender)
}

func (r *RPCExecutor) SimulateQuote(ctx context.Context, request lifiapi.TransactionRequest, from common.Address) (uint64, error) {
	return evm.SimulateQuoteTransaction(ctx, r.client, request, from)
}

func (r *RPCExecutor) EstimateQuoteFee(ctx context.Context, request lifiapi.TransactionRequest, gasPolicy string) (*evm.FeeEstimate, error) {
	from := common.HexToAddress(request.From)
	if request.From == "" {
		from = common.Address{}
	}
	return evm.EstimateQuoteFee(ctx, r.client, from, request, gasPolicy)
}

func (r *RPCExecutor) EstimateApprovalFee(ctx context.Context, wallet common.Address, chainID int, tokenAddress string, spender common.Address, amount *big.Int, gasPolicy string) (*evm.FeeEstimate, error) {
	return evm.EstimateApprovalFee(ctx, r.client, wallet, tokenAddress, spender, amount, gasPolicy)
}

func (r *RPCExecutor) Approve(ctx context.Context, wallet *evm.Wallet, chainID int, tokenAddress string, spender common.Address, amount *big.Int, gasPolicy string) (common.Hash, error) {
	return evm.Approve(ctx, r.client, wallet, chainID, tokenAddress, spender, amount, gasPolicy)
}

func (r *RPCExecutor) SendQuote(ctx context.Context, wallet *evm.Wallet, request lifiapi.TransactionRequest, gasPolicy string) (common.Hash, error) {
	return evm.SendQuoteTransaction(ctx, r.client, wallet, request, gasPolicy)
}

func (r *RPCExecutor) WaitForReceipt(ctx context.Context, txHash common.Hash, interval time.Duration) (*types.Receipt, error) {
	return evm.WaitForReceipt(ctx, r.client, txHash, interval)
}
