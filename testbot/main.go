package main

import (
	"log"
	"fmt"
	"github.com/toorop/go-bittrex"
	"github.com/shopspring/decimal"
)

const (
	API_KEY       = "API_KEY"
	API_SECRET    = "API_SECRET"
	BUY_STRING    = "BTC"
	SELL_STRING   = "VTC"
	MARKET_STRING = BUY_STRING + "-" + SELL_STRING
	MIN_GAIN      = 0.02
	MAX_LOSS      = 0.02
	ORDER_RANGE      = 0.02
 
	BUY_TRIGGER    = 10000.0
	SELL_TRIGGER   = -10000.0
	ORDER_VARIANCE = 0.02
) 

var (
	balances     []bittrex.Balance
	orders       []bittrex.Order
	ticker       = bittrex.Ticker{}
	lastPrice    float64
	lastBuyPrice = 0.00
	buySellIndex = 0.00
	openOrder    = false
	readyToRun   = false
 
	highIndex = 0.00
	lowIndex  = 0.00
 )

 func allowSell() bool {
	if lastBuyPrice > 0 {
	   gain := lastPrice / lastBuyPrice
	   if gain < (1.00 - MAX_LOSS) {
		  return true
	   }
	   if gain < (1.00 + MIN_GAIN) {
		  return false
	   }
	}
	return true
 }

func main() {
	// Bittrex client
	bittrex := bittrex.New(API_KEY, API_SECRET)

	// Get markets
	// markets, err := bittrex.GetMarkets()
	// fmt.Println(err, markets)

	balances, err := bittrex.GetBalances()
	if err == nil {
		for _, b := range balances {
		   fmt.Println(b)
		}
	 } else {
		fmt.Println(err)
	 }

	 decideBuySell(bittrex);
}

func decideBuySell(b *bittrex.Bittrex) {

	if openOrder {
		// Should we close the open order?
		for _, o := range orders {
		   ppu, _ := o.PricePerUnit.Float64()
		   log.Printf("Order percent: %.4f\n", ppu/lastPrice)
		   if ppu/lastPrice > (1.00+ORDER_VARIANCE) || ppu/lastPrice < (1.00-ORDER_VARIANCE) {
			  log.Println("Canceled order: ", o.OrderUuid)
			  b.CancelOrder(o.OrderUuid)
			  // We assume we only have one order at a time
		   }
		}
	 }
	 // If we have no open order should we buy or sell?
	 if !openOrder {
		if buySellIndex > BUY_TRIGGER {
		   log.Println("BUY TRIGGER ACTIVE!")
		   for _, bals := range balances {
			  bal, _ := bals.Balance.Float64()
			  if BUY_STRING == bals.Currency {
				 //log.Printf("Bal: %.4f %s == %s\n", bal/lastPrice, SELL_STRING, bals.Currency)
			  }
			  if bal > 0.01 && BUY_STRING == bals.Currency && lastPrice > 0.00 {
				 // Place buy order
				 log.Printf("Placed buy order of %.4f %s at %.8f\n=================================================\n", (bal/lastPrice)-5, BUY_STRING, lastPrice)
				 order, err := b.BuyLimit(MARKET_STRING, decimal.NewFromFloat((bal/lastPrice)-5), decimal.NewFromFloat(lastPrice))
				 if err != nil {
					log.Println("ERROR ", err)
				 } else {
					log.Println("Confirmed: ", order)
				 }
				 lastBuyPrice = lastPrice
				 openOrder = true
			  }
		   }
		} else if buySellIndex < SELL_TRIGGER {
		   log.Println("SELL TRIGGER ACTIVE!")
		   for _, bals := range balances {
			  bal, _ := bals.Balance.Float64()
			  if SELL_STRING == bals.Currency {
				 //allow := "false"
				 //if allowSell() {
				 // allow = "true"
				 //}
				 //log.Printf("Bal: %.4f %s == %s && %s\n", bal, BUY_STRING, bals.Currency, allow)
			  }
			  if bal > 0.01 && SELL_STRING == bals.Currency && lastPrice > 0.00 && allowSell() {
				  //SellLimit
			
				 // Place sell order
				 log.Printf("Placed sell order of %.4f %s at %.8f\n=================================================\n", bal, BUY_STRING, lastPrice)
				//  order, err := b.SellLimit(MARKET_STRING, decimal.NewFromFloat(bal), decimal.NewFromFloat(lastPrice))
				//  if err != nil {
				// 	log.Println("ERROR ", err)
				//  } else {
				// 	log.Println("Confirmed: ", order)
				//  }
				 openOrder = true
			  }
		   }
		}
	 }


}


func calculateIndex(buy bool, q float64, r float64) {

  // q is quantity VTC
   // r is the rate
   percent := 0.00

   // Calculate percentage of rate
   if r > 0 && q > 0 && lastPrice > 0 && readyToRun {
	percent = lastPrice / r
	if buy {
	   //log.Printf("Buy percent: %.4f\n", percent)
	   //log.Printf("Buy quantity: %.4f\n", q)
	   if percent > (1.00 - ORDER_RANGE) && percent < (1.00 + ORDER_RANGE) {
		  buySellIndex = buySellIndex + (percent * q)
	   }
	} else {
	   //log.Printf("Sell percent: %.4f\n", percent)
	   //log.Printf("Sell quantity: %.4f\n", q)
	   if percent > (1.00 - ORDER_RANGE) && percent < (1.00 + ORDER_RANGE) {
		  percent = percent - 2.00 // Reverse percent, lower is higher
		  buySellIndex = buySellIndex + (percent * q)
	   }
	}
 }
 if buySellIndex > highIndex {
	highIndex = buySellIndex
 }
 if buySellIndex < lowIndex {
	lowIndex = buySellIndex
 }
 // Reset really high or low numbers due to startup
 if highIndex > 5000000.00 || lowIndex < -5000000.00 {
	highIndex = 0.00
	lowIndex = 0.00
	buySellIndex = 0.00
 }


}