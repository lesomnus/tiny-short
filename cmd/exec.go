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

type Exec struct {
	Client  bybit.Client
	Move    MoveConfig
	Debug   DebugConfig
	Secrets bybit.SecretStore
}

func (e *Exec) Do(ctx context.Context, coin bybit.Coin) error {
	l := log.From(ctx)
	to := e.Move.to
	pC := pCoin(coin)

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

	var bid bybit.Amount
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
		bid = ticker.Bid1Price

		p_dimmed.Print("⚡")
		fmt.Printf("%s%%", (ticker.FundingRate * 100).String())
		p_dimmed.Print(" M")
		fmt.Print(ticker.MarkPrice)
		p_dimmed.Print(" B")
		fmt.Print(bid, "\n")
	}

	fmt.Println("")

	// 177162484 Bybitplg2HTZRgxP........0.01084342/110.80237644
	h2.Printf("%9s ", to.UserId)
	p_dimmed.Printf("%s", to.Username)
	p_dimmed.Print(strings.Repeat(".", 24-len(to.Username)))
	if res, err := e.Client.Asset().QueryAccountCoinBalance(ctx, bybit.AssetQueryAccountCoinBalanceReq{
		MemberId:    to.UserId.String(),
		AccountType: bybit.AccountTypeContract,
		Coin:        coin,
	}); err != nil {
		return fmt.Errorf("request for query account coin balance: %w", err)
	} else if !res.Ok() {
		return fmt.Errorf("query account coin balance: %w", res.Err())
	} else {
		b := res.Result.Balance
		pC.Print(b.TransferBalance)
		fmt.Print("/")
		pC.Println(b.WalletBalance)
	}

	for _, from := range e.Move.from {
		var balance bybit.Amount
		if res, err := e.Client.Asset().QueryAccountCoinBalance(ctx, bybit.AssetQueryAccountCoinBalanceReq{
			MemberId:      from.UserId.String(),
			ToMemberId:    to.UserId.String(),
			AccountType:   bybit.AccountTypeContract,
			ToAccountType: bybit.AccountTypeContract,
			Coin:          coin,
		}); err != nil {
			return fmt.Errorf("request for query account coin balance: %w", err)
		} else if !res.Ok() {
			return fmt.Errorf("query account coin balance: %w", res.Err())
		} else {
			balance = res.Result.Balance.TransferBalance
		}

		//  ↖ 178225536 BybitydINQ0FotFH.....0.00698261 ✗ IGNORE amount too small
		fmt.Print(" ⮤ ")
		h2.Printf("%9s ", from.UserId)
		p_dimmed.Printf("%s", from.Username)
		p_dimmed.Print(strings.Repeat(".", 21-len(from.Username)))
		pC.Print(balance)

		if balance == 0 {
			fmt.Println(" = SKIP")
			continue
		}

		if e.Debug.Enabled && (e.Debug.SkipTransaction || e.Debug.SkipTransfer) {
			p_warn.Print(" = SKIP ")
			p_dimmed.Println("by config")
		} else if res, err := e.Client.Asset().UniversalTransfer(ctx, bybit.AssetUniversalTransferReq{
			Coin: coin,
			// Amount: "0.00000000001",
			// Amount:          "0.0001",
			Amount:          balance.String(),
			FromMember:      from.UserId,
			ToMember:        to.UserId,
			FromAccountType: bybit.AccountTypeContract,
			ToAccountType:   bybit.AccountTypeContract,
		}); err != nil {
			p_fail.Print(" ✗ REQ FAILED ")
			p_fail_why.Println(err.Error())
			return fmt.Errorf("request for asset transfer: %w", err)
		} else if !res.Ok() {
			switch res.RetCode {
			case bybit.RetCodeUnacceptableAmountAccuracy:
				p_warn.Print(" ✗ IGNORE ")
				p_dimmed.Println("amount too small")
				continue
			default:
				p_fail.Print(" ✗ ABORTED ")
				p_fail_why.Println(res.RetMsg)
				return fmt.Errorf("asset transfer: %w", res.Err())
			}
		} else if res.Result.Status != bybit.TransferStatusSuccess {
			switch res.Result.Status {
			case bybit.TransferStatusUnknown:
				fmt.Print(" ? UNKNOWN ")
				p_dimmed.Println(res.RetMsg)
			case bybit.TransferStatusPending:
				fmt.Print(" ~ PENDING ")
				p_dimmed.Println(res.RetMsg)
			case bybit.TransferStatusFailed:
				p_fail.Print(" ✗ FAILED ")
				p_fail_why.Println(res.RetMsg)
			default:
				p_fail.Print(" ? UNSUPPORTED ")
				p_fail_why.Println("unknown status: ", res.Result.Status)
			}
			return fmt.Errorf("transfer not succeed: %s", res.Result.Status)
		} else {
			p_good.Println(" ✓ SUCCESS")
		}
	}

	var balance bybit.Amount
	p_dimmed.Println("                                  ----------")
	if res, err := e.Client.Asset().QueryAccountCoinBalance(ctx, bybit.AssetQueryAccountCoinBalanceReq{
		MemberId:    to.UserId.String(),
		AccountType: bybit.AccountTypeContract,
		Coin:        coin,
	}); err != nil {
		return fmt.Errorf("request for query account coin balance: %w", err)
	} else if !res.Ok() {
		return fmt.Errorf("query account coin balance: %w", res.Err())
	} else {
		b := res.Result.Balance
		balance = b.TransferBalance

		fmt.Print("Short by market order             ")
		pC.Print(balance)
		fmt.Print("/")
		pC.Println(b.WalletBalance)
	}

	trading_client := e.Client.Clone(e.Move.to.Secret)

	qty := int(balance * bid * (1 - bybit.FeePerpTake))
	fmt.Print("Qty ")
	h2.Printf("%d ", qty)

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
		p_good.Println("✓ SUCCESS")
		order_id = res.Result.OrderId
	}

	if order_id == "" {
		// Do nothing.
		// Anyway, code does reach here if there is no order made.
	} else if res, err := trading_client.Trade().GetOrderHistory(ctx, bybit.TradeGetOrderHistoryReq{
		Category: bybit.ProductTypeInverse,
		OrderId:  order_id,
		Limit:    1,
	}); err != nil {
		p_warn.Print("failed to get order details ")
		p_dimmed.Println(err.Error())
		l.Warn("request for get order history", slog.String("err", err.Error()))
	} else if !res.Ok() {
		p_warn.Print("failed to get order details ")
		p_dimmed.Println(res.Err())
		l.Warn("get order history", slog.String("err", res.Err().Error()))
	} else if len(res.Result.List) == 0 {
		p_warn.Print("failed to get order details ")
		p_dimmed.Println("order list empty")
		l.Warn("order list empty")
	} else if res.Result.List[0].OrderId != order_id {
		p_warn.Print("failed to get order details ")
		p_dimmed.Println("different order ID")
		l.Warn("different order ID")
	} else {
		order := res.Result.List[0]
		fmt.Print(" 🔔 ")
		h2.Print(order.Qty.String())
		if order.Qty == 1 {
			fmt.Print(" contract was")
		} else {
			fmt.Print(" contracts were")
		}
		fmt.Print(" sold at the price of ")
		h2.Println(order.Price.String())
	}

	return nil
}
