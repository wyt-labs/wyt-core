package datapuller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wyt-labs/wyt-core/internal/core/component/datapuller/model"
	"github.com/wyt-labs/wyt-core/internal/core/component/httpclient"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/pkg/basic"
	"github.com/wyt-labs/wyt-core/pkg/cache"
)

func init() {
	basic.RegisterComponents(NewMetabaseDataSource, NewPumpDataService)
}

type MetabaseDataSource struct {
	baseComponent *base.Component
	apiClient     *httpclient.Client
}

func NewMetabaseDataSource(baseComponent *base.Component) (*MetabaseDataSource, error) {
	client, err := httpclient.NewHttpClient(
		httpclient.WithBaseURL(baseComponent.Config.Backends.MetabaseURL),
	)
	if err != nil {
		baseComponent.Logger.WithField("err", err).Error("failed to create http client")
		return nil, err
	}
	return &MetabaseDataSource{
		baseComponent: baseComponent,
		apiClient:     client,
	}, nil
}

func (m *MetabaseDataSource) Auth(username, pwd string) (string, error) {
	val, exist := cache.GetFromMemCache[string](m.baseComponent.MemCache, "auth", username)
	if exist {
		return val, nil
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	plStr := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, pwd)
	payload := strings.NewReader(plStr)
	resp, err := m.apiClient.PostV2("/api/session", payload, headers, nil)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return "", err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(resp, &data); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal auth response")
		return "", err
	}
	token, ok := data["id"].(string)
	if !ok {
		m.baseComponent.Logger.Error("failed to get auth token")
		return "", err
	}
	cache.PutToMemCache(m.baseComponent.MemCache, "auth", username, token)
	return token, nil
}

// DailyLaunchedTokenInfo
// 过去duration天, 代币创建以及上Raydium的数量,UTC或者CST时间
func (m *MetabaseDataSource) DailyLaunchedTokenInfo(duration int, timezone string) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	if duration < 7 {
		duration = 7
	}
	if timezone == "" {
		timezone = "UTC"
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	var id string
	if timezone == "UTC" {
		id = "bb79b15e-1245-4166-96f0-c6baaa71567c"
	} else if timezone == "CST" {
		id = "a3066d17-b2fc-4d12-bf4a-92f81ab71d66"
	}
	plStr := fmt.Sprintf(`
	{
		"ignore_cache": false,
		"collection_preview": false,
		"parameters": [
			{
				"id": "%s",
				"type": "number/=",
				"value": [
					"%d"
				],
			"target": [
				"variable",
				[
					"template-tag",
					"days"
				]
			]
			}
		]
	}`, id, duration)
	payload := strings.NewReader(plStr)
	var resp []byte
	if timezone == "UTC" || timezone == "" {
		resp, err = m.apiClient.PostV2("/api/card/106/query", payload, headers, nil)
	} else if timezone == "CST" {
		resp, err = m.apiClient.PostV2("/api/card/115/query", payload, headers, nil)
	}
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to create token")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal create token response")
		return nil, err
	}
	return &ret, nil
}

// 过去duration天,新Token发射的时间分布（按半小时）,UTC或者CST时间
func (m *MetabaseDataSource) LaunchedTokenTimeDistribution(duration int, timezone string) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	if duration < 7 {
		duration = 7
	}
	if timezone == "" {
		timezone = "UTC"
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	var id string
	if timezone == "UTC" {
		id = "4afba805-4047-477a-b1aa-399e06f5f5e4"
	} else if timezone == "CST" {
		id = "a9330250-fca8-46b2-88ed-0fef40ef981e"
	}
	plStr := fmt.Sprintf(`
	{
		"ignore_cache": false,
		"collection_preview": false,
		"parameters": [
			{
				"id": "%s",
				"type": "number/=",
				"value": [
					"%d"
				],
			"target": [
				"variable",
				[
					"template-tag",
					"days"
				]
			]
			}
		]
	}`, id, duration)
	payload := strings.NewReader(plStr)
	var resp []byte
	if timezone == "UTC" || timezone == "" {
		resp, err = m.apiClient.PostV2("/api/card/107/query", payload, headers, nil)
	} else if timezone == "CST" {
		resp, err = m.apiClient.PostV2("/api/card/114/query", payload, headers, nil)
	}
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to launched token time distribution")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal launched token time distribution response")
		return nil, err
	}
	return &ret, nil
}

// 过去duration天,每日交易量,UTC或者CST时间
func (m *MetabaseDataSource) DailyTradeCounts(duration int, timezone string) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	if duration < 7 {
		duration = 7
	}
	if timezone == "" {
		timezone = "UTC"
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	var id string
	if timezone == "UTC" {
		id = "59c5518a-8697-4d92-8531-4ba301e6ab85"
	} else if timezone == "CST" {
		id = "59c5518a-8697-4d92-8531-4ba301e6ab85"
	}
	plStr := fmt.Sprintf(`
	{
		"ignore_cache": false,
		"collection_preview": false,
		"parameters": [
			{
				"id": "%s",
				"type": "number/=",
				"value": [
					"%d"
				],
			"target": [
				"variable",
				[
					"template-tag",
					"days"
				]
			]
			}
		]
	}`, id, duration)
	payload := strings.NewReader(plStr)
	var resp []byte
	if timezone == "UTC" || timezone == "" {
		resp, err = m.apiClient.PostV2("/api/card/108/query", payload, headers, nil)
	} else if timezone == "CST" {
		resp, err = m.apiClient.PostV2("/api/card/116/query", payload, headers, nil)
	}
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to daily trade counts")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal daily trade counts response")
		return nil, err
	}
	return &ret, nil
}

