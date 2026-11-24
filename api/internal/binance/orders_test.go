package binance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/shopspring/decimal"
)

func TestSpotMarketOrder(t *testing.T) {
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		checkSigned(t, r)
		if r.Method != http.MethodPost || r.URL.Path != "/api/v3/order" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "side": "BUY", "type": "MARKET",
			"quantity": "0.03", "newClientOrderId": "market-order-1", "newOrderRespType": "FULL"})
		if r.URL.Query().Has("reduceOnly") {
			t.Error("spot order includes reduceOnly")
		}
		fmt.Fprint(w, `{"orderId":123,"clientOrderId":"market-order-1","status":"FILLED","executedQty":"0.03",
			"cummulativeQuoteQty":"3002","fills":[
			{"price":"100000","qty":"0.01","commission":"0.00001","commissionAsset":"BTC"},
			{"price":"100100","qty":"0.02","commission":"0.00002","commissionAsset":"BTC"}]}`)
	})
	result, err := client.MarketOrder(context.Background(), OrderRequest{Symbol: "BTCUSDT", Side: "buy",
		Quantity: decimal.RequireFromString("0.03"), ReduceOnly: true, ClientOrderID: "market-order-1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.OrderID != 123 || result.ClientOrderID != "market-order-1" || result.Status != "FILLED" || result.FeeAsset != "BTC" {
		t.Fatalf("incorrect result: %+v", result)
	}
	checkDecimal(t, result.ExecutedQty, "0.03")
	checkDecimal(t, result.AvgPrice, "100066.6666666666666667")
	checkDecimal(t, result.Fee, "0.00003")
}

func TestFuturesMarketOrder(t *testing.T) {
	for _, reduceOnly := range []bool{false, true} {
		t.Run(fmt.Sprintf("reduceOnly=%t", reduceOnly), func(t *testing.T) {
			orderCalls, feeCalls := 0, 0
			client := newTestClient(t, "futures", func(w http.ResponseWriter, r *http.Request) {
				checkSigned(t, r)
				switch r.URL.Path {
				case "/fapi/v1/order":
					orderCalls++
					if r.Method != http.MethodPost {
						t.Errorf("method = %s", r.Method)
					}
					checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "side": "SELL", "type": "MARKET",
						"quantity": "0.03", "reduceOnly": fmt.Sprint(reduceOnly), "newOrderRespType": "RESULT", "newClientOrderId": "futures-1"})
					fmt.Fprint(w, `{"orderId":456,"clientOrderId":"futures-1","status":"FILLED","executedQty":"0.03","avgPrice":"100100.5","cumQuote":"3003.015"}`)
				case "/fapi/v1/userTrades":
					feeCalls++
					if r.Method != http.MethodGet {
						t.Errorf("method = %s", r.Method)
					}
					checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "orderId": "456"})
					fmt.Fprint(w, `[{"orderId":456,"commission":"0.5","commissionAsset":"USDT"},{"orderId":456,"commission":"1.0015075","commissionAsset":"USDT"}]`)
				default:
					t.Errorf("unexpected path: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			})
			result, err := client.MarketOrder(context.Background(), OrderRequest{Symbol: "BTCUSDT", Side: "sell",
				Quantity: decimal.RequireFromString("0.03"), ReduceOnly: reduceOnly, ClientOrderID: "futures-1"})
			if err != nil {
				t.Fatal(err)
			}
			if orderCalls != 1 || feeCalls != 1 || result.OrderID != 456 || result.Status != "FILLED" || result.ClientOrderID != "futures-1" || result.FeeAsset != "USDT" {
				t.Fatalf("incorrect result: %+v (orders=%d, fees=%d)", result, orderCalls, feeCalls)
			}
			checkDecimal(t, result.ExecutedQty, "0.03")
			checkDecimal(t, result.AvgPrice, "100100.5")
			checkDecimal(t, result.Fee, "1.5015075")
		})
	}
}

