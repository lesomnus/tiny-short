package bybit

import "context"

type TradeApi interface {
	OrderCreate(ctx context.Context, req TradeOrderCreateApiReq) (TradeOrderCreateApiRes, error)
	OrderHistory(ctx context.Context, req TradeOrderHistoryReq) (TradeOrderHistoryRes, error)
	ExecutionList(ctx context.Context, req TradeExecutionListReq) (TradeExecutionListRes, error)
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

type TradeOrderHistoryReq struct {
	Category ProductType `url:"category"`
	OrderId  string      `url:"orderId"`
	Limit    uint        `url:"limit"`
}
type TradeOrderHistoryRes struct {
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

type TradeExecutionListReq struct {
	Category ProductType `url:"category"`
	OrderId  string      `url:"orderId"`
	Limit    uint        `url:"limit"`
}
type TradeExecutionListRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		Category ProductType `json:"category"`
		List     []struct {
			OrderId   string    `json:"orderId"`
			ExecPrice Amount    `json:"execPrice"`
			ExecQty   Amount    `json:"execQty"`
			ExecTime  Timestamp `json:"execTime"`
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

func (a *tradeApi) OrderHistory(ctx context.Context, req TradeOrderHistoryReq) (res TradeOrderHistoryRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/order/history")
	err = a.client.get(ctx, url, &req, &res)
	return
}

func (a *tradeApi) ExecutionList(ctx context.Context, req TradeExecutionListReq) (res TradeExecutionListRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/execution/list")
	err = a.client.get(ctx, url, &req, &res)
	return
}
