package trading

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/markcheno/go-talib"
)

func NewService(apiKey, secretKey string, logger *log.Logger) *Service {
	return &Service{
		logger: logger,
		bnb:    binance.NewClient(apiKey, secretKey),
	}
}

type Service struct {
	logger *log.Logger

	bnb *binance.Client
}

type BuyCommand struct {
	Symbol   string
	QuoteQty string
}

func (s *Service) Buy(ctx context.Context, cmd BuyCommand) {
	r, err := s.bnb.NewCreateOrderService().
		Side(binance.SideTypeBuy).
		Symbol(cmd.Symbol).
		Type(binance.OrderTypeMarket).
		QuoteOrderQty(cmd.QuoteQty).
		Do(ctx)
	if err != nil {
		s.logger.Printf("can't place buy: %+v\n", err)
		return
	}

	s.logger.Printf("buy=%s amount=%s id=%d", r.Symbol, r.OrigQuantity, r.OrderID)
}

type SellCommand struct {
	Symbol string
}

func (s *Service) Sell(ctx context.Context, cmd SellCommand) {
	trades, err := s.bnb.NewListTradesService().
		Symbol(cmd.Symbol).
		Limit(1).
		Do(ctx)
	if err != nil {
		s.logger.Printf("can't get trade: %+v", err)
		return
	}
	if len(trades) < 1 {
		s.logger.Printf("you no have enough coin")
		return
	}
	trade := trades[0]

	r, err := s.bnb.NewCreateOrderService().
		Side(binance.SideTypeSell).
		Symbol(cmd.Symbol).
		Type(binance.OrderTypeMarket).
		Quantity(trade.Quantity).
		Do(ctx)
	if err != nil {
		s.logger.Printf("can't place buy: %+v\n", err)
		return
	}

	s.logger.Printf("sell=%s amount=%s id=%d", r.Symbol, r.OrigQuantity, r.OrderID)
}

func (s *Service) Monitoring(ctx context.Context, coins []string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(10 * time.Second):
			for _, coin := range coins {
				lines, err := s.bnb.NewKlinesService().
					Symbol(coin).
					Interval("15m").
					Do(ctx)
				if err != nil {
					s.logger.Printf("can't get klines: %s %+v", coin, err)
					continue
				}

				closes := make([]float64, 0, len(lines))
				for _, line := range lines {
					close, _ := strconv.ParseFloat(line.Close, 64)
					closes = append(closes, close)
				}

				ema12 := talib.Ema(closes, 12)
				ema26 := talib.Ema(closes, 26)
				rsi := talib.Rsi(closes, 14)

				ema12Len, ema26Len := len(ema12), len(ema26)
				buy := (ema12[ema12Len-2] < ema26[ema26Len-2]) && (ema12[ema12Len-1] > ema26[ema26Len-1])
				sell := (ema12[ema12Len-2] > ema26[ema26Len-2]) && (ema12[ema12Len-1] < ema26[ema26Len-1])

				oversold := rsi[len(rsi)-1] < 30
				overbought := rsi[len(rsi)-1] > 70

				if oversold {
					s.logger.Printf("%s OVERSOLD RSI < 30", coin)
				} else if overbought {
					s.logger.Printf("%s OVERBOUGHT RSI > 70", coin)
				} else {
					s.logger.Printf("%s MIDDLE 30 < RSI < 70", coin)
				}

				if buy {
					s.Buy(ctx, BuyCommand{
						Symbol:   coin,
						QuoteQty: "15",
					})
				} else if sell {
					s.Sell(ctx, SellCommand{
						Symbol: coin,
					})
					s.logger.Printf("Sell %s", coin)
				} else {
					s.logger.Printf("None")
				}
			}
		}
		s.logger.Println("-----------------------------------------------")
	}
}
