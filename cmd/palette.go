package cmd

import (
	"github.com/fatih/color"
	"github.com/lesomnus/tiny-short/bybit"
)

var (
	h1         = color.New(color.FgHiWhite, color.Underline)
	h2         = color.New(color.FgHiWhite)
	p_good     = color.New(color.FgHiGreen)
	p_warn     = color.New(color.FgYellow)
	p_fail     = color.New(color.FgHiRed)
	p_fail_why = color.New(color.FgRed)
	p_dimmed   = color.New(color.Faint)
)

func pCoin(coin bybit.Coin) *color.Color {
	c := color.New(color.Bold)
	switch coin {
	case bybit.CoinSol:
		return c.Add(color.FgHiCyan)

	case bybit.CoinBtc:
		fallthrough
	default:
		return c.Add(color.FgHiYellow)
	}
}
