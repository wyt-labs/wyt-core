package model

import (
	"github.com/wyt-labs/wyt-core/internal/core/component/datapuller/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatMsgRole = string

const (
	ChatMsgRoleUserBuiltin ChatMsgRole = "user_builtin"
	ChatMsgRoleUser        ChatMsgRole = "user"
	ChatMsgRoleSystem      ChatMsgRole = "system"
	ChatMsgRoleAssistant   ChatMsgRole = "assistant"
)

type ChatContentAssistantType = string

const (
	ChatContentAssistantTypeFill           ChatContentAssistantType = "fill"
	ChatContentAssistantTypeProjectInfo    ChatContentAssistantType = "project_info"
	ChatContentAssistantTypeGeneral        ChatContentAssistantType = "general_answer"
	ChatContentAssistantTypeProjectCompare ChatContentAssistantType = "project_compare"
	ChatContentAssistantTypeSwap           ChatContentAssistantType = "swap"
)

type ChatContentAssistantProjectInfoView = string

const (
	ChatContentAssistantProjectInfoViewOverview      ChatContentAssistantProjectInfoView = "overview"
	ChatContentAssistantProjectInfoViewTokenomics    ChatContentAssistantProjectInfoView = "tokenomics"
	ChatContentAssistantProjectInfoViewProfitability ChatContentAssistantProjectInfoView = "profitability"
	ChatContentAssistantProjectInfoViewTeam          ChatContentAssistantProjectInfoView = "team"
	ChatContentAssistantProjectInfoViewFunding       ChatContentAssistantProjectInfoView = "funding"
	ChatContentAssistantProjectInfoViewExchanges     ChatContentAssistantProjectInfoView = "exchanges"
	ChatContentAssistantProjectInfoViewAll           ChatContentAssistantProjectInfoView = "all"
	ChatContentAssistantProjectInfoViewGeneralAnswer ChatContentAssistantProjectInfoView = "general_answer"
)

type ChatContentAssistantView = string

const (
	ChatContentAssistantSwapViewNative     ChatContentAssistantView = "nativeswap"
	ChatContentAssistantSwapViewUniswap    ChatContentAssistantView = "uniswap"
	ChatContentAssistantSwapViewSushi      ChatContentAssistantView = "sushiswap"
	ChatContentAssistantNewTokenView       ChatContentAssistantView = "daily_new_token"
	ChatContentAssistantTokenLTView        ChatContentAssistantView = "token_launched_time_distribution"
	ChatContentAssistantTokenSwapCountView ChatContentAssistantView = "daily_token_swap_count"
	ChatContentAssistantTopTrader          ChatContentAssistantView = "top_trader"
	ChatContentAssistantTraderOverview     ChatContentAssistantView = "trader_overview"
)

type FuncCallingType = string

const (
	FCSwap FuncCallingType = "swap"
	FCMeMe FuncCallingType = "meme"
	// DailyNewTokens
	FCDailyNewToken FuncCallingType = "daily_new_token"
	// TokenLaunchedTimeDistribution
	FCTokenLaunchedTimeDistribution FuncCallingType = "token_launched_time_distribution"
	// DailyTokenSwapCounts
	FCDailyTokenSwapCount FuncCallingType = "daily_token_swap_count"
	// TopTraders
	FCTopTrader FuncCallingType = "top_trader"
	// TraderOverview
	FCTraderOverview FuncCallingType = "trader_overview"
	FCUniswap        FuncCallingType = "uniswap"
)

type ChatContentUser struct {
	Content string `json:"content" bson:"content"`
}

type ChatContentUserBuiltin struct {
	Content string `json:"content" bson:"content"`
}

type ChatContentSystem struct {
	Content string `json:"content" bson:"content"`
}

type ChatContentAssistantProjectInfoTrack struct {
	ID          primitive.ObjectID `json:"id" bson:"id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
}

type ChatContentAssistantProjectInfoTag struct {
	ID          primitive.ObjectID `json:"id" bson:"id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
}

type ChatContentAssistantProjectInfoTeamImpression struct {
	ID          primitive.ObjectID `json:"id" bson:"id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
}

type ChatContentAssistantProjectOverview struct {
	Name            string                                          `json:"name" bson:"name"`
	Description     string                                          `json:"description" bson:"description"`
	LogoURL         string                                          `json:"logo_url" bson:"logo_url"`
	TokenSymbol     string                                          `json:"token_symbol" bson:"token_symbol"`
	TokenPrice      float64                                         `json:"token_price" bson:"token_price"`
	TokenMarketCap  uint64                                          `json:"token_market_cap" bson:"token_market_cap"`
	TeamImpressions []ChatContentAssistantProjectInfoTeamImpression `json:"team_impressions" bson:"team_impressions"`
	Tracks          []ChatContentAssistantProjectInfoTrack          `json:"tracks" bson:"tracks"`
	Tags            []ChatContentAssistantProjectInfoTag            `json:"tags" bson:"tags"`
	RelatedLinks    ProjectRelatedLinks                             `json:"related_links" bson:"related_links"`
}

type ChatContentAssistantProjectInfoInvestor struct {
	ID               primitive.ObjectID `json:"id" bson:"id"`
	Name             string             `json:"name" bson:"name"`
	Description      string             `json:"description" bson:"description"`
	AvatarURL        string             `json:"avatar_url" bson:"avatar_url"`
	Subject          InvestorSubject    `json:"subject" bson:"subject"`
	Type             InvestorType       `json:"type" bson:"type"`
	SocialMediaLinks []LinkInfo         `json:"social_media_links" bson:"social_media_links"`
}

type ChatContentAssistantProjectFundingDetail struct {
	Round                 string                                    `json:"round" bson:"round"`
	Date                  string                                    `json:"date" bson:"date"`
	Amount                uint64                                    `json:"amount" bson:"amount"`
	Valuation             uint64                                    `json:"valuation" bson:"valuation"`
	Investors             string                                    `json:"investors" bson:"investors"`
	LeadInvestors         string                                    `json:"lead_investors" bson:"lead_investors"`
	InvestorsRefactor     []ChatContentAssistantProjectInfoInvestor `json:"investors_refactor" bson:"investors_refactor"`
	LeadInvestorsRefactor []ChatContentAssistantProjectInfoInvestor `json:"lead_investors_refactor" bson:"lead_investors_refactor"`
}

type ChatContentAssistantProjectInfoFunding struct {
	TopInvestors   []ChatContentAssistantProjectInfoInvestor   `json:"top_investors" bson:"top_investors"`
	FundingDetails []*ChatContentAssistantProjectFundingDetail `json:"funding_details" bson:"funding_details"`
	Reference      string                                      `json:"reference" bson:"reference"`

	// auto generate
	Highlights ProjectFundingHighlights `json:"highlights" bson:"highlights"`
}

type ChatContentAssistantProjectInfo struct {
	ID            primitive.ObjectID                      `json:"id" bson:"id"`
	Overview      *ChatContentAssistantProjectOverview    `json:"overview" bson:"overview"`
	Tokenomics    *ProjectTokenomics                      `json:"tokenomics" bson:"tokenomics"`
	Profitability *ProjectProfitability                   `json:"profitability" bson:"profitability"`
	Team          *ProjectTeam                            `json:"team" bson:"team"`
	Funding       *ChatContentAssistantProjectInfoFunding `json:"funding" bson:"funding"`
	Exchanges     *ProjectExchanges                       `json:"exchanges" bson:"exchanges"`
}

type ChatContentAssistantGeneralAnswer struct {
	ID      primitive.ObjectID `json:"id" bson:"id"`
	Content string             `json:"content" bson:"content"`
}

type ChatContentAssistantGeneralAnswerRes struct {
	View          ChatContentAssistantProjectInfoView `json:"view" bson:"view"`
	GeneralAnswer ChatContentAssistantGeneralAnswer   `json:"general_answer" bson:"general_answer"`
}

type ChatContentAssistantProjectInfoRes struct {
	View    ChatContentAssistantProjectInfoView `json:"view" bson:"view"`
	Project ChatContentAssistantProjectInfo     `json:"project" bson:"project"`
}

type ChatContentAssistantProjectCompareRes struct {
	View     ChatContentAssistantProjectInfoView `json:"view" bson:"view"`
	Projects []ChatContentAssistantProjectInfo   `json:"projects" bson:"projects"`
}

type ChatContentAssistant struct {
	Tips string                   `json:"tips" bson:"tips"`
	Type ChatContentAssistantType `json:"type" bson:"type"`

	// use one of the following fields according to the type
	Fill string `json:"fill" bson:"fill"`

	ProjectKeys    []string                               `json:"project_keys" bson:"project_keys"`
	ProjectInfo    *ChatContentAssistantProjectInfoRes    `json:"project_info" bson:"project_Info"`
	GeneralAnswer  *ChatContentAssistantGeneralAnswerRes  `json:"general_answer" bson:"general_answer"`
	ProjectCompare *ChatContentAssistantProjectCompareRes `json:"project_compare" bson:"projectCompare"`
	SwapInfo       *ChatContentAssistantSwapRes           `json:"swap_info" bson:"swap_info"`
	// pump.fun data
	DailyNewToken     *ChatContentAssistantNewTokenRes       `json:"daily_new_token" bson:"daily_new_token"`
	TokenLaunchedTime *ChatContentAssistantTokenLtDistRes    `json:"token_launched_time_distribution" bson:"token_launched_time_distribution"`
	TokenSwapCount    *ChatContentAssistantTokenSwapCountRes `json:"daily_token_swap_count" bson:"daily_token_swap_count"`
	TopTrader         *ChatContentAssistantTopTraderRes      `json:"top_trader" bson:"top_trader"`
	TraderOverview    *ChatContentAssistantTraderOverviewRes `json:"trader_overview" bson:"trader_overview"`
	Uniswap           *ChatContentAssistantUniswapRes        `json:"uniswap" bson:"uniswap"`
}

type ChatContentAssistantSwapRes struct {
	View ChatContentAssistantView `json:"view" bson:"view"`
	Swap ChatContentAssistantInfo `json:"swap" bson:"swap"`
}

type ChatContentAssistantInfo struct {
	ID             primitive.ObjectID `json:"id" bson:"id"`
	FuncCallingRet FuncCallingRet     `json:"func_calling_ret" bson:"func_calling_ret"`
}

type FuncCallingRet struct {
	FCType          FuncCallingType                       `json:"fc_type" bson:"fc_type"`
	FCUniswapResult UniswapFuncCallingResult              `json:"uniswap" bson:"uniswap"`
	FCSwapResult    SwapFuncCallingResult                 `json:"fc_swap_result" bson:"fc_swap_result"`
	NewTokenResult  DailyNewTokensFuncCallingResult       `json:"daily_new_token" bson:"daily_new_token"`
	TokenLTD        TokenLaunchedTimeDtFuncCallingResult  `json:"token_launched_time_distribution" bson:"token_launched_time_distribution"`
	TokenSwapCount  DailyTokenSwapCountsFuncCallingResult `json:"daily_token_swap_count" bson:"daily_token_swap_count"`
	TopTrader       TopTradersFuncCallingResult           `json:"top_trader" bson:"top_trader"`
	TraderOverview  TraderOverviewFuncCallingResult       `json:"trader_overview" bson:"trader_overview"`

	// RemoteFunctionResult store the result executed by remote function
	RemoteFunctionResult map[string]any `json:"remote_function_result" bson:"remote_function_result"`
}

type ChatContentAssistantTraderOverviewRes struct {
	View           ChatContentAssistantView `json:"view" bson:"view"`
	TraderOverview ChatContentAssistantInfo `json:"trader_overview" bson:"trader_overview"`
}

type ChatContentAssistantTopTraderRes struct {
	View      ChatContentAssistantView `json:"view" bson:"view"`
	TopTrader ChatContentAssistantInfo `json:"top_trader" bson:"top_trader"`
}

type ChatContentAssistantTokenSwapCountRes struct {
	View           ChatContentAssistantView `json:"view" bson:"view"`
	TokenSwapCount ChatContentAssistantInfo `json:"daily_token_swap_count" bson:"daily_token_swap_count"`
}

type ChatContentAssistantTokenLtDistRes struct {
	View                     ChatContentAssistantView `json:"view" bson:"view"`
	LaunchedTimeDistribution ChatContentAssistantInfo `json:"token_launched_time_distribution" bson:"token_launched_time_distribution"`
}

type ChatContentAssistantUniswapRes struct {
	View    ChatContentAssistantView `json:"view" bson:"view"`
	Uniswap ChatContentAssistantInfo `json:"uniswap" bson:"uniswap"`
}

type ChatContentAssistantNewTokenRes struct {
	View     ChatContentAssistantView `json:"view" bson:"view"`
	NewToken ChatContentAssistantInfo `json:"daily_new_token" bson:"daily_new_token"`
}

type DailyNewTokensFuncCallingResult struct {
	DailyNewToken *model.NewTokensVO `json:"daily_new_token" bson:"daily_new_token"`
}

type TokenLaunchedTimeDtFuncCallingResult struct {
	LaunchTimeDt *model.LaunchTimeVO `json:"token_launched_time_distribution" bson:"token_launched_time_distribution"`
}

type DailyTokenSwapCountsFuncCallingResult struct {
	TxCounts *model.TransactionsVO `json:"tx_counts" bson:"tx_counts"`
}

type TopTradersFuncCallingResult struct {
	TopTraders *model.TopTradersVO `json:"top_traders" bson:"top_traders"`
}

type TraderOverviewFuncCallingResult struct {
	TraderDetails *model.TraderDetailVO `json:"trader_details" bson:"trader_details"`
}

type UniswapFuncCallingResult struct {
	Url string `json:"url" bson:"url"`
}

type SwapFuncCallingResult struct {
	SourceChain  string  `json:"source_chain" bson:"source_chain"`
	SwapInToken  string  `json:"swap_in_token" bson:"swap_in_token"`
	AmountIn     float64 `json:"amount_in" bson:"amount_in"`
	SwapOutToken string  `json:"swap_out_token" bson:"swap_out_token"`
	SwapOut      float64 `json:"swap_out" bson:"swap_out"`
	DestChain    string  `json:"dest_chain" bson:"dest_chain"`
	DEX          string  `json:"dex" bson:"dex"`
}

type ChatMsg struct {
	Timestamp JSONTime    `json:"timestamp" bson:"timestamp"`
	Role      ChatMsgRole `json:"role" bson:"role"`

	// use one of the following fields according to the role
	ContentUser *ChatContentUser `json:"content_user" bson:"content_user"`

	ContentUserBuiltin *ChatContentUserBuiltin `json:"content_user_builtin" bson:"content_user_builtin"`
	ContentSystem      *ChatContentSystem      `json:"content_system" bson:"content_system"`
	ContentAssistant   *ChatContentAssistant   `json:"content_assistant" bson:"content_assistant"`
}

type ChatAIAnalyticalIntention = string

const (
	ChatAIAnalyticalIntentionUniswap ChatAIAnalyticalIntention = "uniswap"
	ChatAIAnalyticalIntentionDex     ChatAIAnalyticalIntention = "dex"
	ChatAIAnalyticalIntentionSearch  ChatAIAnalyticalIntention = "search"
	ChatAIAnalyticalIntentionCompare ChatAIAnalyticalIntention = "compare"
	ChatAIAnalyticalIntentionSwap    ChatAIAnalyticalIntention = "swap"
	// DailyNewTokens
	ChatAIDailyNewToken ChatAIAnalyticalIntention = "daily_new_token"
	// TokenLaunchedTimeDistribution
	ChatAITokenLaunchedTimeDistribution ChatAIAnalyticalIntention = "token_launched_time_distribution"
	// DailyTokenSwapCounts
	ChatAIDailyTokenSwapCount ChatAIAnalyticalIntention = "daily_token_swap_count"
	// TopTraders
	ChatAITopTrader ChatAIAnalyticalIntention = "top_trader"
	// TraderOverview
	ChatAITraderOverview             ChatAIAnalyticalIntention = "trader_overview"
	ChatAIAnalyticalIntentionGeneral ChatAIAnalyticalIntention = "general"
)

type ChatAIAnalyticalResult struct {
	Intention  ChatAIAnalyticalIntention `json:"intention"`
	IntentKeys []string                  `json:"intent_keys"`
	View       string                    `json:"view"`
	Fill       string                    `json:"fill"`
	Content    string                    `json:"content"`
	ProjectIDs []primitive.ObjectID      `json:"project_ids"`
}

type ChatWindow struct {
	BaseModel `bson:"inline"`

	Title      string `json:"title" bson:"title,omitempty"`
	TitleIsSet bool   `json:"-" bson:"title_is_set"`

	// keep the last five for interactive conversations with ai
	LastUserMsgs []ChatMsg `json:"-" bson:"last_user_msgs"`

	MsgNum uint64 `json:"msg_num" bson:"msg_num"`

	ProjectId string `json:"project_id" bson:"project_id"`
}

type ChatHistory struct {
	BaseModel `bson:"inline"`
	WindowID  primitive.ObjectID `json:"window_id" bson:"window_id"`
	Index     uint64             `json:"index" bson:"index"`
	Msg       ChatMsg            `json:"msg" bson:"msg"`
	ProjectId string             `json:"project_id" bson:"project_id"`
}
