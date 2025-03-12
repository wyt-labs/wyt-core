package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wyt-labs/wyt-core/internal/pkg/config"
)

type LinkInfo struct {
	Type string `json:"type" bson:"type"`
	Link string `json:"link" bson:"link"`
}

// fetch from other platform
type ProjectCoin struct {
	MarketCapRank uint    `json:"market_cap_rank" bson:"market_cap_rank"`
	MarketCap     uint64  `json:"market_cap" bson:"market_cap"`
	CurrentPrice  float64 `json:"current_price" bson:"current_price"`
	MarketCapAth  uint64  `json:"market_cap_ath" bson:"market_cap_ath"`
	MarketCapCL   uint64  `json:"market_cap_cl" bson:"market_cap_cl"`
	UpFromCL      uint64  `json:"up_from_cl" bson:"up_from_cl"`

	Volume              uint64 `json:"volume" bson:"volume"`
	UniqueAddressNumber uint64 `json:"unique_address_number" bson:"unique_address_number"`
}

type InfluenceType = uint32

const (
	InfluenceTypeKOLHold InfluenceType = iota
	InfluenceTypeKOLSupport
)

type Influence struct {
	Type   InfluenceType `json:"type" bson:"type"`
	Detail string        `json:"detail" bson:"detail"`
	Link   string        `json:"link" bson:"link"`
}

type ProjectBasic struct {
	Name         string               `json:"name" bson:"name"`
	LogoURL      string               `json:"logo_url" bson:"logo_url"`
	Description  string               `json:"description" bson:"description"`
	Chains       []primitive.ObjectID `json:"chains" bson:"chains"`
	Tracks       []primitive.ObjectID `json:"tracks" bson:"tracks"`
	Tags         []primitive.ObjectID `json:"tags" bson:"tags"`
	Influences   []Influence          `json:"influences" bson:"influences"`
	FoundedDate  string               `json:"founded_date" bson:"founded_date"`
	LaunchDate   string               `json:"launch_date" bson:"launch_date"`
	IsOpenSource bool                 `json:"is_open_source" bson:"is_open_source"`
	Reference    string               `json:"reference" bson:"reference"`

	// internal use
	InternalFoundedDate JSONTime `json:"-" bson:"internal_founded_date"`

	InternalLaunchDate JSONTime `json:"-" bson:"internal_launch_date"`

	// zh
	DescriptionZH string `json:"-" bson:"description_zh"`

	// 根据token symbol去间接查询第三方平台信息
	Coin ProjectCoin `json:"coin" bson:"coin"`
}

type ProjectRelatedLinks []LinkInfo

type ProjectTeamMember struct {
	Name             string     `json:"name" bson:"name"`
	AvatarURL        string     `json:"avatar_url" bson:"avatar_url"`
	Title            string     `json:"title" bson:"title"`
	IsDeparted       bool       `json:"is_departed" bson:"is_departed"`
	Description      string     `json:"description" bson:"description"`
	SocialMediaLinks []LinkInfo `json:"social_media_links" bson:"social_media_links"`

	// zh
	DescriptionZH string `json:"-" bson:"description_zh"`
}

type ProjectTeam struct {
	Impressions []primitive.ObjectID `json:"impressions" bson:"impressions"`
	Members     []ProjectTeamMember  `json:"members" bson:"members"`
	Reference   string               `json:"reference" bson:"reference"`
}

type ProjectFundingHighlights struct {
	FundingRounds        uint64   `json:"funding_rounds" bson:"funding_rounds"`
	TotalFundingAmount   uint64   `json:"total_funding_amount" bson:"total_funding_amount"`
	LeadInvestorsNumber  uint64   `json:"lead_investors_number" bson:"lead_investors_number"`
	InvestorsNumber      uint64   `json:"investors_number" bson:"investors_number"`
	LargeFundingIndexes  []uint64 `json:"large_funding_indexes" bson:"large_funding_indexes"`
	RecentFundingIndexes []uint64 `json:"recent_funding_indexes" bson:"recent_funding_indexes"`
}

