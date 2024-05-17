package bybit

import "context"

type TradeApi interface {
	OrderCreate(ctx context.Context, req TradeOrderCreateApiReq) (TradeOrderCreateApiRes, error)
	GetOrderHistory(ctx context.Context, req TradeGetOrderHistoryReq) (TradeGetOrderHistoryRes, error)
}

type TradeOrderCreateApiReq struct {
	Category  ProductType `json:"category"`
	Symbol    Symbol      `json:"symbol"`
	Side      OrderSide   `json:"side"`
	OrderType OrderType   `json:"orderType"`
	Quantity  string      `json:"qty"`
	Price     string      `json:"price,omitempty"` // Market order will ignore this field
}
type TradeOrderCreateApiRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		OrderId string `json:"orderId"`
	} `json:"result"`
}

type TradeGetOrderHistoryReq struct {
	Category ProductType `url:"category"`
	OrderId  string      `url:"orderId"`
	Limit    uint        `url:"limit"`
}
type TradeGetOrderHistoryRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		List []struct {
			OrderId  string `json:"orderId"`
			Price    Amount `json:"price"` // Order price.
			Qty      Amount `json:"qty"`
			AvgPrice Amount `json:"avgPrice"` // Average filled price.

			CreatedTime Timestamp `json:"createdTime"`
			UpdatedTime Timestamp `json:"updatedTime"`
		} `json:"list"`
	} `json:"result"`
}

type tradeApi struct {
	client *client
}

func (a *tradeApi) OrderCreate(ctx context.Context, req TradeOrderCreateApiReq) (res TradeOrderCreateApiRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/order/create")
	err = a.client.post(ctx, url, &req, &res)
	return
}

func (a *tradeApi) GetOrderHistory(ctx context.Context, req TradeGetOrderHistoryReq) (res TradeGetOrderHistoryRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/order/history")
	err = a.client.get(ctx, url, &req, &res)
	return
}