// trader总览信息, 74tYkMYmwnmi44PQo6L6QpkxmdNTdX5AZaiKMrMAncwW
func (m *MetabaseDataSource) TraderOverview(trader string) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	id := "cd81bc1d-4a11-4375-84ea-e018c5d9ddef"
	plStr := fmt.Sprintf(`
		{
			"ignore_cache": false,
			"collection_preview": false,
			"parameters": [
				{
					"id": "%s",
					"type": "category",
					"value": "%s",
					"target": [
						"variable",
						[
							"template-tag",
							"trader"
						]
					]
				}
			]
		}`, id, trader)
	payload := strings.NewReader(plStr)
	resp, err := m.apiClient.PostV2("/api/card/110/query", payload, headers, nil)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to trader overview")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal trader overview response")
		return nil, err
	}
	return &ret, nil
}

// trader V2总览信息, 9YqDWbEpKME1pjM91FYDMYfuTbFCqmWT8GLitfe19Ngr
func (m *MetabaseDataSource) TraderOverviewV2(trader string, tz string, days int) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	if tz == "" {
		tz = "CST"
	}
	plStr := fmt.Sprintf(`{
	"ignore_cache": false,
	"collection_preview": false,
	"parameters": [
		{
		"id": "3332d084-956d-489e-87f2-e634c4d0e42d",
		"type": "category",
		"value": "%s",
		"target": [
			"variable",
			[
			"template-tag",
			"trader"
			]
		]
		},
		{
		"id": "d988ac7f-ecb0-40d7-8123-5410e5598ebc",
		"type": "category",
		"value": "%s",
		"target": [
			"variable",
			[
			"template-tag",
			"tz"
			]
		]
		},
		{
		"id": "a9218770-8bb2-474d-ab33-11ffdc7e2052",
		"type": "number/=",
		"value": [
			"%d"
		],
		"target": [
			"variable",
			[
			"template-tag",
			"days"
			]
		]
		}
	]
	}`, trader, tz, days)
	payload := strings.NewReader(plStr)
	resp, err := m.apiClient.PostV2("/api/card/140/query", payload, headers, nil)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to trader overview")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal trader overview response")
		return nil, err
	}
	return &ret, nil
}

// Trader出手时间分布, 74tYkMYmwnmi44PQo6L6QpkxmdNTdX5AZaiKMrMAncwW, EPCwxro6Nf3PEtMxR638qwyyccH7F7dbH2RuXY6HVN8s
// 过去duration天, trader的交易时间分布, UTC或者CST时间
func (m *MetabaseDataSource) TraderTxTimeDistribution(trader string, duration int, timezone string) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	var id string
	var id2 string
	if timezone == "UTC" {
		id = "b8f94534-7734-4e40-b8e4-212e0d876cda"
		id2 = "e164924c-5361-411d-82de-34de55cd3e67"
	} else if timezone == "CST" {
		id = "b8f94534-7734-4e40-b8e4-212e0d876cda"
		id2 = "e164924c-5361-411d-82de-34de55cd3e67"
	}
	plStr := fmt.Sprintf(`
	{
		"ignore_cache": false,
		"collection_preview": false,
		"parameters": [
			{
			"id": "%s",
			"type": "category",
			"value": "%s",
			"target": [
				"variable",
				[
				"template-tag",
				"trader"
				]
			]
			},
			{
			"id": "%s",
			"type": "number/=",
			"value": [
				"%d"
			],
			"target": [
				"variable",
				[
				"template-tag",
				"days"
				]
			]
			}
		]
	}`, id, trader, id2, duration)
	payload := strings.NewReader(plStr)
	var resp []byte
	if timezone == "UTC" || timezone == "" {
		resp, err = m.apiClient.PostV2("/api/card/112/query", payload, headers, nil)
	} else if timezone == "CST" {
		resp, err = m.apiClient.PostV2("/api/card/120/query", payload, headers, nil)
	}
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to launched token time distribution")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal launched token time distribution response")
		return nil, err
	}
	return &ret, nil
}

