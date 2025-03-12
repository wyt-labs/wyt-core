package datasource

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
)

var ErrUnsupportedToken = errors.New("unsupported token")

const (
	quoteAsset = "USDT"

	projectMarketInfoCacheID = "project_market_info_cache"
)

type WsMarketStatEvent struct {
	Timestamp   int64
	Symbol      string
	Price       float64
	Supply      float64
	TotalSupply float64
}

type WsMarketStatEventHandler func(event WsMarketStatEvent) error

type MarketDriver interface {
	Name() string

	Config(subscribeTokenSymbols []string, wsMarketStatEventHandler WsMarketStatEventHandler)

	Start() error

	Stop() error

	FetchLast7DaysKlinesData(tokenSymbol string) ([]float64, []time.Time, error)

	FetchKlinesData(tokenSymbol string, interval string, start uint64, end uint64) ([]float64, []time.Time, error)

	UpdateSubscribeTokenSymbols(subscribeTokenSymbols []string) error

	FlushCache() error
}

type ProjectMarketInfo struct {
	ID                            string
	Symbol                        string
	Price                         float64
	UpdateTimestamp               int64
	CirculatingSupply             float64
	TotalSupply                   float64
	MarketCap                     uint64
	Last7DaysKlinesDataPictureURL string
	Rank                          int
}

type ProjectMarketInfoCache struct {
	UpdateTime int64

	// only cache view info
	ProjectMarketInfoMap map[string]*ProjectMarketInfo
}

type SortType uint32

const (
	ProjectMarketInfosSortByMarketcap SortType = iota
	ProjectMarketInfosSortByPrice
)

type ProjectMarketInfosSort struct {
	List     []ProjectMarketInfo
	SortType SortType
	IsAsc    bool
}

func (p ProjectMarketInfosSort) Len() int { return len(p.List) }

func (p ProjectMarketInfosSort) Swap(i, j int) { p.List[i], p.List[j] = p.List[j], p.List[i] }

func (p ProjectMarketInfosSort) Less(i, j int) bool {
	var less bool
	switch p.SortType {
	case ProjectMarketInfosSortByPrice:
		less = p.List[i].Price < p.List[j].Price
	default:
		less = p.List[i].MarketCap < p.List[j].MarketCap
	}
	if p.IsAsc {
		return less
	}
	return !less
}

func (p ProjectMarketInfosSort) Sort() {
	sort.Sort(p)
}

type ProjectListElementTokenomicsInfo struct {
	TokenSymbol       string  `json:"token_symbol" bson:"token_symbol"`
	CirculatingSupply float64 `json:"circulating_supply" bson:"circulating_supply"`
	TotalSupply       float64 `json:"total_supply" bson:"total_supply"`
}

type ProjectListElement struct {
	ID         primitive.ObjectID               `json:"id" bson:"_id"`
	Tokenomics ProjectListElementTokenomicsInfo `json:"tokenomics" bson:"tokenomics"`
}

type Market struct {
	baseComponent  *base.Component
	projectDao     *dao.ProjectDao
	fileSystemDao  *dao.FileSystemDao
	systemCacheDao *dao.SystemCacheDao
	drivers        []MarketDriver
	lock           *sync.RWMutex

	// id -> ProjectMarketInfo
	// update by websocket event

	realTimeProjectMarketInfoMap map[string]*ProjectMarketInfo

	// id -> ProjectMarketInfo
	// regular update from realTimeProjectMarketInfoMap
	viewProjectMarketInfoMap map[string]*ProjectMarketInfo

	viewProjectMarketInfosSortByMarketCap []ProjectMarketInfo
	currentSubscribeSymbols               []string
	tokenSymbolToProjectIDMap             map[string]string

	marketDataIsInit bool
	updateViewIsInit bool
}

