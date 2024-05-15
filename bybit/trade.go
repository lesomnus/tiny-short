package bybit

import "context"

type TradeApi interface {
	OrderCreate(ctx context.Context, req TradeOrderCreateApiReq) (TradeOrderCreateApiRes, error)
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

type tradeApi struct {
	client *client
}

func (a *tradeApi) OrderCreate(ctx context.Context, req TradeOrderCreateApiReq) (res TradeOrderCreateApiRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/order/create")
	err = a.client.post(ctx, url, &req, &res)
	return
}