// Trader利润分布, 74tYkMYmwnmi44PQo6L6QpkxmdNTdX5AZaiKMrMAncwW, EPCwxro6Nf3PEtMxR638qwyyccH7F7dbH2RuXY6HVN8s
// 过去duration天, Trader利润分布, UTC或者CST时间
func (m *MetabaseDataSource) TraderProfitTokenDistribution(trader string, duration int, timezone string) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	var id string
	var id2 string
	if timezone == "UTC" {
		id = "586e6393-9ff0-4eac-9918-d10984d441c7"
		id2 = "43a557d5-fbc8-4cd7-b8e4-101564b47695"
	} else if timezone == "CST" {
		id = "596faff0-6ef4-42cf-be21-b854f2db58af"
		id2 = "208313ef-a20b-427c-a4ea-9d154e053aed"
	}
	plStr := fmt.Sprintf(`
	{
		"ignore_cache": false,
		"collection_preview": false,
		"parameters": [
			{
			"id": "%s",
			"type": "category",
			"value": "%s",
			"target": [
				"variable",
				[
				"template-tag",
				"trader"
				]
			]
			},
			{
			"id": "%s",
			"type": "number/=",
			"value": [
				"%d"
			],
			"target": [
				"variable",
				[
				"template-tag",
				"days"
				]
			]
			}
		]
	}`, id, trader, id2, duration)
	payload := strings.NewReader(plStr)
	var resp []byte
	if timezone == "UTC" || timezone == "" {
		resp, err = m.apiClient.PostV2("/api/card/113/query", payload, headers, nil)
	} else if timezone == "CST" {
		resp, err = m.apiClient.PostV2("/api/card/119/query", payload, headers, nil)
	}
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to get trader profit token distribution")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal trader profit token distribution response")
		return nil, err
	}
	return &ret, nil
}

// Trader近7日收益分布, 74tYkMYmwnmi44PQo6L6QpkxmdNTdX5AZaiKMrMAncwW, EPCwxro6Nf3PEtMxR638qwyyccH7F7dbH2RuXY6HVN8s, 8i57XsS3E4iuw2qy2cPbKDWnW4pwx6yaBc7N7UQzG3MJ
// 过去duration天, Trader近7日收益分布, UTC或者CST时间
func (m *MetabaseDataSource) TraderProfitDistribution(trader string, duration int, timezone string) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	if timezone == "" {
		timezone = "CST"
	}
	var id string
	var id2 string
	if timezone == "UTC" {
		id = "90848092-e63d-4b0a-8e07-265fd413e2f2"
		id2 = "a8098d1f-6d88-4e64-8b3f-4ee5a0e92994"
	} else if timezone == "CST" {
		id = "f4bba41b-2de9-4fa1-9651-96d696aa221e"
		id2 = "158050e1-75eb-4729-8567-1214ba6ff071"
	}
	plStr := fmt.Sprintf(`
	{
		"ignore_cache": false,
		"collection_preview": false,
		"parameters": [
			{
				"id": "%s",
				"type": "category",
				"value": "%s",
				"target": [
					"variable",
					[
						"template-tag",
						"trader"
					]
				]
			},
			{
				"id": "%s",
				"type": "number/=",
				"value": [
					"%d"
				],
				"target": [
					"variable",
					[
						"template-tag",
						"days"
					]
				]
			}
		]
	}`, id, trader, id2, duration)
	payload := strings.NewReader(plStr)
	var resp []byte
	if timezone == "UTC" || timezone == "" {
		resp, err = m.apiClient.PostV2("/api/card/111/query", payload, headers, nil)
	} else if timezone == "CST" {
		resp, err = m.apiClient.PostV2("/api/card/118/query", payload, headers, nil)
	}
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to get trader profit distribution")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal trader profit distribution response")
		return nil, err
	}
	return &ret, nil
}

// Top Traders
func (m *MetabaseDataSource) TopTrader(duration int, winRatio float32) (*model.DatasetQueryResults, error) {
	token, err := m.Auth(
		m.baseComponent.Config.Backends.MetabaseUserName,
		m.baseComponent.Config.Backends.MetabasePassword,
	)
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to auth metabase")
		return nil, err
	}
	headers := map[string]string{
		"Content-Type":       "application/json",
		"X-Metabase-Session": token,
	}
	var id string
	if duration == 30 {
		id = "04f74b13-e4df-4836-b94b-72c1170dffcd"
	} else if duration == 7 {
		id = "6f67544d-f340-4dd4-91ce-56a225d95a25"
	} else { // 1 天
		id = "ec83280c-ba44-4968-9899-4a37d4f8a318"
	}
	plStr := fmt.Sprintf(`
	{
		"ignore_cache": false,
		"collection_preview": false,
		"parameters": [
			{
			"id": "%s",
			"type": "number/=",
			"value": [
				"%v"
			],
			"target": [
				"variable",
				[
					"template-tag",
					"win_ratio"
				]
			]
			}
		]
	}`, id, winRatio)
	payload := strings.NewReader(plStr)
	var resp []byte
	if duration == 30 {
		resp, err = m.apiClient.PostV2("/api/card/125/query", payload, headers, nil)
	} else if duration == 7 {
		resp, err = m.apiClient.PostV2("/api/card/124/query", payload, headers, nil)
	} else { // 1 天
		resp, err = m.apiClient.PostV2("/api/card/131/query", payload, headers, nil)
	}
	if err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to get trader profit distribution")
		return nil, err
	}
	var ret model.DatasetQueryResults
	if err := json.Unmarshal(resp, &ret); err != nil {
		m.baseComponent.Logger.WithField("err", err).Error("failed to unmarshal trader profit distribution response")
		return nil, err
	}
	return &ret, nil
}
