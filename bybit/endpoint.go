package bybit

import "net/url"

const (
	TestNetAddr1 = "https://api-testnet.bybit.com"
	MainNetAddr1 = "https://api.bybit.com"
	MainNetAddr2 = "https://api.bytick.com"
)

type Endpoint url.URL

func (e *Endpoint) Get(path string) string {
	u := url.URL(*e)
	u.Path = path
	return u.String()
}
