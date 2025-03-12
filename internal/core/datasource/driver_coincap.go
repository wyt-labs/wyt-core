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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

const (
	coincapCacheID = "coincap_cache"
)

var tokenBlackList = map[string]struct{}{"LUNC": {}}

type CoincapWsInfo struct {
	baseComponent          *base.Component
	tokenSymbolToAssetInfo map[string]*CoincapAssetInfo

	tokenSymbols []string
	handler      func(rawEvent []byte)

	c                *websocket.Conn
	lock             *sync.Mutex
	isManualShutdown bool
}

func CoincapConnectToWs(baseComponent *base.Component, tokenSymbolToAssetInfo map[string]*CoincapAssetInfo, tokenSymbols []string, handler func(rawEvent []byte)) (*CoincapWsInfo, error) {
	ws := &CoincapWsInfo{
		baseComponent:          baseComponent,
		tokenSymbolToAssetInfo: tokenSymbolToAssetInfo,
		tokenSymbols:           tokenSymbols,
		handler:                handler,
		lock:                   new(sync.Mutex),
	}

	if err := ws.reSubscribe(tokenSymbols, false); err != nil {
		return nil, err
	}
	return ws, nil
}

func (w *CoincapWsInfo) reSubscribe(tokenSymbols []string, isReconnect bool) error {
	tokenSymbols = lo.Map(tokenSymbols, func(item string, index int) string {
		return strings.ToUpper(item)
	})
	var filteredTokenSymbols []string
	for _, token := range tokenSymbols {
		if _, ok := w.tokenSymbolToAssetInfo[token]; ok {
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

	endpoint := w.baseComponent.Config.Datasource.Market.Coincap.WebsocketEndpoint + "/prices?assets="
	for _, s := range w.tokenSymbols {
		endpoint += w.tokenSymbolToAssetInfo[s].ID + ","
	}
	endpoint = endpoint[:len(endpoint)-1]

	dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  20 * time.Second,
		EnableCompression: false,
	}

	c, _, err := dialer.Dial(endpoint, nil)
	if err != nil {
		return err
	}
	c.SetReadLimit(655350)
	w.c = c

	if isReconnect {
		w.baseComponent.Logger.WithFields(logrus.Fields{"driver": config.MarketDriverTypeCoincap}).Info("Reconnect to market websocket service")
	} else {
		w.baseComponent.Logger.WithFields(logrus.Fields{"driver": config.MarketDriverTypeCoincap}).Info("Connect to market websocket service")
	}

	w.baseComponent.SafeGo(func() {
		websocketkeepAlive(c, 30*time.Second)

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if _, ok := err.(*websocket.CloseError); !ok {
					w.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "driver": config.MarketDriverTypeCoincap}).Warn("Failed to handle market event")
					w.baseComponent.SafeGo(func() {
						err := util.Retry(w.baseComponent.Config.App.RetryInterval.ToDuration(), w.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
							err = w.reSubscribe(filteredTokenSymbols, true)
							return err != nil, err
						})
						if err != nil {
							w.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "driver": config.MarketDriverTypeCoincap}).Error("Failed to try reconnect to market driver by websocket")
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

func (w *CoincapWsInfo) stop() {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.c != nil {
		_ = w.c.Close()
	}
	w.isManualShutdown = true
}

type CoincapAssetInfo struct {
	ID         string  `json:"id"`
	Symbol     string  `json:"symbol"`
	Name       string  `json:"name"`
	Supply     float64 `json:"supply"`
	MaxSupply  float64 `json:"maxSupply"`
	NotHistory bool    `json:"not_history"`
}

type CoincapDataCache struct {
	UpdateTime int64

	// symbol -> CoincapAssetInfo
	AssetInfoMap map[string]*CoincapAssetInfo
}

type CoincapDriver struct {
	baseComponent             *base.Component
	systemCacheDao            *dao.SystemCacheDao
	httpClient                *resty.Client
	lock                      *sync.RWMutex
	subscribeTokenSymbols     []string
	waitSubscribeTokenSymbols []string
	wsMarketStatEventHandler  WsMarketStatEventHandler
	cache                     *CoincapDataCache
	wsList                    []*CoincapWsInfo

	// asset id -> CoincapAssetInfo
	assetInfoIDMap map[string]*CoincapAssetInfo
}

func NewCoincapDriver(baseComponent *base.Component, systemCacheDao *dao.SystemCacheDao) *CoincapDriver {
	httpClient := resty.New()
	httpClient.SetBaseURL(baseComponent.Config.Datasource.Market.Coincap.APIEndpoint)
	httpClient.SetAuthToken(baseComponent.Config.Datasource.Market.Coincap.APIKey)
	return &CoincapDriver{
		baseComponent:  baseComponent,
		systemCacheDao: systemCacheDao,
		httpClient:     httpClient,
		lock:           new(sync.RWMutex),
		assetInfoIDMap: map[string]*CoincapAssetInfo{},
	}
}

func (d *CoincapDriver) Name() string {
	return config.MarketDriverTypeCoincap
}

func (d *CoincapDriver) Config(subscribeTokenSymbols []string, wsMarketStatEventHandler WsMarketStatEventHandler) {
	d.waitSubscribeTokenSymbols = lo.Map(subscribeTokenSymbols, func(item string, index int) string {
		return strings.ToUpper(item)
	})
	d.wsMarketStatEventHandler = wsMarketStatEventHandler
}

func (d *CoincapDriver) FlushCache() error {
	if err := d.systemCacheDao.Put(d.baseComponent.BackgroundContext(), coincapCacheID, d.cache); err != nil {
		return err
	}
	return nil
}

func (d *CoincapDriver) Start() error {
	var cache CoincapDataCache
	if err := d.systemCacheDao.Get(d.baseComponent.BackgroundContext(), coincapCacheID, &cache); err != nil {
		if err != mongo.ErrNoDocuments {
			return err
		}
	}
	d.cache = &cache
	if d.cache.AssetInfoMap == nil {
		d.cache.AssetInfoMap = make(map[string]*CoincapAssetInfo)
	}
	if err := d.initCache(); err != nil {
		return err
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	if err := d.reSubscribeMarketStatByWebsocket(d.waitSubscribeTokenSymbols); err != nil {
		return errors.Wrap(err, "failed to subscribe market stat by websocket")
	}
	return nil
}

func (d *CoincapDriver) Stop() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	for _, ws := range d.wsList {
		ws.stop()
	}
	return nil
}

type coincapFetchKlinesDataResp struct {
	Data []struct {
		PriceUsd string `json:"priceUsd"`
		Time     int64  `json:"time"`
	} `json:"data"`
	Timestamp int64 `json:"timestamp"`
}

func (d *CoincapDriver) FetchKlinesData(tokenSymbol string, interval string, start uint64, end uint64) ([]float64, []time.Time, error) {
	assertInfo, ok := d.cache.AssetInfoMap[tokenSymbol]
	if !ok {
		return nil, nil, ErrUnsupportedToken
	}
	if assertInfo.NotHistory {
		return nil, nil, ErrUnsupportedToken
	}
	id := assertInfo.ID

	bar := ""
	switch interval {
	case "1d":
		bar = "d1"
	case "1w":
		bar = "d1"
	case "1M":
		bar = "d1"
	case "m15":
		bar = "m15"
	default:
		return nil, nil, errcode.ErrRequestParameter.Wrap("unsupported interval")
	}

	var fetchKlinesDataResp coincapFetchKlinesDataResp
	err := util.Retry(d.baseComponent.Config.App.RetryInterval.ToDuration(), d.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
		resp, err := d.httpClient.R().
			SetQueryParams(map[string]string{
				"interval": bar,
				"start":    strconv.Itoa(int(start * 1000)),
				"end":      strconv.Itoa(int(end * 1000)),
			}).
			Get("/assets/" + id + "/history")
		if err != nil {
			return true, err
		}
		if resp.StatusCode() != http.StatusOK {
			if resp.StatusCode() == http.StatusNotFound {
				assertInfo.NotHistory = true
				return false, ErrUnsupportedToken
			}
			return true, errors.Errorf("http request failed, code: %d, msg: %s", resp.StatusCode(), resp.String())
		}
		if err := json.Unmarshal(resp.Body(), &fetchKlinesDataResp); err != nil {
			return false, err
		}

		return false, nil
	})
	if err != nil {
		return nil, nil, err
	}

	if len(fetchKlinesDataResp.Data) == 0 {
		assertInfo.NotHistory = true
		return nil, nil, ErrUnsupportedToken
	}

	extractionInterval := 1
	switch interval {
	case "1w":
		extractionInterval = 7
	case "1M":
		extractionInterval = 30
	}

	var dates []time.Time
	var prices []float64
	for i := 0; i < len(fetchKlinesDataResp.Data); i += extractionInterval {
		kline := fetchKlinesDataResp.Data[i]
		price, err := strconv.ParseFloat(kline.PriceUsd, 64)
		if err != nil {
			return nil, nil, err
		}
		prices = append(prices, price)
		dates = append(dates, time.Unix(0, kline.Time*int64(time.Millisecond)))
	}
	return prices, dates, nil
}

func (d *CoincapDriver) FetchLast7DaysKlinesData(tokenSymbol string) ([]float64, []time.Time, error) {
	endTime := time.Now().AddDate(0, 0, -1)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
	startTime := endTime.AddDate(0, 0, -7)

	return d.FetchKlinesData(tokenSymbol, "m15", uint64(startTime.Unix()), uint64(endTime.Unix()))
}

func (d *CoincapDriver) UpdateSubscribeTokenSymbols(subscribeTokenSymbols []string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	subscribeTokenSymbols = lo.Map(subscribeTokenSymbols, func(item string, index int) string {
		return strings.ToUpper(item)
	})
	if err := d.completeCache(subscribeTokenSymbols); err != nil {
		return err
	}
	return d.reSubscribeMarketStatByWebsocket(subscribeTokenSymbols)
}

type CoincapAssetInfoResp struct {
	ID        string `json:"id"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Supply    string `json:"supply"`
	MaxSupply string `json:"maxSupply"`
}

type coincapFetchAllAssetsResp struct {
	Data      []*CoincapAssetInfoResp `json:"data"`
	Timestamp int64                   `json:"timestamp"`
}

func (d *CoincapDriver) parseFetchAssetsResp(raw []byte) ([]*CoincapAssetInfo, error) {
	var res []*CoincapAssetInfo
	var fetchAllAssetsResp coincapFetchAllAssetsResp
	if err := json.Unmarshal(raw, &fetchAllAssetsResp); err != nil {
		return nil, err
	}

	for _, data := range fetchAllAssetsResp.Data {
		supply, _ := strconv.ParseFloat(data.Supply, 64)
		maxSupply, _ := strconv.ParseFloat(data.MaxSupply, 64)
		if data.ID != "" {
			res = append(res, &CoincapAssetInfo{
				ID:        data.ID,
				Symbol:    data.Symbol,
				Name:      data.Name,
				Supply:    supply,
				MaxSupply: maxSupply,
			})
		}
	}
	return res, nil
}

func (d *CoincapDriver) fetchAllAssets() (map[string]*CoincapAssetInfo, error) {
	pageFetch := func(offset int, limit int) ([]*CoincapAssetInfo, error) {
		var res []*CoincapAssetInfo
		err := util.Retry(d.baseComponent.Config.App.RetryInterval.ToDuration(), d.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
			resp, err := d.httpClient.R().
				SetQueryParams(map[string]string{
					"offset": strconv.Itoa(offset),
					"limit":  strconv.Itoa(limit),
				}).
				Get("/assets")
			if err != nil {
				return true, err
			}
			if resp.StatusCode() != http.StatusOK {
				return true, errors.Errorf("http request failed, code: %d, msg: %s", resp.StatusCode(), resp.String())
			}
			res, err = d.parseFetchAssetsResp(resp.Body())
			if err != nil {
				return false, err
			}
			return false, nil
		})
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	offset := 0
	limit := 1000

	var res []*CoincapAssetInfo
	for {
		list, err := pageFetch(offset, limit)
		if err != nil {
			return nil, err
		}
		res = append(res, list...)
		if len(list) == 0 {
			break
		}
		offset += limit
	}

	m := map[string]*CoincapAssetInfo{}
	for _, item := range res {
		item.Symbol = strings.ToUpper(item.Symbol)
		if _, ok := m[item.Symbol]; !ok {
			m[item.Symbol] = item
		}
	}
	return m, nil
}

func (d *CoincapDriver) searchAssetInfoByTokenSymbol(tokenSymbol string) (*CoincapAssetInfo, error) {
	var res []*CoincapAssetInfo
	err := util.Retry(d.baseComponent.Config.App.RetryInterval.ToDuration(), d.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
		resp, err := d.httpClient.R().
			SetQueryParams(map[string]string{
				"search": tokenSymbol,
			}).
			Get("/assets")
		if err != nil {
			return true, err
		}
		if resp.StatusCode() != http.StatusOK {
			return true, errors.Errorf("http request failed, code: %d, msg: %s", resp.StatusCode(), resp.String())
		}
		res, err = d.parseFetchAssetsResp(resp.Body())
		if err != nil {
			return false, err
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrUnsupportedToken
	}
	return res[0], nil
}

func (d *CoincapDriver) completeCache(subscribeTokenSymbols []string) error {
	needUpdateSymbolMap := make(map[string]struct{})
	for _, symbol := range subscribeTokenSymbols {
		if _, ok := d.cache.AssetInfoMap[symbol]; !ok {
			needUpdateSymbolMap[symbol] = struct{}{}
		}
	}

	if len(needUpdateSymbolMap) != 0 {
		for s := range needUpdateSymbolMap {
			assetInfo, err := d.searchAssetInfoByTokenSymbol(s)
			if err != nil && err != ErrUnsupportedToken {
				return err
			}
			if assetInfo != nil {
				d.cache.AssetInfoMap[s] = assetInfo
			} else {
				d.cache.AssetInfoMap[s] = &CoincapAssetInfo{
					Symbol: s,
				}
			}
		}
		if err := d.systemCacheDao.Put(d.baseComponent.BackgroundContext(), coincapCacheID, d.cache); err != nil {
			return err
		}
	}

	newAssetInfoIDMap := make(map[string]*CoincapAssetInfo)
	for _, info := range d.cache.AssetInfoMap {
		newAssetInfoIDMap[info.ID] = info
	}
	d.assetInfoIDMap = newAssetInfoIDMap
	d.baseComponent.Logger.WithFields(logrus.Fields{"driver": config.MarketDriverTypeCoincap}).Infof("Finished complete coincap cacheDao, token count: %d", len(newAssetInfoIDMap))
	return nil
}

func (d *CoincapDriver) initCache() error {
	needUpdate := false
	for _, symbol := range d.waitSubscribeTokenSymbols {
		if _, ok := d.cache.AssetInfoMap[symbol]; !ok {
			needUpdate = true
			break
		}
	}
	if needUpdate {
		assetInfoMap, err := d.fetchAllAssets()
		if err != nil {
			return err
		}
		d.cache.AssetInfoMap = assetInfoMap
		for _, symbol := range d.waitSubscribeTokenSymbols {
			if _, ok := d.cache.AssetInfoMap[symbol]; !ok {
				d.cache.AssetInfoMap[symbol] = &CoincapAssetInfo{
					Symbol: "symbol",
				}
			}
		}
		if err := d.systemCacheDao.Put(d.baseComponent.BackgroundContext(), coincapCacheID, d.cache); err != nil {
			return err
		}
	}
	newAssetInfoIDMap := make(map[string]*CoincapAssetInfo)
	for _, info := range d.cache.AssetInfoMap {
		newAssetInfoIDMap[info.ID] = info
	}
	d.assetInfoIDMap = newAssetInfoIDMap
	d.baseComponent.Logger.WithFields(logrus.Fields{"driver": config.MarketDriverTypeCoincap}).Infof("Finished init coincap cacheDao, token count: %d", len(newAssetInfoIDMap))
	return nil
}

func (d *CoincapDriver) reSubscribeMarketStatByWebsocket(newSubscribeTokenSymbols []string) error {
	subscribeTokenSymbolsFilter := lo.SliceToMap(d.subscribeTokenSymbols, func(item string) (string, struct{}) {
		return item, struct{}{}
	})
	var needSubscribeTokenSymbols []string
	for _, token := range newSubscribeTokenSymbols {
		if _, ok := subscribeTokenSymbolsFilter[token]; !ok {
			if _, ok := d.cache.AssetInfoMap[token]; ok {
				needSubscribeTokenSymbols = append(needSubscribeTokenSymbols, token)
			}
		}
	}

	if len(d.wsList) != 0 {
		lastWs := d.wsList[len(d.wsList)-1]
		if len(lastWs.tokenSymbols) < d.baseComponent.Config.Datasource.Market.Coincap.SingleWsTokenLimit {
			if len(needSubscribeTokenSymbols) <= d.baseComponent.Config.Datasource.Market.Coincap.SingleWsTokenLimit-len(lastWs.tokenSymbols) {
				if err := lastWs.reSubscribe(append(lastWs.tokenSymbols, needSubscribeTokenSymbols...), false); err != nil {
					return err
				}

				needSubscribeTokenSymbols = []string{}
			} else {
				if err := lastWs.reSubscribe(append(lastWs.tokenSymbols, needSubscribeTokenSymbols[0:d.baseComponent.Config.Datasource.Market.Coincap.SingleWsTokenLimit-len(lastWs.tokenSymbols)]...), false); err != nil {
					return err
				}

				needSubscribeTokenSymbols = needSubscribeTokenSymbols[d.baseComponent.Config.Datasource.Market.Coincap.SingleWsTokenLimit-len(lastWs.tokenSymbols):]
			}
		}
	}
	var subTokens []string
	for _, token := range needSubscribeTokenSymbols {
		subTokens = append(subTokens, token)
		if len(subTokens) == d.baseComponent.Config.Datasource.Market.Coincap.SingleWsTokenLimit {
			ws, err := CoincapConnectToWs(d.baseComponent, d.cache.AssetInfoMap, subTokens, d.handleMarketWebsocketEvent)
			if err != nil {
				return err
			}
			d.wsList = append(d.wsList, ws)

			subTokens = []string{}
		}
	}
	if len(subTokens) != 0 {
		ws, err := CoincapConnectToWs(d.baseComponent, d.cache.AssetInfoMap, subTokens, d.handleMarketWebsocketEvent)
		if err != nil {
			return err
		}
		d.wsList = append(d.wsList, ws)
	}

	d.subscribeTokenSymbols = newSubscribeTokenSymbols
	return nil
}

func (d *CoincapDriver) handleMarketWebsocketEvent(rawEvent []byte) {
	now := time.Now().UnixMilli()

	event := map[string]string{}
	err := json.Unmarshal(rawEvent, &event)
	if err != nil {
		d.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "driver": config.MarketDriverTypeCoincap}).Warn("Failed to parse market event")
		return
	}

	for id, p := range event {
		err := func() error {
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return errors.Wrap(err, "failed to parse event price: "+p)
			}
			assertInfo, ok := d.assetInfoIDMap[id]
			if !ok {
				return nil
			}
			if _, ok := tokenBlackList[assertInfo.Symbol]; ok {
				return nil
			}
			return d.wsMarketStatEventHandler(WsMarketStatEvent{
				Timestamp:   now,
				Symbol:      assertInfo.Symbol,
				Price:       f,
				Supply:      assertInfo.Supply,
				TotalSupply: assertInfo.MaxSupply,
			})
		}()
		if err != nil {
			d.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "token_id": id, "driver": config.MarketDriverTypeCoincap}).Warn("Failed to handle market event")
		}
	}
}
