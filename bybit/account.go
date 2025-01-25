package bybit

import (
	"context"
)

type AccountApi interface {
	WalletBalance(ctx context.Context, req AccountWalletBalanceReq) (AccountWalletBalanceRes, error)
	TransferableAmount(ctx context.Context, req AccountTransferableAmountReq) (AccountTransferableAmountRes, error)
}

type AccountWalletBalanceReq struct {
	AccountType AccountType `url:"accountType"`
	Coin        Coin        `url:"coin"`
}
type AccountWalletBalanceRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		List []struct {
			AccountType AccountType `json:"accountType"`
			Coin        []struct {
				Coin   Coin   `json:"coin"`
				Equity Amount `json:"equity"`
				// UsdValue      Amount `json:"usdValue"`
				UnrealisedPnl Amount `json:"unrealisedPnl"`
			}
		} `json:"list"`
	} `json:"result"`
}

type AccountTransferableAmountReq struct {
	CoinName Coin `url:"coinName"`
}
type AccountTransferableAmountRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		AvailableWithdrawal Amount `json:"availableWithdrawal"`
	} `json:"result"`
}

type accountApi struct {
	client *client
}

func (a *accountApi) WalletBalance(ctx context.Context, req AccountWalletBalanceReq) (res AccountWalletBalanceRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/account/wallet-balance")
	err = a.client.get(ctx, url, &req, &res)
	return
}

func (a *accountApi) TransferableAmount(ctx context.Context, req AccountTransferableAmountReq) (res AccountTransferableAmountRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/account/withdrawal")
	err = a.client.get(ctx, url, &req, &res)
	return
}