func TestSpotStopOrder(t *testing.T) {
	for _, test := range []struct{ side, price string }{{"sell", "99.5"}, {"buy", "100.5"}} {
		t.Run(test.side, func(t *testing.T) {
			client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v3/exchangeInfo" {
					fmt.Fprint(w, exchangeInfo)
					return
				}
				checkSigned(t, r)
				if r.Method != http.MethodPost || r.URL.Path != "/api/v3/order" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "type": "STOP_LOSS_LIMIT", "quantity": "0.03",
					"stopPrice": "100.03", "price": test.price, "timeInForce": "GTC", "newClientOrderId": "stop-1"})
				if r.URL.Query().Has("reduceOnly") || r.URL.Query().Has("workingType") {
					t.Error("spot stop contains futures parameters")
				}
				fmt.Fprint(w, `{"orderId":789,"clientOrderId":"stop-1","status":"NEW","executedQty":"0"}`)
			})
			result, err := client.StopOrder(context.Background(), StopRequest{Symbol: "BTCUSDT", Side: test.side,
				Quantity: decimal.RequireFromString("0.03"), StopPrice: decimal.RequireFromString("100.03"), ClientOrderID: "stop-1"})
			if err != nil {
				t.Fatal(err)
			}
			if result.OrderID != 789 || result.ClientOrderID != "stop-1" || result.Status != "NEW" {
				t.Fatalf("incorrect stop result: %+v", result)
			}
			checkDecimal(t, result.ExecutedQty, "0")
			checkDecimal(t, result.AvgPrice, "0")
		})
	}
}

func TestFuturesStopOrder(t *testing.T) {
	client := newTestClient(t, "futures", func(w http.ResponseWriter, r *http.Request) {
		checkSigned(t, r)
		if r.Method != http.MethodPost || r.URL.Path != "/fapi/v1/order" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "side": "SELL", "type": "STOP_MARKET", "quantity": "0.03",
			"stopPrice": "99000.1", "reduceOnly": "true", "workingType": "MARK_PRICE", "newClientOrderId": "stop-futures-1"})
		if r.URL.Query().Has("price") || r.URL.Query().Has("timeInForce") || r.URL.Query().Has("closePosition") {
			t.Error("futures stop contains incompatible parameters")
		}
		fmt.Fprint(w, `{"orderId":789,"clientOrderId":"stop-futures-1","status":"NEW","executedQty":"0","avgPrice":"0"}`)
	})
	result, err := client.StopOrder(context.Background(), StopRequest{Symbol: "BTCUSDT", Side: "sell",
		Quantity: decimal.RequireFromString("0.03"), StopPrice: decimal.RequireFromString("99000.1"), ClientOrderID: "stop-futures-1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.OrderID != 789 || result.Status != "NEW" {
		t.Fatalf("incorrect result: %+v", result)
	}
}

func TestCancelOrderIdempotence(t *testing.T) {
	for _, market := range []Market{"spot", "futures"} {
		for _, code := range []int{0, -2011, -2013, -1013, -2015} {
			t.Run(fmt.Sprintf("%s/%d", market, code), func(t *testing.T) {
				client := newTestClient(t, market, func(w http.ResponseWriter, r *http.Request) {
					checkSigned(t, r)
					path := "/api/v3/order"
					if market == "futures" {
						path = "/fapi/v1/order"
					}
					if r.Method != http.MethodDelete || r.URL.Path != path {
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					}
					checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "orderId": "789"})
					if code == 0 {
						fmt.Fprint(w, `{"orderId":789,"status":"CANCELED"}`)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, `{"code":%d,"msg":"order error"}`, code)
				})
				err := client.CancelOrder(context.Background(), "BTCUSDT", 789)
				switch code {
				case 0, -2011, -2013:
					if err != nil {
						t.Fatal(err)
					}
				case -2015:
					if !errors.Is(err, ErrInvalidKeys) {
						t.Fatal(err)
					}
				default:
					if !errors.Is(err, ErrRejected) {
						t.Fatal(err)
					}
				}
			})
		}
	}
}

