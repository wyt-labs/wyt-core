package entity

type SwapParams struct {
	SourceChain  string  `json:"source_chain"`
	SwapInToken  string  `json:"swap_in_token"`
	AmountIn     float64 `json:"amount_in"`
	DestChain    string  `json:"dest_chain"`
	SwapOutToken string  `json:"swap_out_token"`
	SwapOut      float64 `json:"swap_out"`
	DEX          string  `json:"dex"`
}

type DailyNewTokensParams struct {
	LaunchpadName string `json:"launchpad_name"`
	Duration      int    `json:"duration"`
	Timezone      string `json:"timezone"`
}

type TokenLaunchedTimeDtParams struct {
	LaunchpadName string `json:"launchpad_name"`
	Duration      int    `json:"duration"`
	Timezone      string `json:"timezone"`
}

type DailyTokenSwapCountsParams struct {
	LaunchpadName string `json:"launchpad_name"`
	Duration      int    `json:"duration"`
	Timezone      string `json:"timezone"`
}

type TopTraderParams struct {
	WinRatio float32 `json:"win_ratio"`
	Duration int     `json:"duration"`
	Timezone string  `json:"timezone"`
}

type TraderOverviewParams struct {
	TraderAddr string `json:"address"`
}
