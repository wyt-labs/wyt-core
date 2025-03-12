package datapuller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wyt-labs/wyt-core/internal/core/component/datapuller/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
)

type PumpDataService struct {
	BaseComponent      *base.Component
	metabaseDataSource *MetabaseDataSource
}

func NewPumpDataService(baseComponent *base.Component, metabaseDataSource *MetabaseDataSource) *PumpDataService {
	return &PumpDataService{
		BaseComponent:      baseComponent,
		metabaseDataSource: metabaseDataSource,
	}
}

func (pd *PumpDataService) NewTokens(ctx context.Context, req *model.CommonPumpDataQuery) (*model.NewTokensVO, error) {
	if "testdata" == req.Source {
		return &model.NewTokensVO{
			Rows: []*model.DailyTokensData{
				{
					Date:       "2024-09-21",
					TotalCount: 100,
					P2RCount:   50,
					P2RRatio:   0.5,
				},
				{
					Date:       "2024-09-22",
					TotalCount: 200,
					P2RCount:   100,
					P2RRatio:   0.5,
				},
			},
		}, nil
	}
	metaRes, err := pd.metabaseDataSource.DailyLaunchedTokenInfo(req.Duration, req.Timezone)
	if err != nil {
		return nil, err
	}

	res := &model.NewTokensVO{
		Rows: make([]*model.DailyTokensData, 0),
	}
	for _, row := range metaRes.Data.Rows {
		dateStr, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid date format: %s", row[0])
		}
		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %v", err)
		}
		totalCount, ok := row[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid total count format: %s", row[1])
		}
		p2rCount, ok := row[2].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid p2r count format: %s", row[2])
		}
		p2rRatio, ok := row[3].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid p2r ratio format: %s", row[3])
		}
		dailyTokensData := &model.DailyTokensData{
			Date:       date.Format(time.RFC3339),
			TotalCount: int64(totalCount),
			P2RCount:   int64(p2rCount),
			P2RRatio:   p2rRatio,
		}
		res.Rows = append(res.Rows, dailyTokensData)
	}

	return res, nil
}

func (pd *PumpDataService) LaunchTime(ctx context.Context, req *model.CommonPumpDataQuery) (*model.LaunchTimeVO, error) {
	if "testdata" == req.Source {
		return &model.LaunchTimeVO{
			Rows: []*model.LaunchTimeData{
				{
					TimeRange:     "[00:00~00:30)",
					LaunchedCount: 766,
				},
				{
					TimeRange:     "[00:30~01:00)",
					LaunchedCount: 686,
				},
			},
		}, nil
	}

	metaRes, err := pd.metabaseDataSource.LaunchedTokenTimeDistribution(req.Duration, req.Timezone)
	if err != nil {
		return nil, err
	}

	res := &model.LaunchTimeVO{
		Rows: make([]*model.LaunchTimeData, 0),
	}

	for _, row := range metaRes.Data.Rows {
		timeRange, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid time range format: %s", row[0])
		}
		launchedCount, ok := row[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid launched count format: %s", row[1])
		}
		launchTimeData := &model.LaunchTimeData{
			TimeRange:     timeRange,
			LaunchedCount: int64(launchedCount),
		}
		res.Rows = append(res.Rows, launchTimeData)
	}

	return res, nil
}

func (pd *PumpDataService) Transactions(ctx context.Context, req *model.CommonPumpDataQuery) (*model.TransactionsVO, error) {
	duration := req.Duration
	timezone := req.Timezone
	source := req.Source

	if source == "testdata" {
		return &model.TransactionsVO{
			Rows: []*model.TradeCountData{
				{
					Date:       "2024-09-13T00:00:00+08:00",
					TradeCount: 1462284,
				},
				{
					Date:       "2024-09-14T00:00:00+08:00",
					TradeCount: 1353946,
				},
			},
		}, nil
	}

	results, err := pd.metabaseDataSource.DailyTradeCounts(duration, timezone)
	if err != nil {
		return nil, err
	}

	rows := make([]*model.TradeCountData, len(results.Data.Rows))
	for i, row := range results.Data.Rows {
		dateStr, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid date format: %s", row[0])
		}
		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %v", err)
		}
		tradeCount, ok := row[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid trade count format")
		}

		rows[i] = &model.TradeCountData{
			Date:       date.Format(time.RFC3339),
			TradeCount: int64(tradeCount),
		}
	}

	return &model.TransactionsVO{Rows: rows}, nil
}