func TestQueryOrder(t *testing.T) {
	for _, market := range []Market{"spot", "futures"} {
		t.Run(string(market), func(t *testing.T) {
			client := newTestClient(t, market, func(w http.ResponseWriter, r *http.Request) {
				checkSigned(t, r)
				checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "orderId": "789"})
				path, trades := "/api/v3/order", "/api/v3/myTrades"
				if market == "futures" {
					path, trades = "/fapi/v1/order", "/fapi/v1/userTrades"
				}
				if r.Method == http.MethodGet && r.URL.Path == trades {
					fmt.Fprint(w, `[{"commission":"0.004","commissionAsset":"USDT"},{"commission":"0.006","commissionAsset":"USDT"}]`)
					return
				}
				if r.Method != http.MethodGet || r.URL.Path != path {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				fmt.Fprint(w, `{"orderId":789,"clientOrderId":"stop-1","status":"FILLED","executedQty":"0.1","cummulativeQuoteQty":"10000"}`)
			})
			result, err := client.QueryOrder(context.Background(), "BTCUSDT", 789)
			if err != nil {
				t.Fatal(err)
			}
			if result.OrderID != 789 || result.Status != "FILLED" || result.ClientOrderID != "stop-1" {
				t.Fatalf("incorrect result: %+v", result)
			}
			checkDecimal(t, result.ExecutedQty, "0.1")
			checkDecimal(t, result.AvgPrice, "100000")
			checkDecimal(t, result.Fee, "0.01")
		})
	}
}

func TestFuturesFeeFailurePreservesOrder(t *testing.T) {
	orders := 0
	client := newTestClient(t, "futures", func(w http.ResponseWriter, r *http.Request) {
		checkSigned(t, r)
		if r.URL.Path == "/fapi/v1/order" {
			orders++
			fmt.Fprint(w, `{"orderId":123,"status":"FILLED","executedQty":"1","avgPrice":"100"}`)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	result, err := client.MarketOrder(context.Background(), OrderRequest{Symbol: "BTCUSDT", Side: "buy", Quantity: decimal.NewFromInt(1)})
	// The fill must survive a failed fee lookup: no error, zero fee, one order.
	if err != nil || result.OrderID != 123 || !result.Fee.IsZero() || orders != 1 {
		t.Fatalf("result = %+v, error = %v, orders = %d", result, err, orders)
	}
}

func TestInvalidOrdersDoNotSend(t *testing.T) {
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		t.Error("invalid order sent to exchange")
		w.WriteHeader(http.StatusBadRequest)
	})
	ctx := context.Background()
	for _, request := range []OrderRequest{
		{Symbol: "BTCUSDT", Side: "buy"},
		{Symbol: "BTCUSDT", Side: "invalid", Quantity: decimal.NewFromInt(1)},
		{Side: "buy", Quantity: decimal.NewFromInt(1)},
	} {
		if _, err := client.MarketOrder(ctx, request); !errors.Is(err, ErrRejected) {
			t.Fatalf("validation error = %v", err)
		}
	}
	if _, err := client.StopOrder(ctx, StopRequest{Symbol: "BTCUSDT", Side: "sell", Quantity: decimal.NewFromInt(1)}); !errors.Is(err, ErrRejected) {
		t.Fatalf("stop validation error = %v", err)
	}
}

func TestSpotFillPriceFallback(t *testing.T) {
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"orderId":1,"status":"FILLED","executedQty":"3","fills":[
			{"price":"100","qty":"1","commission":"0.1","commissionAsset":"USDT"},
			{"price":"103","qty":"2","commission":"0.2","commissionAsset":"USDT"}]}`)
	})
	result, err := client.MarketOrder(context.Background(), OrderRequest{Symbol: "BTCUSDT", Side: "buy", Quantity: decimal.NewFromInt(3)})
	if err != nil {
		t.Fatal(err)
	}
	checkDecimal(t, result.AvgPrice, "102")
	checkDecimal(t, result.Fee, "0.3")
}

func TestQueryOrderByClientID(t *testing.T) {
	client := newTestClient(t, "spot", func(w http.ResponseWriter, r *http.Request) {
		checkSigned(t, r)
		if r.URL.Path == "/api/v3/myTrades" {
			fmt.Fprint(w, `[]`)
			return
		}
		checkParams(t, r, map[string]string{"symbol": "BTCUSDT", "origClientOrderId": "open-1"})
		fmt.Fprint(w, `{"orderId":5,"clientOrderId":"open-1","status":"FILLED","executedQty":"2","cummulativeQuoteQty":"200"}`)
	})
	result, err := client.QueryOrderByClientID(context.Background(), "BTCUSDT", "open-1")
	if err != nil || result.OrderID != 5 || result.Status != "FILLED" {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
	checkDecimal(t, result.AvgPrice, "100")
}