type ProjectFundingDetail struct {
	Round         string `json:"round" bson:"round"`
	Date          string `json:"date" bson:"date"`
	Amount        uint64 `json:"amount" bson:"amount"`
	Valuation     uint64 `json:"valuation" bson:"valuation"`
	Investors     string `json:"investors" bson:"investors"`
	LeadInvestors string `json:"lead_investors" bson:"lead_investors"`

	InvestorsRefactor     []primitive.ObjectID `json:"investors_refactor" bson:"investors_refactor"`
	LeadInvestorsRefactor []primitive.ObjectID `json:"lead_investors_refactor" bson:"lead_investors_refactor"`

	// internal use
	InternalDate JSONTime `json:"-" bson:"internal_date"`
}

type ProjectFunding struct {
	TopInvestors   []primitive.ObjectID   `json:"top_investors" bson:"top_investors"`
	FundingDetails []ProjectFundingDetail `json:"funding_details" bson:"funding_details"`
	Reference      string                 `json:"reference" bson:"reference"`

	// auto generate
	Highlights ProjectFundingHighlights `json:"highlights" bson:"highlights"`
}

type DistributionInfo struct {
	Slice      string `json:"slice" bson:"slice"`
	Percentage uint64 `json:"percentage" bson:"percentage"`
}

type ProjectTokenomics struct {
	TokenIssuance                 bool               `json:"token_issuance" bson:"token_issuance"`
	TokenName                     string             `json:"token_name" bson:"token_name"`
	TokenSymbol                   string             `json:"token_symbol" bson:"token_symbol"`
	CirculatingSupply             float64            `json:"circulating_supply" bson:"circulating_supply"`
	TotalSupply                   float64            `json:"total_supply" bson:"total_supply"`
	TokenIssuanceDate             string             `json:"token_issuance_date" bson:"token_issuance_date"`
	InitialDistributionPictureURL string             `json:"initial_distribution_picture_url" bson:"initial_distribution_picture_url"`
	InitialDistribution           []DistributionInfo `json:"initial_distribution" bson:"initial_distribution"`
	InitialDistributionSourceLink string             `json:"initial_distribution_source_link" bson:"initial_distribution_source_link"`
	Description                   string             `json:"description" bson:"description"`
	MetricsLink                   string             `json:"metrics_link" bson:"metrics_link"`

	// TODO: fetch from website
	MetricsLinkLogoURL string `json:"metrics_link_logo_url" bson:"metrics_link_logo_url"`

	HoldersLink string `json:"holders_link" bson:"holders_link"`

	// TODO: fetch from website
	HoldersLinkLogoURL string `json:"holders_link_logo_url" bson:"holders_link_logo_url"`

	BigEventsLink string `json:"big_events_link" bson:"big_events_link"`

	// TODO: fetch from website
	BigEventsLinkLogoURL string `json:"big_events_link_logo_url" bson:"big_events_link_logo_url"`

	Reference string `json:"reference" bson:"reference"`

	// internal use
	InternalTokenIssuanceDate JSONTime `json:"-" bson:"internal_token_issuance_date"`

	// zh
	DescriptionZH string `json:"-" bson:"description_zh"`
}

type ProjectEcosystem struct {
	TotalAmount           uint64               `json:"total_amount" bson:"total_amount"`
	GrowthCurvePictureURL string               `json:"growth_curve_picture_url" bson:"growth_curve_picture_url"`
	GrowthCurveSourceLink string               `json:"growth_curve_source_link" bson:"growth_curve_source_link"`
	TopProjects           []primitive.ObjectID `json:"top_projects" bson:"top_projects"`
	Reference             string               `json:"reference" bson:"reference"`
}

type ProjectExchanges struct {
	BinanceLink  string `json:"binance_link" bson:"binance_link"`
	OKXLink      string `json:"okx_link" bson:"okx_link"`
	CoinbaseLink string `json:"coinbase_link" bson:"coinbase_link"`
	KrakenLink   string `json:"kraken_link" bson:"kraken_link"`
	BitstampLink string `json:"bitstamp_link" bson:"bitstamp_link"`
	KuCoinLink   string `json:"kucoin_link" bson:"kucoin_link"`
}

