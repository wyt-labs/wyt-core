package datasource

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	binanceconnector "github.com/binance/binance-connector-go"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

type BinanceDriver struct {
	baseComponent            *base.Component
	client                   *binanceconnector.Client
	lock                     *sync.RWMutex
	subscribeTokenSymbols    []string
	wsMarketStatEventHandler WsMarketStatEventHandler
	websocketStreamStopCh    chan struct{}
	websocketStreamDoneCh    chan struct{}
}

func NewBinanceDriver(baseComponent *base.Component) *BinanceDriver {
	binanceconnector.WebsocketKeepalive = true
	return &BinanceDriver{
		baseComponent: baseComponent,
		client:        binanceconnector.NewClient("", ""),
		lock:          new(sync.RWMutex),
	}
}

func (d *BinanceDriver) Name() string {
	return config.MarketDriverTypeBinance
}

func (d *BinanceDriver) Config(subscribeTokenSymbols []string, wsMarketStatEventHandler WsMarketStatEventHandler) {
	d.subscribeTokenSymbols = subscribeTokenSymbols
	d.wsMarketStatEventHandler = wsMarketStatEventHandler
}

func (d *BinanceDriver) Start() error {
	if err := d.reSubscribeMarketStatByWebsocket(); err != nil {
		return errors.Wrap(err, "failed to subscribe market stat by websocket")
	}
	return nil
}

func (d *BinanceDriver) Stop() error {
	if d.websocketStreamStopCh != nil {
		d.websocketStreamStopCh <- struct{}{}
		<-d.websocketStreamDoneCh
	}
	return nil
}

func (d *BinanceDriver) FlushCache() error {
	return nil
}

func (d *BinanceDriver) UpdateSubscribeTokenSymbols(subscribeTokenSymbols []string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.subscribeTokenSymbols = subscribeTokenSymbols
	return d.reSubscribeMarketStatByWebsocket()
}

func (d *BinanceDriver) FetchKlinesData(tokenSymbol string, interval string, start uint64, end uint64) ([]float64, []time.Time, error) {
	return nil, nil, nil
}

func (d *BinanceDriver) FetchLast7DaysKlinesData(tokenSymbol string) ([]float64, []time.Time, error) {
	symbol := strings.ToUpper(tokenSymbol) + quoteAsset
	endTime := time.Now().AddDate(0, 0, -1)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
	startTime := endTime.AddDate(0, 0, -7)

	var klines []*binanceconnector.KlinesResponse
	err := util.Retry(d.baseComponent.Config.App.RetryInterval.ToDuration(), d.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
		klines, err = d.client.NewKlinesService().
			Symbol(symbol).
			Interval("15m").
			StartTime(uint64(startTime.UnixMilli())).
			EndTime(uint64(endTime.UnixMilli())).
			Do(context.Background())
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				return true, err
			}
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	dates := make([]time.Time, len(klines))
	prices := make([]float64, len(klines))
	for i, kline := range klines {
		closePrice, err := strconv.ParseFloat(kline.Close, 64)
		if err != nil {
			return nil, nil, err
		}
		prices[i] = closePrice
		dates[i] = time.Unix(0, int64(kline.CloseTime)*int64(time.Millisecond))
	}
	return prices, dates, nil
}

func (d *BinanceDriver) reSubscribeMarketStatByWebsocket() error {
	// stop the last stream
	if d.websocketStreamStopCh != nil {
		d.websocketStreamStopCh <- struct{}{}
		<-d.websocketStreamDoneCh
	}

	if len(d.subscribeTokenSymbols) == 0 {
		return nil
	}

	errHandler := func(err error) {
		d.baseComponent.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Warn("Receive error from market websocket connector")
	}
	symbols := lo.Map(d.subscribeTokenSymbols, func(item string, index int) string {
		return item + quoteAsset
	})
	// sdk not support incremental update subscription
	doneCh, stopCh, err := binanceconnector.WsCombinedMarketStatServe(symbols, d.handleMarketWebsocketEvent, errHandler)
	if err != nil {
		return errors.Wrap(err, "failed to receive market stat by websocket")
	}

	d.websocketStreamStopCh = stopCh
	d.websocketStreamDoneCh = doneCh

	d.baseComponent.Logger.Info("Connect to market websocket service")
	return nil
}

func (d *BinanceDriver) handleMarketWebsocketEvent(event *binanceconnector.WsMarketStatEvent) {
	err := func() error {
		symbol := strings.TrimSuffix(event.Symbol, quoteAsset)
		f, err := strconv.ParseFloat(event.WeightedAvgPrice, 64)
		if err != nil {
			return errors.Wrap(err, "failed to parse event price: "+event.WeightedAvgPrice)
		}

		return d.wsMarketStatEventHandler(WsMarketStatEvent{
			Timestamp: event.CloseTime,
			Symbol:    symbol,
			Price:     f,
		})
	}()
	if err != nil {
		d.baseComponent.Logger.WithFields(logrus.Fields{"symbol": event.Symbol, "err": err}).Warn("Failed to handle market event")
	}
}