func (pd *PumpDataService) TopTraders(ctx context.Context, req *model.CommonPumpDataQuery) (*model.TopTradersVO, error) {
	duration := req.Duration
	maxWinRate := req.MaxWinRate
	source := req.Source

	if duration == 0 {
		duration = 7
	}

	if maxWinRate == 0 {
		maxWinRate = 1.0
	}

	if source == "testdata" {
		return getTopTradersTestData(), nil
	}

	results, err := pd.metabaseDataSource.TopTrader(duration, float32(maxWinRate))
	if err != nil {
		return nil, err
	}

	rows := make([]*model.TopTraderData, len(results.Data.Rows))
	for i, row := range results.Data.Rows {
		trader, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid trader format: %v", row[0])
		}

		totalNetProfit, ok := row[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid total net profit format: %v", row[1])
		}

		netProfitWinRatio, ok := row[2].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid net profit win ratio format: %v", row[2])
		}

		grossProfitWinRatio, ok := row[3].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid gross profit win ratio format: %v", row[3])
		}

		totalTxCount, ok := row[4].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid total transaction count format: %v", row[4])
		}

		rows[i] = &model.TopTraderData{
			Trader:              trader,
			TotalNetProfit:      totalNetProfit,
			NetProfitWinRatio:   netProfitWinRatio,
			GrossProfitWinRatio: grossProfitWinRatio,
			TotalTxCount:        int64(totalTxCount),
		}
	}

	// 只返回前10条数据
	if len(rows) > 10 {
		rows = rows[:10]
	}

	return &model.TopTradersVO{Rows: rows}, nil
}

func getTopTradersTestData() *model.TopTradersVO {
	return &model.TopTradersVO{
		Rows: []*model.TopTraderData{
			{
				Trader:              "74tYkMYmwnmi44PQo6L6QpkxmdNTdX5AZaiKMrMAncwW",
				TotalNetProfit:      0.12,
				NetProfitWinRatio:   0.12,
				GrossProfitWinRatio: 0.12,
				TotalTxCount:        1200,
			},
		},
	}
}

func (pd *PumpDataService) TraderInfo(ctx context.Context, req *model.CommonPumpDataQuery) (*model.TraderInfoVO, error) {
	//address := req.Address
	source := req.Source

	if source == "testdata" {
		return getTraderInfoTestData(), nil
	}

	overview, err := pd.TraderOverviewV2(ctx, req)
	if err != nil {
		return nil, err
	}
	if overview.Info == nil {
		return nil, fmt.Errorf("trader info not found")
	}

	traderInfo := &model.TraderInfo{
		Address: req.Address,
	}
	return &model.TraderInfoVO{Info: traderInfo}, nil
}

func getTraderInfoTestData() *model.TraderInfoVO {
	return &model.TraderInfoVO{
		Info: &model.TraderInfo{
			Address: "74tYkMYmwnmi44PQo6L6QpkxmdNTdX5AZaiKMrMAncwW",
			Tag:     []string{"sniper", "MEV", "creator"},
		},
	}
}

