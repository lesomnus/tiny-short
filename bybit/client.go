package bybit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/lesomnus/tiny-short/log"
)

type Client interface {
	User() UserApi
	Account() AccountApi
	Asset() AssetApi
	Market() MarketApi
	Trade() TradeApi

	Clone(secret SecretRecord) Client
}

type client struct {
	conf   clientConfig
	secret SecretRecord
}

type clientConfig struct {
	endpoint Endpoint
}

type ClientOption = func(c *clientConfig)

func WithNetwork(u url.URL) ClientOption {
	u.Path = ""
	return func(c *clientConfig) {
		c.endpoint = Endpoint(u)
	}
}

func NewClient(secret SecretRecord, opts ...ClientOption) Client {
	testnet, err := url.Parse(TestNetAddr1)
	if err != nil {
		panic("invalid URL of test net")
	}

	c := clientConfig{
		endpoint: Endpoint(*testnet),
	}
	for _, opt := range opts {
		opt(&c)
	}

	return &client{
		conf:   c,
		secret: secret,
	}
}

func (c *client) Clone(secret SecretRecord) Client {
	c_ := *c
	c_.secret = secret
	return &c_
}

func (c *client) User() UserApi {
	return &userApi{client: c}
}

func (c *client) Account() AccountApi {
	return &accountApi{client: c}
}

func (c *client) Asset() AssetApi {
	return &assetApi{client: c}
}

func (c *client) Market() MarketApi {
	return &marketApi{client: c}
}

func (c *client) Trade() TradeApi {
	return &tradeApi{client: c}
}

func (c *client) makeReq(ctx context.Context, method string, url string, data []byte) (*http.Request, error) {
	now := time.Now()
	ts := now.UTC().UnixMilli()
	ts_str := strconv.FormatInt(ts, 10)

	var w bytes.Buffer
	w.Write([]byte(ts_str))
	w.Write([]byte(c.secret.ApiKey))
	w.Write([]byte("5000"))
	w.Write(data)

	payload := w.Bytes()
	signature, err := c.secret.Sign(payload)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	var body io.Reader = nil
	if method == http.MethodPost {
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-API-KEY", c.secret.ApiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", ts_str)
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("X-BAPI-RECV-WINDOW", "5000")

	return req, nil
}

func (c *client) get(ctx context.Context, url string, req any, res any) error {
	vs, err := query.Values(req)
	if err != nil {
		panic("invalid req")
	}

	q := vs.Encode()
	url = fmt.Sprintf("%s?%s", url, q)

	l := log.From(ctx)
	l.Info("REQ->", slog.String("method", "GET "), slog.String("data", q))
	return c.exec(ctx, http.MethodGet, url, []byte(q), res)
}

func (c *client) post(ctx context.Context, url string, req any, res any) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	l := log.From(ctx)
	l.Info("REQ->", slog.String("method", "POST"), slog.String("data", string(data)))
	return c.exec(ctx, http.MethodPost, url, data, res)
}

func (c *client) exec(ctx context.Context, method string, url string, data []byte, res any) error {
	req, err := c.makeReq(ctx, method, url, data)
	if err != nil {
		return fmt.Errorf("make req: %w", err)
	}

	res_, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("roundtrip: %w", err)
	}

	body, err := io.ReadAll(res_.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	l := log.From(ctx)
	l.Info("<-RES", slog.String("body", string(body)))

	if err := json.Unmarshal(body, res); err != nil {
		return fmt.Errorf("unmarshal body: %w", err)
	}
	return nil
}
