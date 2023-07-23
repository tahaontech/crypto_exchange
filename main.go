package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tahaontech/crypto_exchange/orderbook"
)

func main() {
	e := echo.New()
	ex := NewExchange()

	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.handleCancelOrder)

	e.Start(":3000")
}

type OrderType string

const (
	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
)

type Market string

const (
	MarketETH Market = "ETH"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange() *Exchange {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	return &Exchange{
		orderbooks: orderbooks,
	}
}

type Order struct {
	ID        int64
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}
type OrderbookData struct {
	TotalBidVolume float64
	TotalAskVolume float64
	Asks           []*Order
	Bids           []*Order
}

type PlaceOrderRequest struct {
	Type   OrderType // Limit or Market
	Bid    bool
	Size   float64
	Price  float64
	Market Market
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)
	ob := ex.orderbooks[market]
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size)

	if placeOrderData.Type == LimitOrder {
		ob.PlaceLimitOrder(placeOrderData.Price, order)
		return c.JSON(200, map[string]any{"msg": "order placed"})
	}

	if placeOrderData.Type == MarketOrder {
		matches := ob.PlaceMarketOrder(order)
		return c.JSON(200, map[string]any{"msg": "order placed", "matches": len(matches)})
	}

	return c.JSON(http.StatusBadRequest, "type is invalid")
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketETH]

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			if order.ID == int64(id) {
				ob.CancelOrder(order)
				return c.JSON(http.StatusOK, map[string]any{"msg": "order canceled"})
			}
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			if order.ID == int64(id) {
				ob.CancelOrder(order)
				return c.JSON(http.StatusOK, map[string]any{"msg": "order canceled"})
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]any{"msg": "order not canceled"})
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob, ok := ex.orderbooks[market]

	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "market not found"})
	}

	orderBookData := OrderbookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}
	for _, limit := range ob.Asks() {
		for _, ord := range limit.Orders {
			o := Order{
				ID:        ord.ID,
				Price:     limit.Price,
				Size:      ord.Size,
				Bid:       ord.Bid,
				Timestamp: ord.Timestamp,
			}
			orderBookData.Asks = append(orderBookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, ord := range limit.Orders {
			o := Order{
				Price:     limit.Price,
				Size:      ord.Size,
				Bid:       ord.Bid,
				Timestamp: ord.Timestamp,
			}
			orderBookData.Bids = append(orderBookData.Asks, &o)
		}
	}

	return c.JSON(http.StatusOK, orderBookData)
}
