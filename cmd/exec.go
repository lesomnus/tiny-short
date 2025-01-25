package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lesomnus/tiny-short/bybit"
	"github.com/lesomnus/tiny-short/log"
)

type TransferPlan struct {
	Users []bybit.AccountInfo // [to, from...]
}

func (p *TransferPlan) Dest() *bybit.AccountInfo {
	return &p.Users[0]
}

func (p *TransferPlan) Source() []bybit.AccountInfo {
	return p.Users[1:]
}

type Exec struct {
	Client bybit.Client

	TransferPlan TransferPlan
	Secrets      bybit.SecretStore

	Debug DebugConfig
}

func (e *Exec) Do(ctx context.Context, coin bybit.Coin) error {
	l := log.From(ctx)
	p_coin := pCoin(coin)

	fmt.Printf("\n----------------\n")
	color.New(color.BgMagenta, color.FgHiWhite).Print(" SHORT ")
	fmt.Print(" ")
	pCoin(coin).Add(color.Underline).Printf("%s", coin)

	if res, err := e.Client.Market().FundingHistory(ctx, bybit.MarketFundingHistoryReq{
		Category:  bybit.ProductTypeInverse,
		Symbol:    coin.InvPerceptual(),
		StartTime: bybit.Timestamp(time.Now().Add(-8 * time.Hour)),
		EndTime:   bybit.Timestamp(time.Now()),
		Limit:     1,
	}); err != nil {
		l.Warn("request for funding history", slog.String("err", err.Error()))
	} else if !res.Ok() {
		l.Warn("funding history", slog.String("err", res.Err().Error()))
	} else if len(res.Result.List) == 0 {
		l.Warn("funding history empty")
	} else {
		history := res.Result.List[0]
		fmt.Print("⚡")
		color.New(color.FgHiYellow).Printf("%s%% ", (history.FundingRate * 100).String())
		fmt.Print(time.Since(history.FundingRateTimestamp.Time()).Truncate(time.Second))
		p_dimmed.Print(" ago ")
		fmt.Print("|")
	}

	var (
		mark_price bybit.Amount
		bid1_price bybit.Amount
	)
	if res, err := e.Client.Market().Tickers(ctx, bybit.MarketTickersReq{
		Category: bybit.ProductTypeInverse,
		Symbol:   coin.InvPerceptual(),
	}); err != nil {
		return fmt.Errorf("request for tickers: %w", err)
	} else if !res.Ok() {
		return fmt.Errorf("tickers: %w", res.Err())
	} else if len(res.Result.List) == 0 {
		return fmt.Errorf("tickers empty: %w", res.Err())
	} else {
		ticker := res.Result.List[0]
		mark_price = ticker.MarkPrice
		bid1_price = ticker.Bid1Price

		p_dimmed.Print("⚡")
		fmt.Printf("%s%%", (ticker.FundingRate * 100).String())
		p_dimmed.Print(" M")
		fmt.Print(mark_price)
		p_dimmed.Print(" B")
		fmt.Print(bid1_price, "\n")
	}

	fmt.Println()

	dst := e.TransferPlan.Dest()

	// nickname......0.01084342 ≈ 42 USD
	//  + nickname...0.01084342 ≈ 42 USD
	{
		name := dst.DisplayNameTrunc(8)
		h2.Print(name)
		p_dimmed.Print(strings.Repeat(".", (3+8+3)-len(name)))
	}
	if res, err := e.Client.Asset().QueryAccountCoinBalance(ctx, bybit.AssetQueryAccountCoinBalanceReq{
		MemberId:    dst.UserId.String(),
		AccountType: bybit.AccountTypeUnified,
		Coin:        coin,
	}); err != nil {
		return fmt.Errorf("request for query account coin balance: %w", err)
	} else if !res.Ok() {
		return fmt.Errorf("query account coin balance: %w", res.Err())
	} else {
		b := res.Result.Balance.TransferBalance
		p_coin.Printf("%8f", b)
		p_dimmed.Printf(" ≈ %8f USD\n", b*mark_price)
	}

	for _, src := range e.TransferPlan.Source() {
		{
			name := src.DisplayNameTrunc(8)
			fmt.Print(" + ")
			h2.Print(name)
			p_dimmed.Print(strings.Repeat(".", (8+3)-len(name)))
		}

		var balance bybit.Amount
		if res, err := e.Client.Asset().QueryAccountCoinBalance(ctx, bybit.AssetQueryAccountCoinBalanceReq{
			MemberId:      src.UserId.String(),
			ToMemberId:    dst.UserId.String(),
			AccountType:   bybit.AccountTypeUnified,
			ToAccountType: bybit.AccountTypeUnified,
			Coin:          coin,
		}); err != nil {
			return fmt.Errorf("request for query account coin balance: %w", err)
		} else if !res.Ok() {
			return fmt.Errorf("query account coin balance: %w", res.Err())
		} else {
			balance = res.Result.Balance.TransferBalance
		}

		p_coin.Printf("%8f", balance)
		p_dimmed.Printf(" ≈ %8f USD ", balance*mark_price)

		if balance == 0 {
			fmt.Println("= SKIP")
			continue
		}

		if e.Debug.Enabled && (e.Debug.SkipTransaction || e.Debug.SkipTransfer) {
			p_warn.Print("= SKIP ")
			p_dimmed.Println("by config")
		} else if res, err := e.Client.Asset().UniversalTransfer(ctx, bybit.AssetUniversalTransferReq{
			Coin:            coin,
			Amount:          balance.String(),
			FromMember:      src.UserId,
			ToMember:        dst.UserId,
			FromAccountType: bybit.AccountTypeUnified,
			ToAccountType:   bybit.AccountTypeUnified,
		}); err != nil {
			p_fail.Print("✗ REQ FAILED ")
			p_fail_why.Println(err.Error())
			return fmt.Errorf("request for asset transfer: %w", err)
		} else if !res.Ok() {
			switch res.RetCode {
			case bybit.RetCodeUnacceptableAmountAccuracy:
				p_warn.Print("✗ IGNORE ")
				p_dimmed.Println("amount too small")
				continue
			default:
				p_fail.Print("✗ ABORTED ")
				p_fail_why.Println(res.RetMsg)
				return fmt.Errorf("asset transfer: %w", res.Err())
			}
		} else if res.Result.Status != bybit.TransferStatusSuccess {
			switch res.Result.Status {
			case bybit.TransferStatusUnknown:
				fmt.Print("? UNKNOWN ")
				p_dimmed.Println(res.RetMsg)
			case bybit.TransferStatusPending:
				fmt.Print("~ PENDING ")
				p_dimmed.Println(res.RetMsg)
			case bybit.TransferStatusFailed:
				p_fail.Print("✗ FAILED ")
				p_fail_why.Println(res.RetMsg)
			default:
				p_fail.Print("? UNSUPPORTED ")
				p_fail_why.Println("unknown status: ", res.Result.Status)
			}
			return fmt.Errorf("transfer not succeed: %s", res.Result.Status)
		} else {
			p_good.Println("✓ SUCCESS")
		}
	}

	trading_client := e.Client.Clone(e.TransferPlan.Dest().Secret)

	var balance bybit.Amount
	p_dimmed.Println("              ----------")
	if res, err := trading_client.Account().TransferableAmount(ctx, bybit.AccountTransferableAmountReq{
		CoinName: coin,
	}); err != nil {
		return fmt.Errorf("request for wallet balance: %w", err)
	} else if !res.Ok() {
		return fmt.Errorf("wallet balance: %w", res.Err())
	} else {
		balance = res.Result.AvailableWithdrawal

		p_dimmed.Print("              ")
		p_coin.Printf("%8f", balance)
		p_dimmed.Printf(" ≈ %8f USD\n", balance*mark_price)
	}

	fmt.Printf("\nShort by market order\n")

	qty := int(balance * bid1_price * (1 - bybit.FeePerpTake))
	fmt.Print("Places ")
	h2.Print(qty)
	if qty == 1 {
		h2.Print(" contract ")
	} else {
		h2.Print(" contracts ")
	}

	if qty == 0 {
		fmt.Println("= SKIP")
		return nil
	}
	if e.Debug.Enabled && e.Debug.SkipTransaction {
		p_warn.Print("= SKIP ")
		p_dimmed.Println("by config")
		return nil
	}

	order_id := ""
	if e.Debug.Enabled && e.Debug.SkipTransaction {
		// Do NOT remove this block to prevent mistake.
		fmt.Println()
	} else if res, err := trading_client.Trade().OrderCreate(ctx, bybit.TradeOrderCreateApiReq{
		Category:  bybit.ProductTypeInverse,
		Symbol:    coin.InvPerceptual(),
		Side:      bybit.OrderSideSell,
		OrderType: bybit.OrderTypeMarket,
		Quantity:  strconv.Itoa(qty),
	}); err != nil {
		p_fail.Print("✗ REQ FAILED ")
		p_fail_why.Println(err.Error())
		return fmt.Errorf("request for order create: %w", err)
	} else if !res.Ok() {
		p_fail.Print("✗ FAILED ")
		p_fail_why.Println(res.RetMsg)
		return fmt.Errorf("order create: %w", res.Err())
	} else {
		order_id = res.Result.OrderId

		p_good.Print("✓ SUCCESS ")
		p_dimmed.Println(order_id)
	}

	if order_id == "" {
		// Do nothing.
		// Anyway, code does not reach here if there is no order made.
	} else {
		//        "Places N contracts ..."
		fmt.Print("     ↳ ")

		// Wait for the trading system closes the order.
		// Note that `Tarde.GetOrderHistory` only queries closed orders.
		time.Sleep(3 * time.Second)

		RetryCount := 3
		i := 0
		for ; i < RetryCount; i++ {
			res, err := trading_client.Trade().OrderHistory(ctx, bybit.TradeOrderHistoryReq{
				Category: bybit.ProductTypeInverse,
				OrderId:  order_id,
				Limit:    1,
			})
			if err != nil {
				p_warn.Print("failed to get order details ")
				p_dimmed.Println(err.Error())
				l.Warn("request for get order history", slog.String("err", err.Error()))
				break
			}
			if !res.Ok() {
				p_warn.Print("failed to get order details ")
				p_dimmed.Println(res.Err())
				l.Warn("get order history", slog.String("err", res.Err().Error()))
				break
			}
			if len(res.Result.List) == 0 {
				// Order not yes closed?
				continue
			}
			if res.Result.List[0].OrderId != order_id {
				p_warn.Print("failed to get order details ")
				p_dimmed.Println("different order ID")
				l.Warn("different order ID")
				break
			}

			order := res.Result.List[0]
			h2.Print(qty)
			if order.Qty == 1 {
				h2.Print(" contract ")
				fmt.Print("was")
			} else {
				h2.Print(" contracts ")
				fmt.Print("were")
			}
			fmt.Print(" sold at the price of ")
			h2.Printf("%s USD\n", order.AvgPrice.String())

			//              "Places N contracts ..."
			p_dimmed.Printf("       %s\n", order.UpdatedTime.Time())
			break
		}

		if i == RetryCount {
			p_warn.Print("failed to get order details ")
			p_dimmed.Println("order does not closed")
			l.Warn("order does not closed")
		}
	}

	return nil
}
