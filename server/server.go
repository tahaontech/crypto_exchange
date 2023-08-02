package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/tahaontech/crypto_exchange/orderbook"
)

const (
	exchangePrivateKey = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"

	MarketETH Market = "ETH"

	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
)

type (
	OrderType string
	Market    string

	Order struct {
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		UserID    int64
		Timestamp int64
	}

	OrderbookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Asks           []*Order
		Bids           []*Order
	}

	PlaceOrderRequest struct {
		UserID int64
		Type   OrderType // Limit or Market
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	MatchedOrder struct {
		ID    int64
		Size  float64
		Price float64
	}

	PlaceOrderResponse struct {
		OrderID int64
	}
)

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

func StartServer() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(exchangePrivateKey, client)
	if err != nil {
		log.Fatal(err)
	}

	// demo user
	pkStr8 := "829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4"
	user8 := NewUser(pkStr8, 8)
	ex.users[user8.ID] = user8

	pkStr7 := "a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3"
	user7 := NewUser(pkStr7, 7)
	ex.users[user7.ID] = user7

	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.handleCancelOrder)

	address8 := "0xACa94ef8bD5ffEE41947b4585a84BdA5a3d3DA6E"
	balance8, _ := GetBalance(client, common.HexToAddress(address8))
	fmt.Println("userID 8 balance: ", balance8)

	address7 := "0x28a8746e75304c0780E011BEd21C72cD78cd535E"
	balance7, _ := GetBalance(client, common.HexToAddress(address7))
	fmt.Println("userID 7 balance: ", balance7)

	e.Start(":3000")
}

type User struct {
	ID         int64
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(pk string, id int64) *User {
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:         id,
		PrivateKey: privateKey,
	}
}

type Exchange struct {
	Client     *ethclient.Client
	users      map[int64]*User
	orders     map[int64]int64
	PrivateKey *ecdsa.PrivateKey
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange(privKey string, client *ethclient.Client) (*Exchange, error) {
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return nil, err
	}
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	return &Exchange{
		Client:     client,
		users:      make(map[int64]*User),
		orders:     make(map[int64]int64),
		PrivateKey: privateKey,
		orderbooks: orderbooks,
	}, nil
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
				UserID:    ord.UserID,
				Timestamp: ord.Timestamp,
			}
			orderBookData.Asks = append(orderBookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, ord := range limit.Orders {
			o := Order{
				ID:        ord.ID,
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

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []*MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchesOrder := make([]*MatchedOrder, len(matches))
	isBid := order.Bid
	for i := 0; i < len(matches); i++ {
		var id int64
		if isBid {
			id = matches[i].Ask.ID
		} else {
			id = matches[i].Bid.ID
		}
		matchesOrder[i] = &MatchedOrder{
			ID:    id,
			Size:  matches[i].SizeFilled,
			Price: matches[i].Price,
		}
	}

	return matches, matchesOrder
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)

	fmt.Printf("new LIMIT order => bid [%t] | price [%.2f] | size [%.2f] \n", order.Bid, price, order.Size)

	return nil
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := Market(placeOrderData.Market)
	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	pt := OrderType(placeOrderData.Type)

	if pt == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderData.Price, order); err != nil {
			return c.JSON(500, map[string]any{"error": err.Error()})
		}
		resp := &PlaceOrderResponse{
			OrderID: order.ID,
		}
		return c.JSON(200, resp)
	}

	if pt == MarketOrder {
		matches, matchesOrder := ex.handlePlaceMarketOrder(market, order)
		if err := ex.handleMatches(matches); err != nil {
			return c.JSON(500, map[string]any{"error": err.Error()})
		}
		return c.JSON(200, map[string]any{"msg": "order placed", "matches": matchesOrder})
	}

	return c.JSON(http.StatusBadRequest, "type is invalid")
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketETH]

	order, ok := ob.Orders[int64(id)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "order not found!"})
	}
	ob.CancelOrder(order)

	log.Println("order canceled id => ", order.ID)

	return c.JSON(http.StatusOK, map[string]any{"msg": "order canceled"})
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	for _, matche := range matches {
		fromUser, ok := ex.users[matche.Ask.UserID]
		if !ok {
			return fmt.Errorf("from user not found: %d", matche.Ask.UserID)
		}

		toUser, ok := ex.users[matche.Bid.UserID]
		if !ok {
			return fmt.Errorf("to user not found: %d", matche.Bid.UserID)
		}
		toAddres := crypto.PubkeyToAddress(toUser.PrivateKey.PublicKey)

		amount := big.NewInt(int64(matche.SizeFilled))

		err := transferETH(ex.Client, fromUser.PrivateKey, toAddres, amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func transferETH(client *ethclient.Client, from *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()

	publicKey := from.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
	chainID := big.NewInt(1337)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), from)
	if err != nil {
		return err
	}

	return client.SendTransaction(ctx, signedTx)
}

func GetBalance(client *ethclient.Client, address common.Address) (*big.Int, error) {
	ctx := context.Background()
	return client.BalanceAt(ctx, address, nil)
}
