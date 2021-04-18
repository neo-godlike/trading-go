package main

import (
	"context"
	"log"
	"os"

	"github.com/jvongxay/trading-go/trading"
)

func main() {
	apiKey, secretKey := os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY")
	s := trading.NewService(apiKey, secretKey, log.Default())
	s.Monitoring(context.Background(), []string{"BTCUSDT", "ETHUSDT", "BCHUSDT", "LTCUSDT", "DOGEUSDT", "DOTUSDT", "ADAUSDT", "BNBUSDT"})
}
