package binance

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

func (c *Client) VerifyKeys(ctx context.Context) error {
	path := c.path("account")
	if c.market == "futures" {
		path = "/fapi/v2/account"
	}
	var account struct {
		CanTrade *bool `json:"canTrade"`
	}
	if err := c.request(ctx, http.MethodGet, path, nil, true, &account); err != nil {
		return err
	}
	if account.CanTrade != nil && !*account.CanTrade {
		return ErrInvalidKeys
	}
	return nil
}

func (c *Client) Balance(ctx context.Context, asset string) (Balance, error) {
	if c.market == "futures" {
		var balances []struct {
			Asset            string          `json:"asset"`
			AvailableBalance decimal.Decimal `json:"availableBalance"`
			Balance          decimal.Decimal `json:"balance"`
			WalletBalance    decimal.Decimal `json:"walletBalance"`
		}
		if err := c.request(ctx, http.MethodGet, "/fapi/v2/balance", nil, true, &balances); err != nil {
			return Balance{}, err
		}
		for _, balance := range balances {
			if balance.Asset == strings.ToUpper(asset) {
				total := balance.WalletBalance
				if total.IsZero() {
					total = balance.Balance
				}
				return Balance{Free: balance.AvailableBalance, Total: total}, nil
			}
		}
		return Balance{}, nil
	}
	var account struct {
		Balances []struct {
			Asset        string
			Free, Locked decimal.Decimal
		} `json:"balances"`
	}
	if err := c.request(ctx, http.MethodGet, c.path("account"), nil, true, &account); err != nil {
		return Balance{}, err
	}
	for _, balance := range account.Balances {
		if balance.Asset == strings.ToUpper(asset) {
			return Balance{Free: balance.Free, Total: balance.Free.Add(balance.Locked)}, nil
		}
	}
	return Balance{}, nil
}

func (c *Client) Price(ctx context.Context, symbol string) (decimal.Decimal, error) {
	var ticker struct {
		Price decimal.Decimal `json:"price"`
	}
	err := c.request(ctx, http.MethodGet, c.path("ticker/price"), url.Values{"symbol": {symbol}}, false, &ticker)
	return ticker.Price, err
}

func (c *Client) SetLeverage(ctx context.Context, symbol string, leverage uint32) error {
	if c.market == "spot" {
		return nil
	}
	return c.request(ctx, http.MethodPost, c.path("leverage"), url.Values{
		"symbol": {symbol}, "leverage": {strconv.FormatUint(uint64(leverage), 10)},
	}, true, nil)
}
