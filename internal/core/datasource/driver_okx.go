package datasource

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

type OkxProduct struct {
	Symbol     string
	ProductID  string
	QuoteToken string
}

type OkxWsInfo struct {
	baseComponent          *base.Component
	tokenSymbolToProductID map[string]string

	tokenSymbols []string
	handler      func(rawEvent []byte)

	c                *websocket.Conn
	lock             *sync.Mutex
	isManualShutdown bool
}

func OkxConnectToWs(baseComponent *base.Component, tokenSymbolToProductID map[string]string, tokenSymbols []string, handler func(rawEvent []byte)) (*OkxWsInfo, error) {
	ws := &OkxWsInfo{
		baseComponent:          baseComponent,
		tokenSymbolToProductID: tokenSymbolToProductID,
		tokenSymbols:           tokenSymbols,
		handler:                handler,
		lock:                   new(sync.Mutex),
	}

	if err := ws.reSubscribe(tokenSymbols, false); err != nil {
		return nil, err
	}
	return ws, nil
}

type okxSubscribeReq struct {
	Op   string   `json:"op"`
	Args []okxArg `json:"args"`
}

type okxSubscribeRes struct {
	Event string `json:"event"`
	Code  string `json:"code"`
	Msg   string `json:"msg"`
}

func (w *OkxWsInfo) reSubscribe(tokenSymbols []string, isReconnect bool) error {
	tokenSymbols = lo.Map(tokenSymbols, func(item string, index int) string {
		return strings.ToUpper(item)
	})
	var filteredTokenSymbols []string
	for _, token := range tokenSymbols {
		if _, ok := w.tokenSymbolToProductID[token]; ok {
			filteredTokenSymbols = append(filteredTokenSymbols, token)
		}
	}

	w.lock.Lock()
	defer w.lock.Unlock()
	if isReconnect {
		if w.isManualShutdown {
			return nil
		}

		d1, d2 := lo.Difference(filteredTokenSymbols, w.tokenSymbols)
		if len(d1) != 0 || len(d2) != 0 {
			// has updated resubscribe
			return nil
		}
	}
	w.tokenSymbols = filteredTokenSymbols
	if len(w.tokenSymbols) == 0 {
		return nil
	}
	if w.c != nil {
		_ = w.c.Close()
		w.c = nil
	}

	dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  20 * time.Second,
		EnableCompression: false,
	}
	c, _, err := dialer.Dial(w.baseComponent.Config.Datasource.Market.Okx.WebsocketEndpoint, nil)
	if err != nil {
		return err
	}
	c.SetReadLimit(655350)

	// send subscribe info
	err = c.WriteJSON(&okxSubscribeReq{
		Op: "subscribe",
		Args: lo.Map(w.tokenSymbols, func(item string, index int) okxArg {
			return okxArg{
				Channel: "candle1m",
				InstID:  w.tokenSymbolToProductID[item],
			}
		}),
	})
	if err != nil {
		return err
	}
	_, message, err := c.ReadMessage()
	if err != nil {
		return err
	}
	var res okxSubscribeRes
	if err := json.Unmarshal(message, &res); err != nil {
		return err
	}
	if res.Event == "error" {
		return errors.Errorf("failed to send subscribe msg, code: %s, err: %s", res.Code, res.Msg)
	}
	w.c = c
	if isReconnect {
		w.baseComponent.Logger.WithField("driver", config.MarketDriverTypeOkx).Info("Reconnect to market websocket service")
	} else {
		w.baseComponent.Logger.WithField("driver", config.MarketDriverTypeOkx).Info("Connect to market websocket service")
	}

	w.baseComponent.SafeGo(func() {
		websocketkeepAlive(c, 30*time.Second)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if _, ok := err.(*websocket.CloseError); !ok {
					w.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "driver": config.MarketDriverTypeOkx}).Warn("Failed to handle market event")

					w.baseComponent.SafeGo(func() {
						err := util.Retry(w.baseComponent.Config.App.RetryInterval.ToDuration(), w.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
							err = w.reSubscribe(filteredTokenSymbols, true)
							return err != nil, err
						})
						if err != nil {
							w.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "driver": config.MarketDriverTypeOkx}).Error("Failed to try reconnect to market driver by websocket")
						}
					})
				}
				return
			}
			w.handler(message)
		}
	})

	return nil
}

