package main

import (
	"fmt"
	"time"

	"github.com/tahaontech/crypto_exchange/client"
	"github.com/tahaontech/crypto_exchange/server"
)

func main() {
	go server.StartServer()

	time.Sleep(1 * time.Second)

	c := client.NewClient()

	bidParams := &client.PlaceOrderParams{
		UserID: 8,
		Bid:    true,
		Size:   1_000_000,
		Price:  10_000,
	}

	go func() {
		for {
			resp, err := c.PlaceLimitOrder(bidParams)
			if err != nil {
				panic(err)
			}

			fmt.Printf("bid order id => %d\n", resp.OrderID)

			if err := c.CancelOrder(resp.OrderID); err != nil {
				panic(err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	askParams := &client.PlaceOrderParams{
		UserID: 8,
		Bid:    false,
		Size:   1_000,
		Price:  8_000,
	}
	go func() {
		for {
			resp, err := c.PlaceLimitOrder(askParams)
			if err != nil {
				panic(err)
			}

			fmt.Printf("ask order id => %d\n", resp.OrderID)

			if err := c.CancelOrder(resp.OrderID); err != nil {
				panic(err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	select {}
}
