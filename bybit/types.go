package bybit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Timestamp time.Time

func (t Timestamp) Time() time.Time {
	return time.Time(t)
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	s := strconv.FormatInt(time.Time(t).UnixMilli(), 10)
	return []byte(s), nil
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var ts int64
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		ts = v
	} else if err := json.Unmarshal(data, &ts); err != nil {
		return err
	}

	sec := ts / 1000
	nsec := (ts % 1000) * int64(time.Millisecond)
	*t = Timestamp(time.Unix(sec, nsec))
	return nil
}

type UserId uint64

func (i UserId) String() string {
	return strconv.Itoa(int(i))
}

func (i *UserId) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}

		*i = UserId(v)
		return nil
	}

	var n uint64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}

	*i = UserId(n)
	return nil
}

type TransferId uuid.UUID

func (i *TransferId) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	id, err := uuid.Parse(s)
	if err != nil {
		return fmt.Errorf("parse UUID: %w", err)
	}

	*i = TransferId(id)
	return nil
}

func (i TransferId) MarshalJSON() ([]byte, error) {
	id := uuid.UUID(i)
	if id == uuid.Nil {
		var err error
		id, err = uuid.NewRandom()
		if err != nil {
			return nil, err
		}
	}

	return []byte(fmt.Sprintf(`"%s"`, id.String())), nil
}

type TransferStatus string

const (
	TransferStatusUnknown = TransferStatus("STATUS_UNKNOWN")
	TransferStatusSuccess = TransferStatus("SUCCESS")
	TransferStatusPending = TransferStatus("PENDING")
	TransferStatusFailed  = TransferStatus("FAILED")
)

type AccountType string

const (
	AccountTypeFund    = AccountType("FUND")
	AccountTypeUnified = AccountType("UNIFIED")
)

type ContractType string

const (
	ContractTypeInversePerpetual = ContractType("InversePerpetual")
	ContractTypeLinearPerpetual  = ContractType("LinearPerpetual")
	ContractTypeLinearFutures    = ContractType("LinearFutures")
	ContractTypeInverseFutures   = ContractType("InverseFutures")
)

type ProductType string

const (
	ProductTypeSpot    = ProductType("spot")
	ProductTypeLinear  = ProductType("linear")
	ProductTypeInverse = ProductType("inverse")
	ProductTypeOption  = ProductType("option")
)

type OrderSide string

const (
	OrderSideBuy  = OrderSide("Buy")
	OrderSideSell = OrderSide("Sell")
)

type OrderType string

const (
	OrderTypeMarket = OrderType("Market")
	OrderTypeLimit  = OrderType("Limit")
)

type Symbol string
type Coin string

const (
	CoinBtc = Coin("BTC")
	CoinSol = Coin("SOL")
)

func (c Coin) InvPerceptual() Symbol {
	return Symbol(fmt.Sprintf("%sUSD", c))
}

type ResponseBase struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Time    int64  `json:"time"`
}

func (r *ResponseBase) Ok() bool {
	return r.RetCode == RetCodeOk
}

func (r *ResponseBase) Err() error {
	return fmt.Errorf("%s (%d)", r.RetMsg, r.RetCode)
}

type Amount float64

func (a Amount) String() string {
	return a.StringPrec(-1)
}

func (a Amount) StringPrec(prec int) string {
	return strconv.FormatFloat(float64(a), 'f', prec, 64)
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*a = 0
		return nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	*a = Amount(f)
	return nil
}

type AccountInfo struct {
	UserId   UserId
	Nickname string
	Username string
	Secret   SecretRecord
}

func (i *AccountInfo) DisplayName() string {
	if i.Nickname != "" {
		return i.Nickname
	}
	return i.Username
}

func (i *AccountInfo) DisplayNameTrunc(l int) string {
	name := i.DisplayName()
	nl := len(name)
	if nl <= l {
		return name
	}

	return fmt.Sprintf("%s..", name[0:l-2])
}

type ApiPermissions struct {
	ContractTrade []string `json:"ContractTrade"`
	Wallet        []string `json:"Wallet"`
	Derivatives   []string `json:"Derivatives"`
}