func (w *OkxWsInfo) stop() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.c != nil {
		_ = w.c.Close()
	}
	w.isManualShutdown = true
}

type OkxDriver struct {
	baseComponent             *base.Component
	systemCacheDao            *dao.SystemCacheDao
	httpClient                *resty.Client
	lock                      *sync.RWMutex
	tokenSymbolToProductID    map[string]string
	tokenProductIDToSymbol    map[string]string
	subscribeTokenSymbols     []string
	waitSubscribeTokenSymbols []string
	wsMarketStatEventHandler  WsMarketStatEventHandler
	wsList                    []*OkxWsInfo
}

func NewOkxDriver(baseComponent *base.Component, systemCacheDao *dao.SystemCacheDao) *OkxDriver {
	httpClient := resty.New()
	httpClient.SetBaseURL(baseComponent.Config.Datasource.Market.Okx.APIEndpoint)
	return &OkxDriver{
		baseComponent:          baseComponent,
		systemCacheDao:         systemCacheDao,
		httpClient:             httpClient,
		lock:                   new(sync.RWMutex),
		tokenSymbolToProductID: make(map[string]string),
		tokenProductIDToSymbol: make(map[string]string),
	}
}

func (d *OkxDriver) Name() string {
	return config.MarketDriverTypeOkx
}

func (d *OkxDriver) FlushCache() error {
	return nil
}

func (d *OkxDriver) Config(subscribeTokenSymbols []string, wsMarketStatEventHandler WsMarketStatEventHandler) {
	d.waitSubscribeTokenSymbols = lo.Map(subscribeTokenSymbols, func(item string, index int) string {
		return strings.ToUpper(item)
	})
	d.wsMarketStatEventHandler = wsMarketStatEventHandler
}

func (d *OkxDriver) Start() error {
	allProducts, err := d.fetchAllProducts()
	if err != nil {
		return err
	}
	for _, product := range allProducts {
		if product.QuoteToken == "USDC" || product.QuoteToken == "USDT" {
			if _, ok := d.tokenSymbolToProductID[product.Symbol]; !ok {
				d.tokenSymbolToProductID[product.Symbol] = product.ProductID
				d.tokenProductIDToSymbol[product.ProductID] = product.Symbol
			}
		}
	}

	if err := d.reSubscribeMarketStatByWebsocket(d.waitSubscribeTokenSymbols); err != nil {
		return errors.Wrap(err, "failed to subscribe market stat by websocket")
	}
	return nil
}

func (d *OkxDriver) Stop() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	for _, ws := range d.wsList {
		ws.stop()
	}
	return nil
}

type okxFetchProductsRes struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data []okxDatum `json:"data"`
}

type okxDatum struct {
	Alias        string `json:"alias"`
	BaseCcy      string `json:"baseCcy"`
	Category     string `json:"category"`
	CTMult       string `json:"ctMult"`
	CTType       string `json:"ctType"`
	CTVal        string `json:"ctVal"`
	CTValCcy     string `json:"ctValCcy"`
	ExpTime      string `json:"expTime"`
	InstFamily   string `json:"instFamily"`
	InstID       string `json:"instId"`
	InstType     string `json:"instType"`
	Lever        string `json:"lever"`
	ListTime     string `json:"listTime"`
	LotSz        string `json:"lotSz"`
	MaxIcebergSz string `json:"maxIcebergSz"`
	MaxLmtSz     string `json:"maxLmtSz"`
	MaxMktSz     string `json:"maxMktSz"`
	MaxStopSz    string `json:"maxStopSz"`
	MaxTriggerSz string `json:"maxTriggerSz"`
	MaxTwapSz    string `json:"maxTwapSz"`
	MinSz        string `json:"minSz"`
	OptType      string `json:"optType"`
	QuoteCcy     string `json:"quoteCcy"`
	SettleCcy    string `json:"settleCcy"`
	State        string `json:"state"`
	Stk          string `json:"stk"`
	TickSz       string `json:"tickSz"`
	Uly          string `json:"uly"`
}