// TODO: timed automatic pull
type ProjectSocials struct {
	GithubCommits      uint64 `json:"github_commits" bson:"github_commits"`
	GithubStars        uint64 `json:"github_stars" bson:"github_stars"`
	GithubForks        uint64 `json:"github_forks" bson:"github_forks"`
	GithubContributors uint64 `json:"github_contributors" bson:"github_contributors"`
	GithubFollowers    uint64 `json:"github_followers" bson:"github_followers"`
	TwitterFollowers   uint64 `json:"twitter_followers" bson:"twitter_followers"`
	RedditFollowers    uint64 `json:"reddit_followers" bson:"reddit_followers"`
}

type ProjectBusinessModel struct {
	Model        string `json:"model" bson:"model"`
	AnnualIncome uint64 `json:"annual_income" bson:"annual_income"`
	Description  string `json:"description" bson:"description"`

	// zh
	DescriptionZH string `json:"-" bson:"description_zh"`
}

type ProjectProfitability struct {
	BusinessModels         []ProjectBusinessModel `json:"business_models" bson:"business_models"`
	FinancialStatementLink string                 `json:"financial_statement_link" bson:"financial_statement_link"`

	// TODO: fetch from website
	FinancialStatementLinkLogoURL string `json:"financial_statement_link_logo_url" bson:"financial_statement_link_logo_url"`

	Reference string `json:"reference" bson:"reference"`
}

type ProjectStatus = uint32

const (
	ProjectStatusIndexed ProjectStatus = iota
	ProjectStatusPublished
	ProjectStatusReedited
)

type ProjectCompletionStatus = uint32

const (
	ProjectCompletionStatusIncomplete ProjectCompletionStatus = iota
	ProjectCompletionStatusCoreDataComplete
	ProjectCompletionStatusComplete
)

type FieldCompletion struct {
	Field      string `json:"field" bson:"field"`
	Completion int    `json:"completion" bson:"completion"`
}

type CrawlerInfo struct {
	Crawler   string   `json:"crawler" bson:"crawler"`
	IsCrawled bool     `json:"is_crawled" bson:"is_crawled"`
	CrawlTime JSONTime `json:"crawl_time" bson:"crawl_time"`
}

type ProjectInternalInfo struct {
	Status ProjectStatus `json:"status" bson:"status"`

	ComponentAutoFillStatus []string `json:"component_auto_fill_status" bson:"component_auto"`

	// 完善项目详情的人用户id
	Completer primitive.ObjectID `json:"completer" bson:"completer"`

	CompletionStatus ProjectCompletionStatus `json:"completion_status" bson:"completion_status"`

	// 字段完成度
	FieldsCompletion []FieldCompletion `json:"fields_completion" bson:"fields_completion"`

	// 爬虫信息
	FetchedCrawlers map[string]CrawlerInfo `json:"fetched_crawlers" bson:"fetched_crawlers"`
}

type Project struct {
	BaseModel           `bson:"inline"`
	ProjectInternalInfo `bson:"inline"`

	Basic         ProjectBasic         `json:"basic" bson:"basic"`
	RelatedLinks  ProjectRelatedLinks  `json:"related_links" bson:"related_links"`
	Team          ProjectTeam          `json:"team" bson:"team"`
	Funding       ProjectFunding       `json:"funding" bson:"funding"`
	Tokenomics    ProjectTokenomics    `json:"tokenomics" bson:"tokenomics"`
	Ecosystem     ProjectEcosystem     `json:"ecosystem" bson:"ecosystem"`
	Profitability ProjectProfitability `json:"profitability" bson:"profitability"`
	Exchanges     ProjectExchanges     `json:"exchanges" bson:"exchanges"`
	Socials       ProjectSocials       `json:"socials" bson:"socials"`
}

var relatedLinksFilter = map[string]struct{}{
	"official website":    {},
	"Github":              {},
	"Twitter":             {},
	"Discord":             {},
	"Facebook":            {},
	"Telegram":            {},
	"Forum":               {},
	"LinkedIn":            {},
	"White Paper":         {},
	"Blog":                {},
	"Explorer":            {},
	"Token Unlocks":       {},
	"Whales Hold":         {},
	"More Token Metrics":  {},
	"Financial Statement": {},
}