func NewMarket(baseComponent *base.Component, projectDao *dao.ProjectDao, fileSystemDao *dao.FileSystemDao, systemCacheDao *dao.SystemCacheDao) (*Market, error) {
	var drivers []MarketDriver

	fmt.Println(baseComponent.Config.Datasource.Market.MarketDrivers)
	for _, driverName := range lo.Uniq(baseComponent.Config.Datasource.Market.MarketDrivers) {
		var d MarketDriver
		switch driverName {
		case config.MarketDriverTypeCoincap:
			d = NewCoincapDriver(baseComponent, systemCacheDao)
		case config.MarketDriverTypeBinance:
			d = NewBinanceDriver(baseComponent)
		case config.MarketDriverTypeOkx:
			d = NewOkxDriver(baseComponent, systemCacheDao)
		default:
			return nil, errors.Errorf("unsupported market driver: %s", driverName)
		}
		drivers = append(drivers, d)
	}

	c := &Market{
		baseComponent:                baseComponent,
		projectDao:                   projectDao,
		fileSystemDao:                fileSystemDao,
		systemCacheDao:               systemCacheDao,
		drivers:                      drivers,
		lock:                         new(sync.RWMutex),
		realTimeProjectMarketInfoMap: map[string]*ProjectMarketInfo{},
		viewProjectMarketInfoMap:     map[string]*ProjectMarketInfo{},
		tokenSymbolToProjectIDMap:    map[string]string{},
	}
	baseComponent.RegisterLifecycleHook(c)
	return c, nil
}

func (c *Market) Start() error {
	if c.baseComponent.Config.Datasource.Market.Disable {
		return nil
	}

	c.baseComponent.SafeGo(func() {
		if err := c.loadAllProjects(); err != nil {
			c.baseComponent.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to load all project info")
			c.baseComponent.ComponentShutdown()
			return
		}
		c.baseComponent.Logger.Infof("Load all project info, count: %d, available count: %d", len(c.viewProjectMarketInfoMap), len(c.currentSubscribeSymbols))

		for _, driver := range c.drivers {
			driver.Config(c.currentSubscribeSymbols, c.handleMarketWebsocketEvent)
			if err := driver.Start(); err != nil {
				c.baseComponent.Logger.WithFields(logrus.Fields{
					"err":    err,
					"driver": driver.Name(),
				}).Error("Failed to start market data driver")
				c.baseComponent.ComponentShutdown()
				return
			}
			c.baseComponent.Logger.Infof("Finished start market datasource driver[%s]", driver.Name())
		}

		c.baseComponent.SafeGoPersistentTask(c.regularUpdateViewInfo)

		var cache ProjectMarketInfoCache
		if err := c.systemCacheDao.Get(c.baseComponent.BackgroundContext(), projectMarketInfoCacheID, &cache); err != nil {
			c.updateLast7DaysKlinesData()
		} else {
			now := time.Now()
			start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			tm := time.Unix(cache.UpdateTime, 0)
			if tm.Before(start) {
				c.updateLast7DaysKlinesData()
			} else {
				for _, info := range c.viewProjectMarketInfoMap {
					if s, ok := cache.ProjectMarketInfoMap[info.ID]; ok {
						info.Last7DaysKlinesDataPictureURL = s.Last7DaysKlinesDataPictureURL
					}
				}
			}
		}

		_, err := c.baseComponent.Cron.AddFunc(c.baseComponent.Config.Datasource.Market.Last7DaysKlinesDataRefreshCron, c.updateLast7DaysKlinesData)
		if err != nil {
			c.baseComponent.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to add Last7DaysKlinesDataRefreshCron task")
			c.baseComponent.ComponentShutdown()
			return
		}
	})
	return nil
}