func (pd *PumpDataService) TraderOverview(ctx context.Context, req *model.CommonPumpDataQuery) (*model.TraderOverviewVO, error) {
	address := req.Address
	source := req.Source

	if source == "testdata" {
		return getTraderOverviewTestData(), nil
	}

	metaRes, err := pd.metabaseDataSource.TraderOverview(address)
	if err != nil {
		return nil, err
	}

	// Map results to TraderOverviewInfo
	var res *model.TraderOverviewInfoV2
	if len(metaRes.Data.Rows) > 0 {
		metaInfo := metaRes.Data.Rows[0]

		totalNetProfit, ok := metaInfo[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid total net profit format: %v", metaInfo[0])
		}

		netProfitWinRatio, ok := metaInfo[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid net profit win ratio format: %v", metaInfo[1])
		}

		// grossProfitWinRatio, ok := metaInfo[2].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid gross profit win ratio format: %v", metaInfo[2])
		// }

		tradedTokenCount, ok := metaInfo[3].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid traded token count format: %v", metaInfo[3])
		}

		// totalTxCount, ok := metaInfo[4].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid total transaction count format: %v", metaInfo[4])
		// }

		// successTxCount, ok := metaInfo[5].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid success transaction count format: %v", metaInfo[5])
		// }

		// revertedTxCount, ok := metaInfo[6].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid reverted transaction count format: %v", metaInfo[6])
		// }

		// tradedTokenCountPercentage, ok := metaInfo[7].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid traded token count percentage format: %v", metaInfo[7])
		// }

		// snipedTokenCount, ok := metaInfo[8].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid sniped token count format: %v", metaInfo[8])
		// }

		// snipedTokenCountPercentage, ok := metaInfo[9].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid sniped token count percentage format: %v", metaInfo[9])
		// }

		// totalGrossProfit, ok := metaInfo[10].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid total gross profit format: %v", metaInfo[10])
		// }

		avgSolCostPerToken, ok := metaInfo[11].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid average solution cost per token format: %v", metaInfo[11])
		}

		// totalGasFee, ok := metaInfo[12].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid total gas fee format: %v", metaInfo[12])
		// }

		// totalTip, ok := metaInfo[13].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid total tip format: %v", metaInfo[13])
		// }

		// totalCommission, ok := metaInfo[14].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid total commission format: %v", metaInfo[14])
		// }

		avgFeePerToken, ok := metaInfo[15].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid average fee per token format: %v", metaInfo[15])
		}

		avgTipPerToken, ok := metaInfo[16].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid average tip per token format: %v", metaInfo[16])
		}

		// avgBuyCountPerToken, ok := metaInfo[17].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid average buy count per token format: %v", metaInfo[17])
		// }

		// avgSellCountPerToken, ok := metaInfo[18].(float64)
		// if !ok {
		// 	return nil, fmt.Errorf("invalid average sell count per token format: %v", metaInfo[18])
		// }

		createdTokens, ok := metaInfo[19].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid create token number: %v", metaInfo[19])
		}

		totalCost := avgSolCostPerToken * tradedTokenCount
		res = &model.TraderOverviewInfoV2{
			TotalNetProfit:     totalNetProfit,
			ProfitRatio:        totalNetProfit / totalCost,
			NetProfitWinRatio:  netProfitWinRatio,
			TradedTokenCount:   int64(tradedTokenCount),
			AvgSolCostPerToken: avgSolCostPerToken,
			TotalCost:          totalCost,
			AvgTipPerToken:     avgTipPerToken,
			AvgFeePerToken:     avgFeePerToken,
			TokenCreateCount:   int64(createdTokens),
		}

		// res = &model.TraderOverviewInfo{
		// 	TotalNetProfit:             totalNetProfit,
		// 	NetProfitWinRatio:          netProfitWinRatio,
		// 	GrossProfitWinRatio:        grossProfitWinRatio,
		// 	TradedTokenCount:           int64(tradedTokenCount),
		// 	TotalTxCount:               int64(totalTxCount),
		// 	SuccessTxCount:             int64(successTxCount),
		// 	RevertedTxCount:            int64(revertedTxCount),
		// 	TradedTokenCountPercentage: tradedTokenCountPercentage,
		// 	SnipedTokenCount:           int64(snipedTokenCount),
		// 	SnipedTokenCountPercentage: snipedTokenCountPercentage,
		// 	TotalGrossProfit:           totalGrossProfit,
		// 	AvgSolCostPerToken:         avgSolCostPerToken,
		// 	TotalGasFee:                totalGasFee,
		// 	TotalTip:                   totalTip,
		// 	TotalCommission:            totalCommission,
		// 	AvgFeePerToken:             avgFeePerToken,
		// 	AvgTipPerToken:             avgTipPerToken,
		// 	AvgBuyCountPerToken:        avgBuyCountPerToken,
		// 	AvgSellCountPerToken:       avgSellCountPerToken,
		// }
	}

	return &model.TraderOverviewVO{Info: res}, nil
}

