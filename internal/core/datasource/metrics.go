package datasource

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

type ProjectMetricsInfoActiveUser struct {
	Time  time.Time
	Value int64
}

type ProjectMetricsInfo struct {
	Name        string
	ActiveUsers []ProjectMetricsInfoActiveUser
}

func (m *ProjectMetricsInfo) FetchActiveUsersByTimeRange(interval string, start uint64, end uint64) ([]int64, []time.Time, error) {
	var nums []int64
	var dates []time.Time
	var extractionInterval = 1
	switch interval {
	case "1d":
		extractionInterval = extractionInterval * 1
	case "1w":
		extractionInterval = extractionInterval * 7
	case "1M":
		extractionInterval = extractionInterval * 30
	default:
		return nil, nil, errcode.ErrRequestParameter.Wrap("unsupported interval")
	}
	startTime := time.Unix(int64(start), 0)
	if startTime.Hour() != 0 || startTime.Minute() != 0 || startTime.Second() != 0 {
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day()+1, 0, 0, 0, 0, startTime.Location())
	}
	endTime := time.Unix(int64(end), 0)
	if endTime.Hour() != 0 || endTime.Minute() != 0 || endTime.Second() != 0 {
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day()-1, 0, 0, 0, 0, endTime.Location())
	}

	if len(m.ActiveUsers) != 0 {
		if m.ActiveUsers[0].Time.After(endTime) || m.ActiveUsers[len(m.ActiveUsers)-1].Time.Before(startTime) {
			return nums, dates, nil
		}
	}

	for _, e := range m.ActiveUsers {
		if !e.Time.After(endTime) && !e.Time.Before(startTime) {
			nums = append(nums, e.Value)
			dates = append(dates, e.Time)
		}
		if e.Time.After(endTime) {
			break
		}
	}
	if extractionInterval != 1 {
		var filteredNums []int64
		var filteredDates []time.Time
		for i := 0; i < len(dates); i += extractionInterval {
			filteredNums = append(filteredNums, nums[i])
			filteredDates = append(filteredDates, dates[i])
		}
		return filteredNums, filteredDates, nil
	}

	return nums, dates, nil
}

type Metrics struct {
	baseComponent *base.Component
	httpClient    *resty.Client
	lock          *sync.RWMutex

	// name -> MetricsInfo
	projectMetricsInfoMap map[string]*ProjectMetricsInfo
}

func NewMetrics(baseComponent *base.Component) (*Metrics, error) {
	httpClient := resty.New()
	httpClient.SetBaseURL(baseComponent.Config.Datasource.Metric.TokenTerminal.APIEndpoint)
	httpClient.SetAuthToken(baseComponent.Config.Datasource.Metric.TokenTerminal.APIKey)
	m := &Metrics{
		baseComponent:         baseComponent,
		httpClient:            httpClient,
		lock:                  new(sync.RWMutex),
		projectMetricsInfoMap: make(map[string]*ProjectMetricsInfo),
	}
	baseComponent.RegisterLifecycleHook(m)
	return m, nil
}

func (m *Metrics) Start() error {
	if m.baseComponent.Config.Datasource.Metric.Disable {
		return nil
	}

	m.baseComponent.SafeGo(func() {
		if err := m.fetchActiveUsersData(); err != nil {
			m.baseComponent.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to fetch active user data")
			m.baseComponent.ComponentShutdown()
		}
		m.baseComponent.Logger.Infof("Fetch active user data, project count: %d", len(m.projectMetricsInfoMap))
		_, err := m.baseComponent.Cron.AddFunc(m.baseComponent.Config.Datasource.Metric.ActiveUserDataRefreshCron, func() {
			if err := m.fetchActiveUsersData(); err != nil {
				m.baseComponent.Logger.WithFields(logrus.Fields{
					"err": err,
				}).Error("Failed to fetch active user data")
			}
		})
		if err != nil {
			m.baseComponent.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to add ActiveUserDataRefreshCron task")
			m.baseComponent.ComponentShutdown()
			return
		}
	})

	return nil
}

func (m *Metrics) Stop() error {
	return nil
}

type tokenTerminalFetchActiveUsersDataResp struct {
	Data []tokenTerminalFetchActiveUsersDataRespElement `json:"data"`
}

type tokenTerminalFetchActiveUsersDataRespElement struct {
	ProjectName string `json:"project_name"`
	Timestamp   string `json:"timestamp"`
	Value       int64  `json:"value"`
}

func (m *Metrics) fetchActiveUsersData() error {
	var fetchActiveUsersDataResp tokenTerminalFetchActiveUsersDataResp
	err := util.Retry(m.baseComponent.Config.App.RetryInterval.ToDuration(), m.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
		resp, err := m.httpClient.R().
			SetQueryParams(map[string]string{
				"metric_id":         "user_dau",
				"limit":             "500",
				"interval":          "365d",
				"historical":        "true",
				"market_sector_ids": "",
			}).
			Get("/internal/metrics/top-n-projects-metrics")
		if err != nil {
			return true, err
		}
		if resp.StatusCode() != http.StatusOK {
			return true, errors.Errorf("http request failed, code: %d, msg: %s", resp.StatusCode(), resp.String())
		}
		if err := json.Unmarshal(resp.Body(), &fetchActiveUsersDataResp); err != nil {
			return false, err
		}

		if len(fetchActiveUsersDataResp.Data) == 0 {
			return true, errors.New("cannot fetch active user data")
		}

		return false, nil
	})
	if err != nil {
		return err
	}
	projectMetricsInfoMap := make(map[string]*ProjectMetricsInfo)
	for _, e := range fetchActiveUsersDataResp.Data {
		info, ok := projectMetricsInfoMap[e.ProjectName]
		if !ok {
			info = &ProjectMetricsInfo{
				Name:        e.ProjectName,
				ActiveUsers: []ProjectMetricsInfoActiveUser{},
			}
			projectMetricsInfoMap[e.ProjectName] = info
		}

		t, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			return err
		}
		info.ActiveUsers = append(info.ActiveUsers, ProjectMetricsInfoActiveUser{
			Time:  t,
			Value: e.Value,
		})
	}

	m.projectMetricsInfoMap = projectMetricsInfoMap
	return nil
}

func (m *Metrics) FindProjectMetricsInfo(projectName string) ProjectMetricsInfo {
	info, ok := m.projectMetricsInfoMap[projectName]
	if ok {
		return *info
	}
	return ProjectMetricsInfo{
		Name: projectName,
	}
}
