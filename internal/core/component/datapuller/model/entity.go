package model

import "time"

type DatasetQueryResultsMetadataColumn struct {
	BaseType      string `json:"base_type,omitempty"`
	DisplayName   string `json:"display_name,omitempty"`
	Name          string `json:"name,omitempty"`
	EffectiveType string `json:"effective_type,omitempty"`
}

type DatasetQueryResultsCol struct {
	Description   string `json:"description,omitempty"`
	TableId       int64  `json:"table_id,omitempty"`
	SchemaName    string `json:"schema_name,omitempty"`
	EffectiveType string `json:"effective_type,omitempty"`
	Name          string `json:"name,omitempty"`
	Source        string `json:"source,omitempty"`
	// unknown type
	RemappedFrom string `json:"remapped_from,omitempty"`
	// can be '{\"target_table_id\":517}'
	ExtraInfo map[string]interface{} `json:"extra_info,omitempty"`
	// unknown type
	FkFieldId string `json:"fk_field_id,omitempty"`
	// unknown type
	RemappedTo     string                            `json:"remapped_to,omitempty"`
	Id             int64                             `json:"id,omitempty"`
	VisibilityType string                            `json:"visibility_type,omitempty"`
	Target         DatasetQueryResultsColTarget      `json:"target,omitempty"`
	DisplayName    string                            `json:"display_name,omitempty"`
	Fingerprint    DatasetQueryResultsColFingerprint `json:"fingerprint,omitempty"`
	BaseType       string                            `json:"base_type,omitempty"`
}

type DatasetQueryResultsColFingerprint struct {
	Global DatasetQueryResultsColFingerprintGlobal `json:"global,omitempty"`
	// map[string]DatasetQueryResultsColFingerprintType results in map[string]interface{}
	Type map[string]interface{} `json:"type,omitempty"`
}

type DatasetQueryResultsColFingerprintGlobal struct {
	DistinctCount int64 `json:"distinct-count,omitempty"`
}