func (pd *PumpDataService) TraderOverviewV2(ctx context.Context, req *model.CommonPumpDataQuery) (*model.TraderOverviewVO, error) {
	address := req.Address
	source := req.Source

	if source == "testdata" {
		return getTraderOverviewTestData(), nil
	}
	if req.Duration == 0 {
		req.Duration = 7
	}
	if req.Timezone == "" {
		req.Timezone = "CST"
	}

	metaRes, err := pd.metabaseDataSource.TraderOverviewV2(address, req.Timezone, req.Duration)
	if err != nil {
		return nil, err
	}

	// Map results to TraderOverviewInfo
	var res *model.TraderOverviewInfoV2
	if len(metaRes.Data.Rows) > 0 {
		metaInfo := metaRes.Data.Rows[0]

		totalNetProfit, ok := metaInfo[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid total net profit format: %v", metaInfo[0])
		}

		netProfitWinRatio, ok := metaInfo[3].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid net profit win ratio format: %v", metaInfo[1])
		}

		tradedTokenCount, ok := metaInfo[5].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid traded token count format: %v", metaInfo[3])
		}

		avgSolCostPerToken, ok := metaInfo[13].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid average solution cost per token format: %v", metaInfo[11])
		}

		avgFeePerToken, ok := metaInfo[17].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid average fee per token format: %v", metaInfo[15])
		}

		avgTipPerToken, ok := metaInfo[18].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid average tip per token format: %v", metaInfo[16])
		}

		createdTokens, ok := metaInfo[21].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid create token number: %v", metaInfo[19])
		}

		totalCost := avgSolCostPerToken * tradedTokenCount
		res = &model.TraderOverviewInfoV2{
			TotalNetProfit:     totalNetProfit,
			ProfitRatio:        totalNetProfit / totalCost,
			NetProfitWinRatio:  netProfitWinRatio,
			TradedTokenCount:   int64(tradedTokenCount),
			AvgSolCostPerToken: avgSolCostPerToken,
			TotalCost:          totalCost,
			AvgTipPerToken:     avgTipPerToken,
			AvgFeePerToken:     avgFeePerToken,
			TokenCreateCount:   int64(createdTokens),
		}
	}

	return &model.TraderOverviewVO{Info: res}, nil
}

func getTraderOverviewTestData() *model.TraderOverviewVO {
	return &model.TraderOverviewVO{
		Info: &model.TraderOverviewInfoV2{
			TotalNetProfit:     6078.282028689,
			NetProfitWinRatio:  0.996003996003996,
			TradedTokenCount:   2002,
			AvgSolCostPerToken: 1.061279634,
			AvgFeePerToken:     1.5114e-5,
			AvgTipPerToken:     1.0145e-5,
		},
	}
}

func (pd *PumpDataService) TraderProfit(ctx context.Context, req *model.CommonPumpDataQuery) (*model.TraderProfitVO, error) {
	address := req.Address
	duration := req.Duration
	if duration == 0 {
		duration = 7
	}
	timezone := req.Timezone
	if timezone == "" {
		timezone = "CST"
	}
	source := req.Source

	if source == "testdata" {
		return getTraderProfitTestData(), nil
	}

	results, err := pd.metabaseDataSource.TraderProfitDistribution(address, duration, timezone)
	if err != nil {
		return nil, err
	}

	rows := make([]*model.TraderProfitData, len(results.Data.Rows))
	for i, row := range results.Data.Rows {
		date, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid date format: %v", row[2])
		}
		netProfit, ok := row[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid net profit format: %v", row[0])
		}
		grossProfit, ok := row[2].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid gross profit format: %v", row[1])
		}

		rows[i] = &model.TraderProfitData{
			NetProfit:   netProfit,
			GrossProfit: grossProfit,
			Date:        date,
		}
	}

	return &model.TraderProfitVO{Rows: rows}, nil
}

func getTraderProfitTestData() *model.TraderProfitVO {
	return &model.TraderProfitVO{
		Rows: []*model.TraderProfitData{
			{
				NetProfit:   -0.110840788,
				GrossProfit: -0.099578228,
				Date:        "2024-09-15T00:00:00+08:00",
			},
			{
				NetProfit:   13.862897836,
				GrossProfit: 15.208789556,
				Date:        "2024-09-16T00:00:00+08:00",
			},
		},
	}
}

func (pd *PumpDataService) TraderProfitDistribution(ctx context.Context, req *model.CommonPumpDataQuery) (*model.ProfitDistributionVO, error) {
	address := req.Address
	duration := req.Duration
	if duration == 0 {
		duration = 7
	}
	timezone := req.Timezone
	if timezone == "" {
		timezone = "CST"
	}
	source := req.Source

	if source == "testdata" {
		return getProfitDistributionTestData(), nil
	}

	results, err := pd.metabaseDataSource.TraderProfitTokenDistribution(address, duration, timezone)
	if err != nil {
		return nil, err
	}

	rows := make([]*model.ProfitDistributionData, len(results.Data.Rows))
	for i, row := range results.Data.Rows {
		profitMarginBucket, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid profit margin bucket format: %v", row[0])
		}
		tokenCount, ok := row[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid token count format: %v", row[1])
		}

		rows[i] = &model.ProfitDistributionData{
			ProfitMarginBucket: profitMarginBucket,
			TokenCount:         int64(tokenCount),
		}
	}

	return &model.ProfitDistributionVO{Rows: rows}, nil
}