func (p *Project) calculateCompletionInfo() {
	var basicTotalFieldNum, basicNotEmptyFieldNum int
	{
		if p.Basic.Name != "" {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++

		if p.Basic.LogoURL != "" {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++

		if p.Basic.Description != "" {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++

		if len(p.Basic.Chains) != 0 {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++

		if len(p.Basic.Tracks) != 0 {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++

		if len(p.Basic.Tags) != 0 {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++

		if len(p.Basic.Influences) != 0 {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++

		if p.Basic.FoundedDate != "" {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++

		if p.Basic.LaunchDate != "" {
			basicNotEmptyFieldNum++
		}
		basicTotalFieldNum++
	}

	relatedLinksTotalFieldNum := len(relatedLinksFilter)
	var relatedLinksNotEmptyFieldNum int
	{
		for _, link := range p.RelatedLinks {
			if _, ok := relatedLinksFilter[link.Type]; ok {
				relatedLinksNotEmptyFieldNum++
			}
		}
	}

	var teamTotalFieldNum, teamNotEmptyFieldNum int
	{
		if len(p.Team.Impressions) != 0 {
			teamNotEmptyFieldNum++
		}
		teamTotalFieldNum++

		if len(p.Team.Members) != 0 {
			teamNotEmptyFieldNum++
		}
		teamTotalFieldNum++
	}

	var fundingTotalFieldNum, fundingNotEmptyFieldNum int
	{
		if len(p.Funding.FundingDetails) != 0 {
			fundingNotEmptyFieldNum++
		}
		fundingTotalFieldNum++
	}

	var tokenTotalFieldNum, tokenNotEmptyFieldNum int
	{
		if p.Tokenomics.TokenIssuance {
			tokenNotEmptyFieldNum++
		}
		tokenTotalFieldNum++

		if p.Tokenomics.TokenSymbol != "" {
			tokenNotEmptyFieldNum++
		}
		tokenTotalFieldNum++

		if p.Tokenomics.TokenIssuanceDate != "" {
			tokenNotEmptyFieldNum++
		}
		tokenTotalFieldNum++

		if len(p.Tokenomics.InitialDistribution) != 0 {
			tokenNotEmptyFieldNum++
		}
		tokenTotalFieldNum++

		if len(p.Tokenomics.Description) != 0 {
			tokenNotEmptyFieldNum++
		}
		tokenTotalFieldNum++
	}

	var profitabilityTotalFieldNum, profitabilityNotEmptyFieldNum int
	{
		if len(p.Profitability.BusinessModels) != 0 {
			profitabilityNotEmptyFieldNum++
		}
		profitabilityTotalFieldNum++
	}

	var fieldsCompletion []FieldCompletion
	fieldsCompletion = append(fieldsCompletion, FieldCompletion{
		Field: "total",
		Completion: (basicNotEmptyFieldNum + relatedLinksNotEmptyFieldNum + teamNotEmptyFieldNum + fundingNotEmptyFieldNum + tokenNotEmptyFieldNum + profitabilityNotEmptyFieldNum) * 100 /
			(basicTotalFieldNum + relatedLinksTotalFieldNum + teamTotalFieldNum + fundingTotalFieldNum + tokenTotalFieldNum + profitabilityTotalFieldNum),
	})

	fieldsCompletion = append(fieldsCompletion, FieldCompletion{
		Field:      "basic",
		Completion: basicNotEmptyFieldNum * 100 / basicTotalFieldNum,
	})
	fieldsCompletion = append(fieldsCompletion, FieldCompletion{
		Field:      "related_links",
		Completion: relatedLinksNotEmptyFieldNum * 100 / relatedLinksTotalFieldNum,
	})
	fieldsCompletion = append(fieldsCompletion, FieldCompletion{
		Field:      "team",
		Completion: teamNotEmptyFieldNum * 100 / teamTotalFieldNum,
	})
	fieldsCompletion = append(fieldsCompletion, FieldCompletion{
		Field:      "funding",
		Completion: fundingNotEmptyFieldNum * 100 / fundingTotalFieldNum,
	})
	fieldsCompletion = append(fieldsCompletion, FieldCompletion{
		Field:      "tokenomics",
		Completion: tokenNotEmptyFieldNum * 100 / tokenTotalFieldNum,
	})
	fieldsCompletion = append(fieldsCompletion, FieldCompletion{
		Field:      "profitability",
		Completion: profitabilityNotEmptyFieldNum * 100 / profitabilityTotalFieldNum,
	})
	p.ProjectInternalInfo.FieldsCompletion = fieldsCompletion
}

func (p *Project) calculateFundingInfo(cfg *config.Config) {
	totalFundingAmount := uint64(0)
	var largeFundingIndexes []uint64
	var recentFundingIndexes []uint64

	totalInvestorsMap := make(map[primitive.ObjectID]struct{})
	totalLeadInvestorsMap := make(map[primitive.ObjectID]struct{})
	// filter repeated investor
	for i, detail := range p.Funding.FundingDetails {
		var investorsRefactor []primitive.ObjectID
		investorsRefactorMap := make(map[primitive.ObjectID]struct{})
		for _, inv := range detail.InvestorsRefactor {
			totalInvestorsMap[inv] = struct{}{}
			if _, ok := investorsRefactorMap[inv]; !ok {
				investorsRefactor = append(investorsRefactor, inv)
				investorsRefactorMap[inv] = struct{}{}
			}
		}

		var leadInvestorsRefactor []primitive.ObjectID
		leadInvestorsRefactorMap := make(map[primitive.ObjectID]struct{})
		for _, inv := range detail.LeadInvestorsRefactor {
			totalLeadInvestorsMap[inv] = struct{}{}
			if _, ok := leadInvestorsRefactorMap[inv]; !ok {
				leadInvestorsRefactor = append(leadInvestorsRefactor, inv)
				leadInvestorsRefactorMap[inv] = struct{}{}
			}
		}

		detail.InvestorsRefactor = investorsRefactor
		detail.LeadInvestorsRefactor = leadInvestorsRefactor
		p.Funding.FundingDetails[i] = detail

		totalFundingAmount += detail.Amount
		if detail.Amount > cfg.App.CaculateLimit.FinancingAmountLimit {
			largeFundingIndexes = append(largeFundingIndexes, uint64(i))
		}
		if detail.Date != "" {
			if int64(time.Since(time.Time(detail.InternalDate))/(24*time.Hour)) < cfg.App.CaculateLimit.FinancingTimeLimit {
				recentFundingIndexes = append(recentFundingIndexes, uint64(i))
			}
		}
	}

	highlights := ProjectFundingHighlights{}
	highlights.FundingRounds = uint64(len(p.Funding.FundingDetails))
	highlights.TotalFundingAmount = totalFundingAmount
	highlights.LeadInvestorsNumber = uint64(len(totalLeadInvestorsMap))
	highlights.InvestorsNumber = uint64(len(totalInvestorsMap))
	highlights.LargeFundingIndexes = largeFundingIndexes
	highlights.RecentFundingIndexes = recentFundingIndexes
	p.Funding.Highlights = highlights
}

func (p *Project) calculateExchanges() {
	tokenSymbol := p.Tokenomics.TokenSymbol
	if tokenSymbol == "" {
		return
	}
	p.Exchanges.BinanceLink = "https://www.binance.com/en/trade/" + tokenSymbol + "_USDT"
	p.Exchanges.CoinbaseLink = "https://pro.coinbase.com/trade/" + tokenSymbol + "-USD"
	p.Exchanges.KrakenLink = "https://trade.kraken.com/markets/kraken/" + tokenSymbol + "/USD"
	p.Exchanges.KuCoinLink = "https://trade.kucoin.com/" + tokenSymbol + "-usdt"
	p.Exchanges.BitstampLink = "https://www.bitstamp.net/trade/" + tokenSymbol + "/USD"
	p.Exchanges.OKXLink = "https://www.okex.com/markets/spot-info/" + tokenSymbol + "-usdt"
}

func (p *Project) CalculateDerivedData(cfg *config.Config) {
	if p.RelatedLinks == nil {
		p.RelatedLinks = ProjectRelatedLinks([]LinkInfo{})
	}
	if p.FetchedCrawlers == nil {
		p.FetchedCrawlers = make(map[string]CrawlerInfo)
	}
	isOpenSource := false
	for _, link := range p.RelatedLinks {
		if (link.Type == "Github" || link.Type == "GitHub" || link.Type == "github") && link.Link != "" {
			isOpenSource = true
			break
		}
	}
	p.Basic.IsOpenSource = isOpenSource

	p.calculateCompletionInfo()
	p.calculateFundingInfo(cfg)
	p.calculateExchanges()
}