type DatasetQueryResultsColTarget struct {
	Id             int64  `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
	TableId        int64  `json:"table_id,omitempty"`
	Description    string `json:"description,omitempty"`
	BaseType       string `json:"base_type,omitempty"`
	EffectiveType  string `json:"effective_type,omitempty"`
	VisibilityType string `json:"visibility_type,omitempty"`
}

type DatasetQueryResultsNativeForm struct {
	Query string `json:"query,omitempty"`
	// unknown type
	// Params string `json:"params,omitempty"`
}

// 响应的data字段
type DatasetQueryResultsData struct {
	Rows            [][]interface{}               `json:"rows,omitempty"`
	NativeForm      DatasetQueryResultsNativeForm `json:"native_form,omitempty"`
	Cols            []DatasetQueryResultsCol      `json:"cols,omitempty"`
	ResultsMetadata DatasetQueryResultsMetadata   `json:"results_metadata,omitempty"`
	RowsTruncated   int64                         `json:"rows_truncated,omitempty"`
}

type DatasetQueryResultsMetadata struct {
	Checksum string                              `json:"checksum,omitempty"`
	Columns  []DatasetQueryResultsMetadataColumn `json:"columns,omitempty"`
}

type DatasetQueryResults struct {
	StartedAt time.Time             `json:"started_at,omitempty"`
	JsonQuery DatasetQueryJsonQuery `json:"json_query,omitempty"`
	// type unknown
	AverageExecutionTime float64                 `json:"average_execution_time,omitempty"`
	Status               string                  `json:"status,omitempty"`
	Context              string                  `json:"context,omitempty"`
	RowCount             int64                   `json:"row_count,omitempty"`
	RunningTime          int64                   `json:"running_time,omitempty"`
	Data                 DatasetQueryResultsData `json:"data,omitempty"`
}

type DatasetQueryJsonQuery struct {
	Database    int64                   `json:"database,omitempty"`
	Type        string                  `json:"type,omitempty"`
	Native      DatasetQueryNative      `json:"native,omitempty"`
	Query       DatasetQueryDsl         `json:"query,omitempty"`
	Constraints DatasetQueryConstraints `json:"constraints,omitempty"`
}

type DatasetQueryNative struct {
	Query string `json:"query,omitempty"`
}

type DatasetQueryDsl struct {
	SourceTable int64               `json:"source_table,omitempty"`
	Limit       int64               `json:"limit,omitempty"`
	Page        DatasetQueryDslPage `json:"page,omitempty"`
}

type DatasetQueryConstraints struct {
	MaxResults         int64 `json:"max-results,omitempty"`
	MaxResultsBareRows int64 `json:"max-results-bare-rows,omitempty"`
}

type DatasetQueryDslPage struct {
	Page  int64 `json:"page,omitempty"`
	Items int64 `json:"items,omitempty"`
}

type CommonPumpDataQuery struct {
	Duration   int     `json:"duration" form:"duration"`
	Timezone   string  `json:"timezone" form:"timezone"` // CST,UTC
	MaxWinRate float64 `json:"max_win_rate" form:"max_win_rate"`
	Address    string  `json:"address" form:"address"`
	Source     string  `json:"source" form:"source"` // design for test: if Source == "testsource", then return mock data
}

type NewTokensVO struct {
	Rows []*DailyTokensData `json:"rows"`
}

type DailyTokensData struct {
	Date       string  `json:"date"`
	TotalCount int64   `json:"total_count"`
	P2RCount   int64   `json:"p2r_count"`
	P2RRatio   float64 `json:"p2r_ratio"`
}

type LaunchTimeVO struct {
	Rows []*LaunchTimeData `json:"rows"`
}

type LaunchTimeData struct {
	TimeRange     string `json:"time_range"`
	LaunchedCount int64  `json:"launched_count"`
}

type TransactionsVO struct {
	Rows []*TradeCountData `json:"rows"`
}

type TradeCountData struct {
	Date       string `json:"date"`
	TradeCount int64  `json:"trade_count"`
}

type TopTraderData struct {
	Trader              string  `json:"trader"`
	TotalNetProfit      float64 `json:"total_net_profit"`
	NetProfitWinRatio   float64 `json:"net_profit_win_ratio"`
	GrossProfitWinRatio float64 `json:"gross_profit_win_ratio"`
	TotalTxCount        int64   `json:"total_tx_count"`
}

type TopTradersVO struct {
	Rows []*TopTraderData `json:"rows"`
}

type TraderInfo struct {
	Address string   `json:"address"`
	Tag     []string `json:"tag"`
}

type TraderInfoVO struct {
	Info *TraderInfo `json:"info"`
}

type TraderOverviewInfo struct {
	TotalNetProfit             float64 `json:"total_net_profit"`
	NetProfitWinRatio          float64 `json:"net_profit_win_ratio"`
	GrossProfitWinRatio        float64 `json:"gross_profit_win_ratio"`
	TradedTokenCount           int64   `json:"traded_token_count"`
	TotalTxCount               int64   `json:"total_tx_count"`
	SuccessTxCount             int64   `json:"success_tx_count"`
	RevertedTxCount            int64   `json:"reverted_tx_count"`
	TradedTokenCountPercentage float64 `json:"traded_token_count_percentage"`
	SnipedTokenCount           int64   `json:"sniped_token_count"`
	SnipedTokenCountPercentage float64 `json:"sniped_token_count_percentage"`
	TotalGrossProfit           float64 `json:"total_gross_profit"`
	AvgSolCostPerToken         float64 `json:"avg_sol_cost_per_token"`
	TotalGasFee                float64 `json:"total_gas_fee"`
	TotalTip                   float64 `json:"total_tip"`
	TotalCommission            float64 `json:"total_commission"`
	AvgFeePerToken             float64 `json:"avg_fee_per_token"`
	AvgTipPerToken             float64 `json:"avg_tip_per_token"`
	AvgBuyCountPerToken        float64 `json:"avg_buy_count_per_token"`
	AvgSellCountPerToken       float64 `json:"avg_sell_count_per_token"`
}

type TraderOverviewInfoV2 struct {
	TotalNetProfit     float64 `json:"total_net_profit"`       // 总利润
	ProfitRatio        float64 `json:"profit_ratio"`           // 利润率
	NetProfitWinRatio  float64 `json:"net_profit_win_ratio"`   // 胜率
	TradedTokenCount   int64   `json:"traded_token_count"`     // 交易了多少个token
	AvgSolCostPerToken float64 `json:"avg_sol_cost_per_token"` // 每个token的平均成本
	TotalCost          float64 `json:"total_cost"`             // 前面两个相乘
	AvgTipPerToken     float64 `json:"avg_tip_per_token"`      // 每个token的平均tip
	AvgFeePerToken     float64 `json:"avg_fee_per_token"`      // 每个token的平均Gas费
	TokenCreateCount   int64   `json:"token_create_count"`     // 创建了多少个token
}

type TraderOverviewVO struct {
	Info *TraderOverviewInfoV2 `json:"info"`
}

type TraderProfitData struct {
	NetProfit   float64 `json:"net_profit"`
	GrossProfit float64 `json:"gross_profit"`
	Date        string  `json:"date"`
}

type TraderProfitVO struct {
	Rows []*TraderProfitData `json:"rows"`
}

type ProfitDistributionData struct {
	ProfitMarginBucket string `json:"profit_margin_bucket"`
	TokenCount         int64  `json:"token_count"`
}

type ProfitDistributionVO struct {
	Rows []*ProfitDistributionData `json:"rows"`
}

type TraderTradesData struct {
	TimeRange string `json:"time_range"`
	TxCount   int64  `json:"tx_count"`
}

type TraderTradesVO struct {
	Rows []*TraderTradesData `json:"rows"`
}

type TraderDetailVO struct {
	Trader             *Trader                   `json:"trader"`
	Info               *TraderOverviewInfoV2     `json:"overview"`
	Profit             []*TraderProfitData       `json:"profit"`
	ProfitDistribution []*ProfitDistributionData `json:"profit_distribution"`
	Trades             []*TraderTradesData       `json:"trades"`
}

type Trader struct {
	Address string `json:"address"`
}

type GetSupportedChainsReq struct {
	ChainId int `json:"chainId" form:"chainId"`
}

type GetSupportedTokensReq struct {
	ChainId int `json:"chainId" form:"chainId"`
}

type GetTokenReq struct {
	ChainId      int    `json:"chainId" form:"chainId"`
	TokenAddress string `json:"tokenAddress" form:"tokenAddress"`
}

type GetQuoteReq struct {
	ChainId          int    `json:"chainId" form:"chainId"`
	FromTokenAddress string `json:"fromTokenAddress" form:"fromTokenAddress"`
	ToTokenAddress   string `json:"toTokenAddress" form:"toTokenAddress"`
	Amount           string `json:"amount" form:"amount"`
}

type ApproveTransactionReq struct {
	ChainId              int    `json:"chainId" form:"chainId"`
	TokenContractAddress string `json:"tokenContractAddress" form:"tokenContractAddress"`
	ApproveAmount        string `json:"approveAmount" form:"approveAmount"`
}

type SwapReq struct {
	ChainId             int    `json:"chainId" form:"chainId"`
	Amount              string `json:"amount" form:"amount"`
	FromTokenAddress    string `json:"fromTokenAddress" form:"fromTokenAddress"`
	ToTokenAddress      string `json:"toTokenAddress" form:"toTokenAddress"`
	UserWalletAddress   string `json:"userWalletAddress" form:"userWalletAddress"`
	SwapReceiverAddress string `json:"swapReceiverAddress" form:"swapReceiverAddress"`
	Slippage            string `json:"slippage" form:"slippage"`
}

type GetBridgeTokensPairsReq struct {
	FromChainId int `json:"fromChainId" form:"fromChainId"`
}

type GetCrossChainQuoteReq struct {
	FromChainId      int    `json:"fromChainId" form:"fromChainId"`
	ToChainId        int    `json:"toChainId" form:"toChainId"`
	FromTokenAddress string `json:"fromTokenAddress" form:"fromTokenAddress"`
	ToTokenAddress   string `json:"toTokenAddress" form:"toTokenAddress"`
	Amount           string `json:"amount" form:"amount"`
	Slippage         string `json:"slippage" form:"slippage"`
}