func getProfitDistributionTestData() *model.ProfitDistributionVO {
	return &model.ProfitDistributionVO{
		Rows: []*model.ProfitDistributionData{
			{
				ProfitMarginBucket: "< -100%",
				TokenCount:         67,
			},
			{
				ProfitMarginBucket: "-100% ~ -50%",
				TokenCount:         1,
			},
		},
	}
}

func (pd *PumpDataService) TraderTrades(ctx context.Context, req *model.CommonPumpDataQuery) (*model.TraderTradesVO, error) {
	address := req.Address
	duration := req.Duration
	if duration == 0 {
		duration = 7
	}
	timezone := req.Timezone
	if timezone == "" {
		timezone = "UTC"
	}
	source := req.Source

	if source == "testdata" {
		return getTraderTradesTestData(), nil
	}

	results, err := pd.metabaseDataSource.TraderTxTimeDistribution(address, duration, timezone)
	if err != nil {
		return nil, err
	}

	rows := make([]*model.TraderTradesData, len(results.Data.Rows))
	for i, row := range results.Data.Rows {
		timeRange, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid time range format: %v", row[0])
		}
		txCount, ok := row[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid transaction count format: %v", row[1])
		}

		rows[i] = &model.TraderTradesData{
			TimeRange: timeRange,
			TxCount:   int64(txCount),
		}
	}

	return &model.TraderTradesVO{Rows: rows}, nil
}

func getTraderTradesTestData() *model.TraderTradesVO {
	return &model.TraderTradesVO{
		Rows: []*model.TraderTradesData{
			{
				TimeRange: "[00:00~00:30)",
				TxCount:   4,
			},
			{
				TimeRange: "[00:30~01:00)",
				TxCount:   0,
			},
		},
	}
}

func (pd *PumpDataService) TraderDetail(ctx context.Context, req *model.CommonPumpDataQuery) (*model.TraderDetailVO, error) {
	finalRes := &model.TraderDetailVO{}
	var wg sync.WaitGroup
	var overviewErr, profitErr, profitDistributionErr, tradesErr error

	wg.Add(4)

	go func() {
		defer wg.Done()
		overview, err := pd.TraderOverviewV2(ctx, req)
		if err != nil {
			overviewErr = err
			return
		}
		if overview != nil && overview.Info != nil {
			finalRes.Info = overview.Info
		}
	}()

	go func() {
		defer wg.Done()
		profitRes, err := pd.TraderProfit(ctx, req)
		if err != nil {
			profitErr = err
			return
		}
		if profitRes != nil && profitRes.Rows != nil {
			finalRes.Profit = profitRes.Rows
		}
	}()

	go func() {
		defer wg.Done()
		profitDistributionRes, err := pd.TraderProfitDistribution(ctx, req)
		if err != nil {
			profitDistributionErr = err
			return
		}
		if profitDistributionRes != nil && profitDistributionRes.Rows != nil {
			finalRes.ProfitDistribution = profitDistributionRes.Rows
		}
	}()

	go func() {
		defer wg.Done()
		tradesRes, err := pd.TraderTrades(ctx, req)
		if err != nil {
			tradesErr = err
			return
		}
		if tradesRes != nil && tradesRes.Rows != nil {
			finalRes.Trades = tradesRes.Rows
		}
	}()

	wg.Wait()

	if overviewErr != nil {
		return nil, overviewErr
	}
	if profitErr != nil {
		return nil, profitErr
	}
	if profitDistributionErr != nil {
		return nil, profitDistributionErr
	}
	if tradesErr != nil {
		return nil, tradesErr
	}

	finalRes.Trader = &model.Trader{
		Address: req.Address,
	}
	if finalRes.Info == nil || len(finalRes.ProfitDistribution) == 0 {
		return nil, fmt.Errorf("trader info not found")
	}
	return finalRes, nil
}
