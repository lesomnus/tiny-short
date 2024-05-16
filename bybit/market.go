package bybit

import "context"

type MarketApi interface {
	InstrumentsInfo(ctx context.Context, req MarketInstrumentsInfoReq) (MarketInstrumentsInfoRes, error)
	Tickers(ctx context.Context, req MarketTickersReq) (MarketTickersRes, error)
	FundingHistory(ctx context.Context, req MarketFundingHistoryReq) (MarketFundingHistoryRes, error)
}

type MarketInstrumentsInfoReq struct {
	Category ProductType `url:"category"`
	Symbol   Symbol      `url:"symbol"`
}
type MarketInstrumentsInfoRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		Category ProductType `json:"category"`
		List     []struct {
			ContractType  ContractType `json:"contractType"`
			PriceScale    string       `json:"priceScale"`
			LotSizeFilter struct {
				QtyStep Amount `json:"qtyStep"`
			} `json:"lotSizeFilter"`
		} `json:"list"`
	} `json:"result"`
}

type MarketTickersReq struct {
	Category ProductType `url:"category"`
	Symbol   Symbol      `url:"symbol"`
}
type MarketTickersRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		Category ProductType `json:"category"`
		List     []struct {
			Symbol      Symbol `json:"symbol"`
			MarkPrice   Amount `json:"markPrice"`
			FundingRate Amount `json:"fundingRate"`
			Bid1Price   Amount `json:"bid1Price"`
		} `json:"list"`
	} `json:"result"`
}

type MarketFundingHistoryReq struct {
	Category  ProductType `url:"category"`
	Symbol    Symbol      `url:"symbol"`
	StartTime Timestamp   `utl:"startTime"`
	EndTime   Timestamp   `url:"endTime"`
	Limit     uint        `url:"limit"`
}
type MarketFundingHistoryRes struct {
	ResponseBase `json:",inline"`

	Result struct {
		Category ProductType `json:"category"`
		List     []struct {
			Symbol               Symbol    `json:"symbol"`
			FundingRate          Amount    `json:"fundingRate"`
			FundingRateTimestamp Timestamp `json:"fundingRateTimestamp"`
		} `json:"list"`
	} `json:"result"`
}

type marketApi struct {
	client *client
}

func (a *marketApi) InstrumentsInfo(ctx context.Context, req MarketInstrumentsInfoReq) (res MarketInstrumentsInfoRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/market/instruments-info")
	err = a.client.get(ctx, url, &req, &res)
	return
}

func (a *marketApi) Tickers(ctx context.Context, req MarketTickersReq) (res MarketTickersRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/market/tickers")
	err = a.client.get(ctx, url, &req, &res)
	return
}

func (a *marketApi) FundingHistory(ctx context.Context, req MarketFundingHistoryReq) (res MarketFundingHistoryRes, err error) {
	url := a.client.conf.endpoint.Get("/v5/market/funding/history")
	err = a.client.get(ctx, url, &req, &res)
	return
}