func (d *OkxDriver) fetchAllProducts() ([]*OkxProduct, error) {
	var res okxFetchProductsRes
	err := util.Retry(d.baseComponent.Config.App.RetryInterval.ToDuration(), d.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
		resp, err := d.httpClient.R().
			Get("/api/v5/public/instruments?instType=SPOT")
		if err != nil {
			return true, err
		}
		if resp.StatusCode() != http.StatusOK {
			return true, errors.Errorf("http request failed, code: %d, msg: %s", resp.StatusCode(), resp.String())
		}
		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return false, err
		}
		if res.Code != "0" {
			return true, errors.Errorf("api request failed, code: %s, msg: %s", res.Code, res.Msg)
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return lo.Map(res.Data, func(item okxDatum, index int) *OkxProduct {
		return &OkxProduct{
			Symbol:     item.BaseCcy,
			ProductID:  item.InstID,
			QuoteToken: item.QuoteCcy,
		}
	}), nil
}

func (d *OkxDriver) UpdateSubscribeTokenSymbols(subscribeTokenSymbols []string) error {
	return d.reSubscribeMarketStatByWebsocket(lo.Map(subscribeTokenSymbols, func(item string, index int) string {
		return strings.ToUpper(item)
	}))
}

type okxFetchKlinesDataResp struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

func (d *OkxDriver) FetchKlinesData(tokenSymbol string, interval string, start uint64, end uint64) ([]float64, []time.Time, error) {
	productID, ok := d.tokenSymbolToProductID[tokenSymbol]
	if !ok {
		return nil, nil, ErrUnsupportedToken
	}

	bar := ""
	switch interval {
	case "1d":
		bar = "1D"
	case "1w":
		bar = "1W"
	case "1M":
		bar = "1M"
	case "15m":
		bar = "15m"
	default:
		return nil, nil, errcode.ErrRequestParameter.Wrap("unsupported interval")
	}

	var fetchKlinesDataResp okxFetchKlinesDataResp
	err := util.Retry(d.baseComponent.Config.App.RetryInterval.ToDuration(), d.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
		resp, err := d.httpClient.R().
			SetQueryParams(map[string]string{
				"instId": productID,
				"bar":    bar,
				"after":  strconv.Itoa(int(start * 1000)),
				"before": strconv.Itoa(int(end * 1000)),
			}).
			Get("/api/v5/market/history-candles")
		if err != nil {
			return true, err
		}
		if resp.StatusCode() != http.StatusOK {
			return true, errors.Errorf("http request failed, code: %d, msg: %s", resp.StatusCode(), resp.String())
		}
		if err := json.Unmarshal(resp.Body(), &fetchKlinesDataResp); err != nil {
			return false, err
		}
		if fetchKlinesDataResp.Code != "0" {
			return true, errors.Errorf("api request failed, code: %s, msg: %s", fetchKlinesDataResp.Code, fetchKlinesDataResp.Msg)
		}

		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if len(fetchKlinesDataResp.Data) == 0 {
		return nil, nil, ErrUnsupportedToken
	}
	dates := make([]time.Time, len(fetchKlinesDataResp.Data))
	prices := make([]float64, len(fetchKlinesDataResp.Data))
	for i, kline := range fetchKlinesDataResp.Data {
		price, err := strconv.ParseFloat(kline[4], 64)
		if err != nil {
			return nil, nil, err
		}
		date, err := strconv.ParseInt(kline[0], 10, 64)
		if err != nil {
			return nil, nil, err
		}

		prices[i] = price
		dates[i] = time.Unix(0, date*int64(time.Millisecond))
	}
	return prices, dates, nil
}

func (d *OkxDriver) FetchLast7DaysKlinesData(tokenSymbol string) ([]float64, []time.Time, error) {
	endTime := time.Now().AddDate(0, 0, -1)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
	startTime := endTime.AddDate(0, 0, -7)

	return d.FetchKlinesData(tokenSymbol, "15m", uint64(startTime.Unix()), uint64(endTime.Unix()))
}

func (d *OkxDriver) reSubscribeMarketStatByWebsocket(newSubscribeTokenSymbols []string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	subscribeTokenSymbolsFilter := lo.SliceToMap(d.subscribeTokenSymbols, func(item string) (string, struct{}) {
		return item, struct{}{}
	})
	var needSubscribeTokenSymbols []string
	for _, token := range newSubscribeTokenSymbols {
		if _, ok := subscribeTokenSymbolsFilter[token]; !ok {
			if _, ok := d.tokenSymbolToProductID[token]; ok {
				needSubscribeTokenSymbols = append(needSubscribeTokenSymbols, token)
			}
		}
	}

	if len(d.wsList) != 0 {
		lastWs := d.wsList[len(d.wsList)-1]
		if len(lastWs.tokenSymbols) < d.baseComponent.Config.Datasource.Market.Okx.SingleWsTokenLimit {
			if len(needSubscribeTokenSymbols) <= d.baseComponent.Config.Datasource.Market.Okx.SingleWsTokenLimit-len(lastWs.tokenSymbols) {
				if err := lastWs.reSubscribe(append(lastWs.tokenSymbols, needSubscribeTokenSymbols...), false); err != nil {
					return err
				}

				needSubscribeTokenSymbols = []string{}
			} else {
				if err := lastWs.reSubscribe(append(lastWs.tokenSymbols, needSubscribeTokenSymbols[0:d.baseComponent.Config.Datasource.Market.Okx.SingleWsTokenLimit-len(lastWs.tokenSymbols)]...), false); err != nil {
					return err
				}

				needSubscribeTokenSymbols = needSubscribeTokenSymbols[d.baseComponent.Config.Datasource.Market.Okx.SingleWsTokenLimit-len(lastWs.tokenSymbols):]
			}
		}
	}
	var subTokens []string
	for _, token := range needSubscribeTokenSymbols {
		subTokens = append(subTokens, token)
		if len(subTokens) == d.baseComponent.Config.Datasource.Market.Okx.SingleWsTokenLimit {
			ws, err := OkxConnectToWs(d.baseComponent, d.tokenSymbolToProductID, subTokens, d.handleMarketWebsocketEvent)
			if err != nil {
				return err
			}
			d.wsList = append(d.wsList, ws)

			subTokens = []string{}
		}
	}
	if len(subTokens) != 0 {
		ws, err := OkxConnectToWs(d.baseComponent, d.tokenSymbolToProductID, subTokens, d.handleMarketWebsocketEvent)
		if err != nil {
			return err
		}
		d.wsList = append(d.wsList, ws)
	}

	d.subscribeTokenSymbols = newSubscribeTokenSymbols
	return nil
}

type okxEvent struct {
	Arg  okxArg     `json:"arg"`
	Data [][]string `json:"data"`
}

type okxArg struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

func (d *OkxDriver) handleMarketWebsocketEvent(rawEvent []byte) {
	var event okxEvent
	err := json.Unmarshal(rawEvent, &event)
	if err != nil {
		d.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "driver": config.MarketDriverTypeOkx}).Warn("Failed to parse market event")
		return
	}

	symbol := d.tokenProductIDToSymbol[event.Arg.InstID]
	for _, p := range event.Data {
		err := func() error {
			t, err := strconv.ParseInt(p[0], 10, 64)
			if err != nil {
				return errors.Wrap(err, "failed to parse event timestamp: "+p[0])
			}

			f, err := strconv.ParseFloat(p[4], 64)
			if err != nil {
				return errors.Wrap(err, "failed to parse event price: "+p[4])
			}

			return d.wsMarketStatEventHandler(WsMarketStatEvent{
				Timestamp: t,
				Symbol:    symbol,
				Price:     f,
			})
		}()
		if err != nil {
			d.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "symbol": symbol, "driver": config.MarketDriverTypeOkx}).Warn("Failed to handle market event")
		}
	}
}