func (c *Market) Stop() error {
	if c.baseComponent.Config.Datasource.Market.Disable {
		return nil
	}
	for _, driver := range c.drivers {
		if err := driver.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// program startup time
func (c *Market) loadAllProjects() error {
	// find all projects id and token symbol

	var list []*ProjectListElement
	_, err := c.projectDao.CustomList(c.baseComponent.BackgroundContext(), false, 0, 0, nil, nil, &list)
	if err != nil {
		return err
	}

	for _, element := range list {
		tokenSymbol := strings.ToUpper(element.Tokenomics.TokenSymbol)
		id := element.ID.Hex()
		if tokenSymbol != "" {
			c.currentSubscribeSymbols = append(c.currentSubscribeSymbols, tokenSymbol)
			c.tokenSymbolToProjectIDMap[tokenSymbol] = id
		}

		c.realTimeProjectMarketInfoMap[id] = &ProjectMarketInfo{
			ID:                id,
			Symbol:            tokenSymbol,
			CirculatingSupply: element.Tokenomics.CirculatingSupply,
			TotalSupply:       element.Tokenomics.TotalSupply,
		}
		c.viewProjectMarketInfoMap[id] = &ProjectMarketInfo{
			ID:                id,
			Symbol:            tokenSymbol,
			CirculatingSupply: element.Tokenomics.CirculatingSupply,
		}
	}

	return nil
}

func (c *Market) handleMarketWebsocketEvent(event WsMarketStatEvent) error {
	id, ok := c.tokenSymbolToProjectIDMap[event.Symbol]
	if ok {
		info := c.realTimeProjectMarketInfoMap[id]
		info.Price = event.Price
		info.UpdateTimestamp = event.Timestamp
		if event.Supply != 0 {
			info.CirculatingSupply = event.Supply
		}
		if event.TotalSupply != 0 {
			info.TotalSupply = event.TotalSupply
		} else {
			if info.TotalSupply == 0 {
				info.TotalSupply = info.CirculatingSupply
			}
		}
		info.MarketCap = uint64(info.Price * info.CirculatingSupply)
		if !c.marketDataIsInit {
			c.baseComponent.Logger.Info("Received first market event")
			c.marketDataIsInit = true
		}
	} else {
		return errors.Errorf("received unsubscribed market event, symbol: %s", event.Symbol)
	}
	return nil
}

// copy market info from real-time data to view data
// to reduce the frequency of lock use
func (c *Market) regularUpdateViewInfo() {
	c.baseComponent.Logger.Info("Start regular update market data view info")
	ticker := time.NewTicker(c.baseComponent.Config.Datasource.Market.MarketDataRefreshInterval.ToDuration())
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			func() {
				c.lock.Lock()
				sourceMap := c.realTimeProjectMarketInfoMap
				targetMap := c.viewProjectMarketInfoMap
				var viewProjectMarketInfoList []ProjectMarketInfo
				for id, sourceInfo := range sourceMap {
					if targetInfo, ok := targetMap[id]; ok {
						targetInfo.UpdateTimestamp = sourceInfo.UpdateTimestamp
						targetInfo.Price = sourceInfo.Price
						targetInfo.MarketCap = sourceInfo.MarketCap
						targetInfo.CirculatingSupply = sourceInfo.CirculatingSupply
						targetInfo.TotalSupply = sourceInfo.TotalSupply
						viewProjectMarketInfoList = append(viewProjectMarketInfoList, *targetInfo)
					}
				}
				ProjectMarketInfosSort{
					List:     viewProjectMarketInfoList,
					SortType: ProjectMarketInfosSortByMarketcap,
					IsAsc:    false,
				}.Sort()
				for i, info := range viewProjectMarketInfoList {
					if targetInfo, ok := targetMap[info.ID]; ok {
						targetInfo.Rank = i + 1
					}
					viewProjectMarketInfoList[i].Rank = i + 1
				}
				c.viewProjectMarketInfosSortByMarketCap = viewProjectMarketInfoList
				c.lock.Unlock()
			}()
			if !c.updateViewIsInit && c.marketDataIsInit {
				c.baseComponent.Logger.Info("Update init market data view info")
				c.updateViewIsInit = true
			}
		case <-c.baseComponent.Ctx.Done():
			c.baseComponent.Logger.Info("Stop regular update market data view info")
			return
		}
	}
}

func (c *Market) generateFileURL(bucketName string, id string) string {
	return fmt.Sprintf("http://%s/api/v1/fs/files/%s/%s", c.baseComponent.Config.App.AccessDomain, bucketName, id)
}

func (c *Market) updateLast7DaysKlinesDataPicture(info *ProjectMarketInfo) error {
	picture, err := c.generateLast7DaysKlinesDataPicture(info.Symbol)
	if err != nil {
		return err
	}
	pictureName := fmt.Sprintf("%s_7days_klines_data.svg", info.Symbol)
	pictureID := fmt.Sprintf("%s_7days_klines_data", info.ID)
	err = c.fileSystemDao.UploadWithID(c.baseComponent.BackgroundContext(), pictureID, entity.BucketTypeMisc.String(), pictureName, picture, "")
	if err != nil {
		return err
	}
	info.Last7DaysKlinesDataPictureURL = c.generateFileURL(entity.BucketTypeMisc.String(), pictureID)
	return nil
}

