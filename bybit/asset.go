package bybit

import (
	"context"
)

type AssetApi interface {
	QueryAccountCoinBalance(ctx context.Context, req AssetQueryAccountCoinBalanceReq) (AssetQueryAccountCoinBalanceRes, error)
	InterTransfer(ctx context.Context, req AssetInterTransferReq) (AssetInterTransferRes, error)
	UniversalTransfer(ctx context.Context, req AssetUniversalTransferReq) (AssetUniversalTransferRes, error)
}

type AssetQueryAccountCoinBalanceReq struct {
	MemberId      string      `url:"memberId,omitempty"`
	ToMemberId    string      `url:"toMemberId,omitempty"`
	AccountType   AccountType `url:"accountType"`
	ToAccountType AccountType `url:"toAccountType,omitempty"`

	Coin Coin `url:"coin"`
}
type AssetQueryAccountCoinBalanceRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		AccountType AccountType `json:"accountType"`
		AccountId   string      `json:"accountId"`
		MemberId    string      `json:"memberId"`
		Balance     struct {
			Coin            Coin   `json:"coin"`
			WalletBalance   Amount `json:"walletBalance"`
			TransferBalance Amount `json:"transferBalance"`
		} `json:"balance"`
	} `json:"result"`
}

type AssetInterTransferReq struct {
	TransferId TransferId `json:"transferId"`

	Coin            Coin        `json:"coin"`
	Amount          Amount      `json:"amount"`
	FromAccountType AccountType `json:"fromAccountType"`
	ToAccountType   AccountType `json:"toAccountType"`
}
type AssetInterTransferRes struct {
}

type AssetUniversalTransferReq struct {
	TransferId TransferId `json:"transferId"`

	Coin            Coin        `json:"coin"`
	Amount          string      `json:"amount"`
	FromMember      UserId      `json:"fromMemberId"`
	ToMember        UserId      `json:"toMemberId"`
	FromAccountType AccountType `json:"fromAccountType"`
	ToAccountType   AccountType `json:"toAccountType"`
}

type AssetUniversalTransferRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		TransferId TransferId     `json:"transferId"`
		Status     TransferStatus `json:"status"`
	} `json:"result"`
}

type assetApi struct {
	client *client
}

func (a *assetApi) QueryAccountCoinBalance(ctx context.Context, req AssetQueryAccountCoinBalanceReq) (res AssetQueryAccountCoinBalanceRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/asset/transfer/query-account-coin-balance")
	err = a.client.get(ctx, url, &req, &res)
	return
}

func (a *assetApi) InterTransfer(ctx context.Context, req AssetInterTransferReq) (res AssetInterTransferRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/asset/transfer/inter-transfer")
	err = a.client.post(ctx, url, &req, &res)
	return
}

func (a *assetApi) UniversalTransfer(ctx context.Context, req AssetUniversalTransferReq) (res AssetUniversalTransferRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/asset/transfer/universal-transfer")
	err = a.client.post(ctx, url, &req, &res)
	return
}
