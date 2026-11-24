package binance

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type orderResponse struct {
	OrderID            int64           `json:"orderId"`
	ClientOrderID      string          `json:"clientOrderId"`
	Status             string          `json:"status"`
	ExecutedQty        decimal.Decimal `json:"executedQty"`
	AvgPrice           decimal.Decimal `json:"avgPrice"`
	CumulativeQuoteQty decimal.Decimal `json:"cummulativeQuoteQty"`
	CumQuote           decimal.Decimal `json:"cumQuote"`
	Fills              []struct {
		Price           decimal.Decimal `json:"price"`
		Qty             decimal.Decimal `json:"qty"`
		Commission      decimal.Decimal `json:"commission"`
		CommissionAsset string          `json:"commissionAsset"`
	} `json:"fills"`
}

func (o orderResponse) result() OrderResult {
	r := OrderResult{OrderID: o.OrderID, ClientOrderID: o.ClientOrderID, Status: o.Status,
		ExecutedQty: o.ExecutedQty, AvgPrice: o.AvgPrice}
	quote := o.CumulativeQuoteQty
	if quote.IsZero() {
		quote = o.CumQuote
	}
	fillQuote, fillQty := decimal.Zero, decimal.Zero
	for _, fill := range o.Fills {
		r.Fee = r.Fee.Add(fill.Commission)
		if r.FeeAsset == "" {
			r.FeeAsset = fill.CommissionAsset
		}
		fillQty = fillQty.Add(fill.Qty)
		fillQuote = fillQuote.Add(fill.Price.Mul(fill.Qty))
	}
	if r.AvgPrice.IsZero() {
		if o.ExecutedQty.IsPositive() && quote.IsPositive() {
			r.AvgPrice = quote.Div(o.ExecutedQty)
		} else if fillQty.IsPositive() {
			r.AvgPrice = fillQuote.Div(fillQty)
		}
	}
	return r
}

func orderParams(symbol, side string, quantity decimal.Decimal, clientID string) (url.Values, error) {
	side = strings.ToUpper(side)
	if symbol == "" || (side != "BUY" && side != "SELL") || !quantity.IsPositive() {
		return nil, &APIError{Msg: "invalid order parameters"}
	}
	params := url.Values{"symbol": {symbol}, "side": {side}, "quantity": {quantity.String()}}
	if clientID != "" {
		params.Set("newClientOrderId", clientID)
	}
	return params, nil
}

func (c *Client) MarketOrder(ctx context.Context, r OrderRequest) (OrderResult, error) {
	params, err := orderParams(r.Symbol, r.Side, r.Quantity, r.ClientOrderID)
	if err != nil {
		return OrderResult{}, err
	}
	params.Set("type", "MARKET")
	params.Set("newOrderRespType", "FULL")
	if c.market == "futures" {
		params.Set("newOrderRespType", "RESULT")
		params.Set("reduceOnly", strconv.FormatBool(r.ReduceOnly))
	}
	var response orderResponse
	if err := c.request(ctx, http.MethodPost, c.path("order"), params, true, &response); err != nil {
		return OrderResult{}, err
	}
	result := response.result()
	if c.market == "futures" {
		// The order already filled: a failed fee lookup must never hide the fill,
		// so retry once and otherwise report the fill with a zero fee.
		if c.orderFees(ctx, r.Symbol, &result) != nil {
			_ = c.orderFees(ctx, r.Symbol, &result)
		}
	}
	return result, nil
}

func (c *Client) StopOrder(ctx context.Context, r StopRequest) (OrderResult, error) {
	params, err := orderParams(r.Symbol, r.Side, r.Quantity, r.ClientOrderID)
	if err != nil {
		return OrderResult{}, err
	}
	if !r.StopPrice.IsPositive() {
		return OrderResult{}, &APIError{Msg: "invalid stop price"}
	}
	params.Set("stopPrice", r.StopPrice.String())
	if c.market == "futures" {
		params.Set("type", "STOP_MARKET")
		params.Set("reduceOnly", "true")
		params.Set("workingType", "MARK_PRICE")
	} else {
		filters, err := c.Filters(ctx, r.Symbol)
		if err != nil {
			return OrderResult{}, err
		}
		factor := decimal.RequireFromString("0.995")
		if params.Get("side") == "BUY" {
			factor = decimal.RequireFromString("1.005")
		}
		params.Set("type", "STOP_LOSS_LIMIT")
		params.Set("timeInForce", "GTC")
		params.Set("price", filters.RoundPrice(r.StopPrice.Mul(factor)).String())
	}
	var response orderResponse
	err = c.request(ctx, http.MethodPost, c.path("order"), params, true, &response)
	return response.result(), err
}

func (c *Client) CancelOrder(ctx context.Context, symbol string, orderID int64) error {
	err := c.request(ctx, http.MethodDelete, c.path("order"), orderIdentity(symbol, orderID), true, nil)
	var apiErr *APIError
	if errors.As(err, &apiErr) && (apiErr.Code == -2011 || apiErr.Code == -2013) {
		return nil
	}
	return err
}

func (c *Client) QueryOrder(ctx context.Context, symbol string, orderID int64) (OrderResult, error) {
	return c.queryOrder(ctx, symbol, orderIdentity(symbol, orderID))
}

// QueryOrderByClientID recovers an order whose placement response was lost,
// for example after a timeout, using the client order id sent with it.
func (c *Client) QueryOrderByClientID(ctx context.Context, symbol, clientOrderID string) (OrderResult, error) {
	return c.queryOrder(ctx, symbol, url.Values{"symbol": {symbol}, "origClientOrderId": {clientOrderID}})
}

func (c *Client) queryOrder(ctx context.Context, symbol string, params url.Values) (OrderResult, error) {
	var response orderResponse
	if err := c.request(ctx, http.MethodGet, c.path("order"), params, true, &response); err != nil {
		return OrderResult{}, err
	}
	result := response.result()
	if result.ExecutedQty.IsPositive() {
		// Order queries carry no commissions; best effort from the trade list.
		_ = c.orderFees(ctx, symbol, &result)
	}
	return result, nil
}

func orderIdentity(symbol string, orderID int64) url.Values {
	return url.Values{"symbol": {symbol}, "orderId": {strconv.FormatInt(orderID, 10)}}
}

func (c *Client) orderFees(ctx context.Context, symbol string, result *OrderResult) error {
	var trades []struct {
		Commission      decimal.Decimal `json:"commission"`
		CommissionAsset string          `json:"commissionAsset"`
	}
	if err := c.request(ctx, http.MethodGet, c.tradesPath(), orderIdentity(symbol, result.OrderID), true, &trades); err != nil {
		return err
	}
	result.Fee, result.FeeAsset = decimal.Zero, ""
	for _, trade := range trades {
		result.Fee = result.Fee.Add(trade.Commission)
		if result.FeeAsset == "" {
			result.FeeAsset = trade.CommissionAsset
		}
	}
	return nil
}

func (c *Client) tradesPath() string {
	if c.market == "futures" {
		return "/fapi/v1/userTrades"
	}
	return "/api/v3/myTrades"
}