func (c *Market) updateLast7DaysKlinesData() {
	for _, info := range c.viewProjectMarketInfoMap {
		if info.Symbol == "" {
			continue
		}

		err := c.updateLast7DaysKlinesDataPicture(info)
		if err != nil {
			if err != ErrUnsupportedToken {
				c.baseComponent.Logger.WithFields(logrus.Fields{
					"err":    err,
					"symbol": info.Symbol,
				}).Warn("Failed to generate Last7DaysKlinesDataPicture")
			}
		}
	}
	c.baseComponent.Logger.Infof("Update last 7 days klines data")

	for _, driver := range c.drivers {
		if err := driver.FlushCache(); err != nil {
			c.baseComponent.Logger.WithFields(logrus.Fields{
				"err":    err,
				"driver": driver.Name(),
			}).Warn("Failed to flush market driver cache")
		}
	}

	if err := c.systemCacheDao.Put(c.baseComponent.BackgroundContext(), projectMarketInfoCacheID, &ProjectMarketInfoCache{
		UpdateTime:           time.Now().Unix(),
		ProjectMarketInfoMap: c.viewProjectMarketInfoMap,
	}); err != nil {
		c.baseComponent.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Warn("Failed to update projectMarketCacheInfo")
	}
}

func (c *Market) generateLast7DaysKlinesDataPicture(tokenSymbol string) (*bytes.Buffer, error) {
	var prices []float64
	var dates []time.Time
	var err error
	for _, driver := range c.drivers {
		prices, dates, err = driver.FetchLast7DaysKlinesData(tokenSymbol)
		if err != nil {
			if err != ErrUnsupportedToken {
				return nil, err
			}
		} else {
			break
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "can not load prices, symbol[%s]", tokenSymbol)
	}

	color := chart.ColorGreen
	if prices[len(prices)-1] < prices[0] {
		color = chart.ColorRed
	}

	graph := chart.Chart{
		Width:  656,
		Height: 192,
		XAxis: chart.XAxis{
			Style: chart.Style{
				Hidden: true,
			},
			GridMajorStyle: chart.Style{
				Hidden:      true,
				StrokeWidth: 0,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				Hidden: true,
			},
			GridMajorStyle: chart.Style{
				Hidden:      true,
				StrokeWidth: 0,
			},
		},
		YAxisSecondary: chart.YAxis{
			Style: chart.Style{
				Hidden: true,
			},
			GridMajorStyle: chart.Style{
				Hidden:      true,
				StrokeWidth: 0,
			},
		},
		Background: chart.Style{
			Hidden: false,
			Padding: chart.Box{
				Top:    20,
				Left:   20,
				Bottom: 20,
				Right:  20,
			},
			StrokeColor: drawing.Color{R: 1, G: 1, B: 1, A: 0},
			FillColor:   drawing.Color{R: 1, G: 1, B: 1, A: 0},
		},
		Canvas: chart.Style{
			Hidden:      false,
			StrokeColor: drawing.Color{R: 1, G: 1, B: 1, A: 0},
			FillColor:   drawing.Color{R: 1, G: 1, B: 1, A: 0},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Style: chart.Style{
					Hidden:      false,
					StrokeColor: color,
					StrokeWidth: 2.0,
					DotColor:    color,
					DotWidth:    1.0,
				},
				XValues: dates,
				YValues: prices,
			},
		},
	}
	var buffer = &bytes.Buffer{}
	if err := graph.Render(chart.SVG, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

func (c *Market) AddProject(id string, tokenSymbol string, circulatingSupply float64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	copyMap := func(s map[string]*ProjectMarketInfo) map[string]*ProjectMarketInfo {
		return lo.MapEntries(s, func(key string, value *ProjectMarketInfo) (string, *ProjectMarketInfo) {
			return key, &ProjectMarketInfo{
				ID:                            value.ID,
				Symbol:                        value.Symbol,
				Price:                         value.Price,
				UpdateTimestamp:               value.UpdateTimestamp,
				CirculatingSupply:             value.CirculatingSupply,
				MarketCap:                     value.MarketCap,
				Last7DaysKlinesDataPictureURL: value.Last7DaysKlinesDataPictureURL,
			}
		})
	}
	tokenSymbol = strings.ToUpper(tokenSymbol)
	info := &ProjectMarketInfo{
		ID:                id,
		Symbol:            tokenSymbol,
		CirculatingSupply: circulatingSupply,
	}

	if tokenSymbol != "" {
		if lo.Contains(c.currentSubscribeSymbols, tokenSymbol) {
			return nil
		}

		c.currentSubscribeSymbols = append(c.currentSubscribeSymbols, tokenSymbol)
		c.tokenSymbolToProjectIDMap[tokenSymbol] = id
		err := c.updateLast7DaysKlinesDataPicture(info)
		if err != nil {
			if err != ErrUnsupportedToken {
				c.baseComponent.Logger.WithFields(logrus.Fields{
					"err":    err,
					"symbol": info.Symbol,
				}).Warn("Failed to generate Last7DaysKlinesDataPicture")
			}
		}
	}

	newRealTimeProjectMarketInfoMap := copyMap(c.realTimeProjectMarketInfoMap)
	newRealTimeProjectMarketInfoMap[id] = &ProjectMarketInfo{
		ID:                            id,
		Symbol:                        tokenSymbol,
		CirculatingSupply:             circulatingSupply,
		Last7DaysKlinesDataPictureURL: info.Last7DaysKlinesDataPictureURL,
	}
	c.realTimeProjectMarketInfoMap = newRealTimeProjectMarketInfoMap

	newViewProjectMarketInfoMap := copyMap(c.viewProjectMarketInfoMap)
	newViewProjectMarketInfoMap[id] = &ProjectMarketInfo{
		ID:                            id,
		Symbol:                        tokenSymbol,
		CirculatingSupply:             circulatingSupply,
		Last7DaysKlinesDataPictureURL: info.Last7DaysKlinesDataPictureURL,
	}
	c.viewProjectMarketInfoMap = newViewProjectMarketInfoMap

	err := c.systemCacheDao.Put(c.baseComponent.BackgroundContext(), projectMarketInfoCacheID, &ProjectMarketInfoCache{
		UpdateTime:           time.Now().Unix(),
		ProjectMarketInfoMap: newViewProjectMarketInfoMap,
	})
	if err != nil {
		c.baseComponent.Logger.WithFields(logrus.Fields{
			"err":    err,
			"symbol": info.Symbol,
		}).Warn("Failed to update projectMarketCacheInfo")
	}

	for _, driver := range c.drivers {
		if err := driver.UpdateSubscribeTokenSymbols(c.currentSubscribeSymbols); err != nil {
			return err
		}
	}
	return nil
}

func (c *Market) FindProjectsByMarketCapSort(min uint64, max uint64) ([]ProjectMarketInfo, error) {
	s := c.viewProjectMarketInfosSortByMarketCap
	if min > uint64(len(s)) {
		return nil, errors.New("not found projects")
	}
	if min == 0 {
		min = 1
	}
	if max > uint64(len(s)) || max == 0 {
		max = uint64(len(s))
	}

	return s[min-1 : max], nil
}

func (c *Market) FindProjectsMarketInfo(ids []string) []ProjectMarketInfo {
	m := c.viewProjectMarketInfoMap
	var res []ProjectMarketInfo
	for _, id := range ids {
		info, ok := m[id]
		if ok {
			res = append(res, *info)
		} else {
			res = append(res, ProjectMarketInfo{
				ID: id,
			})
		}
	}

	return res
}

func (c *Market) FindProjectMarketInfo(id string) ProjectMarketInfo {
	info, ok := c.viewProjectMarketInfoMap[id]
	if ok {
		return *info
	}
	return ProjectMarketInfo{
		ID: id,
	}
}

func (c *Market) FetchKlinesData(id string, interval string, start uint64, end uint64) ([]float64, []time.Time, error) {
	info, ok := c.viewProjectMarketInfoMap[id]
	if !ok || info.Symbol == "" {
		return nil, nil, nil
	}

	var prices []float64
	var dates []time.Time
	var err error
	for _, driver := range c.drivers {
		prices, dates, err = driver.FetchKlinesData(info.Symbol, interval, start, end)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, nil, errors.Wrapf(err, "can not load prices, symbol[%s]", info.Symbol)
	}
	return prices, dates, nil
}

func websocketkeepAlive(c *websocket.Conn, timeout time.Duration) {
	ticker := time.NewTicker(timeout)

	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer ticker.Stop()
		for {
			deadline := time.Now().Add(10 * time.Second)
			err := c.WriteControl(websocket.PingMessage, []byte{}, deadline)
			if err != nil {
				return
			}
			<-ticker.C
			if time.Since(lastResponse) > timeout {
				_ = c.Close()
				return
			}
		}
	}()
}
