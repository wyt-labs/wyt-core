package entity

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

type ProjectBasicInput struct {
	Name         string            `json:"name"`
	LogoURL      string            `json:"logo_url"`
	Description  string            `json:"description"`
	Chains       []string          `json:"chains"`
	Tracks       []string          `json:"tracks"`
	Tags         []string          `json:"tags"`
	Influences   []model.Influence `json:"influences"`
	FoundedDate  string            `json:"founded_date"`
	LaunchDate   string            `json:"launch_date"`
	IsOpenSource bool              `json:"is_open_source"`
	Reference    string            `json:"reference"`
}

func (i *ProjectBasicInput) ToModel(isZHLang bool) (*model.ProjectBasic, error) {
	var internalFoundedDate, internalLaunchDate time.Time
	var err error
	if i.FoundedDate != "" {
		internalFoundedDate, err = util.StringToDate(i.FoundedDate)
		if err != nil {
			return nil, errors.Wrap(err, "field `founded_date` field format error")
		}
	}
	if i.LaunchDate != "" {
		internalLaunchDate, err = util.StringToDate(i.LaunchDate)
		if err != nil {
			return nil, errors.Wrap(err, "field `launch_date` field format error")
		}
	}

	chainIDs, err := model.IDsToObjIDs(i.Chains)
	if err != nil {
		return nil, errors.Wrap(err, "field `chains` field format error")
	}
	trackIDs, err := model.IDsToObjIDs(i.Tracks)
	if err != nil {
		return nil, errors.Wrap(err, "field `tracks` field format error")
	}
	tagIDs, err := model.IDsToObjIDs(i.Tags)
	if err != nil {
		return nil, errors.Wrap(err, "field `tracks` field format error")
	}

	m := &model.ProjectBasic{
		Name:                i.Name,
		LogoURL:             i.LogoURL,
		Chains:              chainIDs,
		Tracks:              trackIDs,
		Tags:                tagIDs,
		Influences:          i.Influences,
		FoundedDate:         i.FoundedDate,
		LaunchDate:          i.LaunchDate,
		IsOpenSource:        i.IsOpenSource,
		Reference:           i.Reference,
		InternalFoundedDate: model.JSONTime(internalFoundedDate),
		InternalLaunchDate:  model.JSONTime(internalLaunchDate),
	}
	if !isZHLang {
		m.Description = i.Description
	} else {
		m.DescriptionZH = i.Description
	}
	return m, nil
}

type ProjectBasicOutput struct {
	Name         string            `json:"name"`
	LogoURL      string            `json:"logo_url"`
	Description  string            `json:"description"`
	Chains       []*model.Chain    `json:"chains"`
	Tracks       []*model.Track    `json:"tracks"`
	Tags         []*model.Tag      `json:"tags"`
	Influences   []model.Influence `json:"influences"`
	FoundedDate  string            `json:"founded_date"`
	LaunchDate   string            `json:"launch_date"`
	IsOpenSource bool              `json:"is_open_source"`
	Reference    string            `json:"reference"`
	Coin         model.ProjectCoin `json:"coin"`
}

func (i *ProjectBasicOutput) FromModel(isZHLang bool, m *model.ProjectBasic) {
	o := ProjectBasicOutput{
		Name:         m.Name,
		LogoURL:      m.LogoURL,
		Influences:   m.Influences,
		FoundedDate:  m.FoundedDate,
		LaunchDate:   m.LaunchDate,
		IsOpenSource: m.IsOpenSource,
		Reference:    m.Reference,
		Coin:         m.Coin,
	}
	if !isZHLang {
		o.Description = m.Description
	} else {
		o.Description = m.DescriptionZH
		if o.Description == "" {
			o.Description = m.Description
		}
	}
	*i = o
}

type ProjectRelatedLinksInput model.ProjectRelatedLinks

func (i *ProjectRelatedLinksInput) ToModel(isZHLang bool) (*model.ProjectRelatedLinks, error) {
	return (*model.ProjectRelatedLinks)(i), nil
}

type ProjectRelatedLinksOutput model.ProjectRelatedLinks

func (i *ProjectRelatedLinksOutput) FromModel(isZHLang bool, m *model.ProjectRelatedLinks) {
	*i = ProjectRelatedLinksOutput(*m)
}

type ProjectTeamInput struct {
	Impressions []string                  `json:"impressions"`
	Members     []model.ProjectTeamMember `json:"members"`
	Reference   string                    `json:"reference"`
}

func (i *ProjectTeamInput) ToModel(isZHLang bool) (*model.ProjectTeam, error) {
	impressionIDs, err := model.IDsToObjIDs(i.Impressions)
	if err != nil {
		return nil, errors.Wrap(err, "field `impressions` field format error")
	}
	m := &model.ProjectTeam{
		Impressions: impressionIDs,
		Members:     i.Members,
		Reference:   i.Reference,
	}

	return m, nil
}

type ProjectTeamOutput struct {
	Impressions []*model.TeamImpression   `json:"impressions"`
	Members     []model.ProjectTeamMember `json:"members"`
	Reference   string                    `json:"reference"`
}

func (i *ProjectTeamOutput) FromModel(isZHLang bool, m *model.ProjectTeam) {
	o := ProjectTeamOutput{
		Members:   m.Members,
		Reference: m.Reference,
	}
	*i = o
}

type ProjectFundingInput struct {
	TopInvestors   []string                     `json:"top_investors"`
	FundingDetails []model.ProjectFundingDetail `json:"funding_details"`
	Reference      string                       `json:"reference"`
}

func (i *ProjectFundingInput) ToModel(cfg *config.Config, isZHLang bool) (*model.ProjectFunding, error) {
	topInvestorIDs, err := model.IDsToObjIDs(i.TopInvestors)
	if err != nil {
		return nil, errors.Wrap(err, "field `top_investors` field format error")
	}
	var fundingDetails []model.ProjectFundingDetail
	totalFundingAmount := uint64(0)
	investorsMap := make(map[string]bool)
	leadInvestorsMap := make(map[string]bool)
	var largeFundingIndexes []uint64
	var recentFundingIndexes []uint64
	for index, detail := range i.FundingDetails {
		var internalDate time.Time
		var err error
		if detail.Date != "" {
			internalDate, err = util.StringToDate(detail.Date)
			if err != nil {
				return nil, errors.Wrap(err, "field `date` field format error")
			}
		}
		if detail.Amount == 0 {
			return nil, errors.New("field `amount` can not be zero")
		}

		fundingDetails = append(fundingDetails, model.ProjectFundingDetail{
			Round:         detail.Round,
			Date:          detail.Date,
			Amount:        detail.Amount,
			Valuation:     detail.Valuation,
			Investors:     detail.Investors,
			LeadInvestors: detail.LeadInvestors,
			InternalDate:  model.JSONTime(internalDate),
		})

		investors := strings.Split(detail.Investors, ",")
		leadInvestors := strings.Split(detail.LeadInvestors, ",")
		for _, investor := range investors {
			investorsMap[investor] = true
		}
		for _, leadInvestor := range leadInvestors {
			leadInvestorsMap[leadInvestor] = true
		}
		totalFundingAmount += detail.Amount
		if detail.Amount > cfg.App.CaculateLimit.FinancingAmountLimit {
			largeFundingIndexes = append(largeFundingIndexes, uint64(index))
		}
		if detail.Date != "" {
			if int64(time.Since(internalDate)/(24*time.Hour)) < cfg.App.CaculateLimit.FinancingTimeLimit {
				recentFundingIndexes = append(recentFundingIndexes, uint64(index))
			}
		}
	}
	highlights := model.ProjectFundingHighlights{}
	highlights.FundingRounds = uint64(len(fundingDetails))
	highlights.TotalFundingAmount = totalFundingAmount
	highlights.LeadInvestorsNumber = uint64(len(leadInvestorsMap))
	highlights.InvestorsNumber = uint64(len(investorsMap))
	highlights.LargeFundingIndexes = largeFundingIndexes
	highlights.RecentFundingIndexes = recentFundingIndexes

	m := &model.ProjectFunding{
		TopInvestors:   topInvestorIDs,
		FundingDetails: fundingDetails,
		Reference:      i.Reference,
		Highlights:     highlights,
	}

	return m, nil
}

type ProjectFundingDetailOutput struct {
	Round         string `json:"round" bson:"round"`
	Date          string `json:"date" bson:"date"`
	Amount        uint64 `json:"amount" bson:"amount"`
	Valuation     uint64 `json:"valuation" bson:"valuation"`
	Investors     string `json:"investors" bson:"investors"`
	LeadInvestors string `json:"lead_investors" bson:"lead_investors"`

	InvestorsRefactor     []*model.Investor `json:"investors_refactor" bson:"investors_refactor"`
	LeadInvestorsRefactor []*model.Investor `json:"lead_investors_refactor" bson:"lead_investors_refactor"`
}

type ProjectFundingOutput struct {
	TopInvestors   []*model.Investor              `json:"top_investors"`
	FundingDetails []*ProjectFundingDetailOutput  `json:"funding_details"`
	Reference      string                         `json:"reference"`
	Highlights     model.ProjectFundingHighlights `json:"highlights"`
}

func (i *ProjectFundingOutput) FromModel(isZHLang bool, m *model.ProjectFunding) {
	o := ProjectFundingOutput{
		FundingDetails: lo.Map(m.FundingDetails, func(item model.ProjectFundingDetail, index int) *ProjectFundingDetailOutput {
			return &ProjectFundingDetailOutput{
				Round:                 item.Round,
				Date:                  item.Date,
				Amount:                item.Amount,
				Valuation:             item.Valuation,
				Investors:             item.Investors,
				LeadInvestors:         item.LeadInvestors,
				InvestorsRefactor:     []*model.Investor{},
				LeadInvestorsRefactor: []*model.Investor{},
			}
		}),
		Reference:  m.Reference,
		Highlights: m.Highlights,
	}
	*i = o
}

type ProjectTokenomicsInput struct {
	TokenIssuance                 bool                     `json:"token_issuance"`
	TokenName                     string                   `json:"token_name"`
	TokenSymbol                   string                   `json:"token_symbol"`
	TokenIssuanceDate             string                   `json:"token_issuance_date"`
	InitialDistributionPictureURL string                   `json:"initial_distribution_picture_url"`
	InitialDistribution           []model.DistributionInfo `json:"initial_distribution"`
	InitialDistributionSourceLink string                   `json:"initial_distribution_source_link"`
	Description                   string                   `json:"description"`
	MetricsLink                   string                   `json:"metrics_link"`
	MetricsLinkLogoURL            string                   `json:"metrics_link_logo_url"`
	HoldersLink                   string                   `json:"holders_link"`
	HoldersLinkLogoURL            string                   `json:"holders_link_logo_url"`
	BigEventsLink                 string                   `json:"big_events_link"`
	BigEventsLinkLogoURL          string                   `json:"big_events_link_logo_url"`
	Reference                     string                   `json:"reference"`
}

func (i *ProjectTokenomicsInput) ToModel(isZHLang bool) (*model.ProjectTokenomics, error) {
	var internalTokenIssuanceDate time.Time
	var err error
	if i.TokenIssuanceDate != "" {
		internalTokenIssuanceDate, err = util.StringToDate(i.TokenIssuanceDate)
		if err != nil {
			return nil, errors.Wrap(err, "field `token_issuance_date` field format error")
		}
	}

	m := &model.ProjectTokenomics{
		TokenIssuance:                 i.TokenIssuance,
		TokenName:                     i.TokenName,
		TokenSymbol:                   i.TokenSymbol,
		TokenIssuanceDate:             i.TokenIssuanceDate,
		InitialDistributionPictureURL: i.InitialDistributionPictureURL,
		InitialDistribution:           i.InitialDistribution,
		InitialDistributionSourceLink: i.InitialDistributionSourceLink,
		MetricsLink:                   i.MetricsLink,
		MetricsLinkLogoURL:            i.MetricsLinkLogoURL,
		HoldersLink:                   i.HoldersLink,
		HoldersLinkLogoURL:            i.HoldersLinkLogoURL,
		BigEventsLink:                 i.BigEventsLink,
		BigEventsLinkLogoURL:          i.BigEventsLinkLogoURL,
		Reference:                     i.Reference,
		InternalTokenIssuanceDate:     model.JSONTime(internalTokenIssuanceDate),
	}
	if !isZHLang {
		m.Description = i.Description
	} else {
		m.DescriptionZH = i.Description
	}
	return m, nil
}

type ProjectTokenomicsOutput struct {
	TokenIssuance                 bool                     `json:"token_issuance"`
	TokenName                     string                   `json:"token_name"`
	TokenSymbol                   string                   `json:"token_symbol"`
	TokenIssuanceDate             string                   `json:"token_issuance_date"`
	CirculatingSupply             float64                  `json:"circulating_supply" `
	TotalSupply                   float64                  `json:"total_supply"`
	InitialDistributionPictureURL string                   `json:"initial_distribution_picture_url"`
	InitialDistribution           []model.DistributionInfo `json:"initial_distribution"`
	InitialDistributionSourceLink string                   `json:"initial_distribution_source_link"`
	Description                   string                   `json:"description"`
	MetricsLink                   string                   `json:"metrics_link"`
	MetricsLinkLogoURL            string                   `json:"metrics_link_logo_url"`
	HoldersLink                   string                   `json:"holders_link"`
	HoldersLinkLogoURL            string                   `json:"holders_link_logo_url"`
	BigEventsLink                 string                   `json:"big_events_link"`
	BigEventsLinkLogoURL          string                   `json:"big_events_link_logo_url"`
	Reference                     string                   `json:"reference"`
}

func (i *ProjectTokenomicsOutput) FromModel(isZHLang bool, m *model.ProjectTokenomics) {
	o := ProjectTokenomicsOutput{
		TokenIssuance:                 m.TokenIssuance,
		TokenName:                     m.TokenName,
		TokenSymbol:                   m.TokenSymbol,
		TokenIssuanceDate:             m.TokenIssuanceDate,
		CirculatingSupply:             m.CirculatingSupply,
		TotalSupply:                   m.TotalSupply,
		InitialDistributionPictureURL: m.InitialDistributionPictureURL,
		InitialDistribution:           m.InitialDistribution,
		InitialDistributionSourceLink: m.InitialDistributionSourceLink,
		MetricsLink:                   m.MetricsLink,
		MetricsLinkLogoURL:            m.MetricsLinkLogoURL,
		HoldersLink:                   m.HoldersLink,
		HoldersLinkLogoURL:            m.HoldersLinkLogoURL,
		BigEventsLink:                 m.BigEventsLink,
		BigEventsLinkLogoURL:          m.BigEventsLinkLogoURL,
		Reference:                     m.Reference,
	}
	if o.CirculatingSupply != 0 && o.TotalSupply == 0 {
		o.TotalSupply = o.CirculatingSupply
	}

	if !isZHLang {
		o.Description = m.Description
	} else {
		o.Description = m.DescriptionZH
		if o.Description == "" {
			o.Description = m.Description
		}
	}
	*i = o
}

type ProjectEcosystemInput struct {
	TotalAmount           uint64   `json:"total_amount"`
	GrowthCurvePictureURL string   `json:"growth_curve_picture_url"`
	GrowthCurveSourceLink string   `json:"growth_curve_source_link"`
	TopProjects           []string `json:"top_projects"`
	Reference             string   `json:"reference"`
}

func (i *ProjectEcosystemInput) ToModel(isZHLang bool) (*model.ProjectEcosystem, error) {
	topProjectIDs, err := model.IDsToObjIDs(i.TopProjects)
	if err != nil {
		return nil, errors.Wrap(err, "field `top_projects` field format error")
	}
	m := &model.ProjectEcosystem{
		TotalAmount:           i.TotalAmount,
		GrowthCurvePictureURL: i.GrowthCurvePictureURL,
		GrowthCurveSourceLink: i.GrowthCurveSourceLink,
		TopProjects:           topProjectIDs,
		Reference:             i.Reference,
	}

	return m, nil
}

type ProjectEcosystemOutput struct {
	TotalAmount           uint64                `json:"total_amount"`
	GrowthCurvePictureURL string                `json:"growth_curve_picture_url"`
	GrowthCurveSourceLink string                `json:"growth_curve_source_link"`
	TopProjects           []ProjectSimpleOutput `json:"top_projects"`
	Reference             string                `json:"reference"`
}

func (i *ProjectEcosystemOutput) FromModel(isZHLang bool, m *model.ProjectEcosystem) {
	o := ProjectEcosystemOutput{
		TotalAmount:           m.TotalAmount,
		GrowthCurvePictureURL: m.GrowthCurvePictureURL,
		GrowthCurveSourceLink: m.GrowthCurveSourceLink,
		Reference:             m.Reference,
	}
	*i = o
}

type ProjectProfitabilityInput struct {
	model.ProjectProfitability
}

func (i *ProjectProfitabilityInput) ToModel(isZHLang bool) (*model.ProjectProfitability, error) {
	businessModels := i.BusinessModels
	if isZHLang {
		for j, businessModel := range i.BusinessModels {
			businessModels[j] = model.ProjectBusinessModel{
				Model:         businessModel.Model,
				AnnualIncome:  businessModel.AnnualIncome,
				DescriptionZH: businessModel.Description,
			}
		}
	}

	m := &model.ProjectProfitability{
		BusinessModels:                businessModels,
		FinancialStatementLink:        i.FinancialStatementLink,
		FinancialStatementLinkLogoURL: i.FinancialStatementLinkLogoURL,
		Reference:                     i.Reference,
	}

	return m, nil
}

type ProjectProfitabilityOutput struct {
	model.ProjectProfitability
}

func (i *ProjectProfitabilityOutput) FromModel(isZHLang bool, m *model.ProjectProfitability) {
	businessModels := m.BusinessModels
	if isZHLang {
		for j, businessModel := range m.BusinessModels {
			description := businessModel.DescriptionZH
			if description == "" {
				description = businessModel.Description
			}
			businessModels[j] = model.ProjectBusinessModel{
				Model:        businessModel.Model,
				AnnualIncome: businessModel.AnnualIncome,
				Description:  description,
			}
		}
	}

	o := ProjectProfitabilityOutput{
		ProjectProfitability: model.ProjectProfitability{
			BusinessModels:                businessModels,
			FinancialStatementLink:        m.FinancialStatementLink,
			FinancialStatementLinkLogoURL: m.FinancialStatementLinkLogoURL,
			Reference:                     m.Reference,
		},
	}
	*i = o
}

type ProjectExchangesOutput struct {
	model.ProjectExchanges
}

func (i *ProjectExchangesOutput) FromModel(isZHLang bool, m *model.ProjectExchanges) {
	o := ProjectExchangesOutput{
		ProjectExchanges: *m,
	}
	*i = o
}

type ProjectSocialsOutput struct {
	model.ProjectSocials
}

func (i *ProjectSocialsOutput) FromModel(isZHLang bool, m *model.ProjectSocials) {
	o := ProjectSocialsOutput{
		ProjectSocials: *m,
	}
	*i = o
}

type ProjectInput struct {
	CompletionStatus model.ProjectCompletionStatus `json:"completion_status"`
	Basic            ProjectBasicInput             `json:"basic"`
	RelatedLinks     ProjectRelatedLinksInput      `json:"related_links"`
	Team             ProjectTeamInput              `json:"team"`
	Funding          ProjectFundingInput           `json:"funding"`
	Tokenomics       ProjectTokenomicsInput        `json:"tokenomics"`
	Ecosystem        ProjectEcosystemInput         `json:"ecosystem"`
	Profitability    ProjectProfitabilityInput     `json:"profitability"`
}

type ProjectSimpleOutputQueryWrapper struct {
	model.BaseModel           `bson:"inline"`
	model.ProjectInternalInfo `bson:"inline"`
	ProjectSimpleOutput       `json:"-" bson:"basic"`
	RelatedLinks              model.ProjectRelatedLinks `json:"related_links" bson:"related_links"`
}

type ProjectSimpleOutput struct {
	ProjectID       primitive.ObjectID `json:"id" bson:"-"`
	Name            string             `json:"name" bson:"name"`
	LogoURL         string             `json:"logo_url" bson:"logo_url"`
	Description     string             `json:"description" bson:"description"`
	OfficialWebsite string             `json:"official_website"  bson:"-"`
	DescriptionZH   string             `json:"-" bson:"description_zh"`
}

type ProjectOutput struct {
	model.BaseModel
	model.ProjectInternalInfo

	Rank          int                        `json:"rank"`
	Basic         ProjectBasicOutput         `json:"basic"`
	RelatedLinks  ProjectRelatedLinksOutput  `json:"related_links"`
	Team          ProjectTeamOutput          `json:"team"`
	Funding       ProjectFundingOutput       `json:"funding"`
	Tokenomics    ProjectTokenomicsOutput    `json:"tokenomics"`
	Ecosystem     ProjectEcosystemOutput     `json:"ecosystem"`
	Profitability ProjectProfitabilityOutput `json:"profitability"`
	Exchanges     ProjectExchangesOutput     `json:"exchanges"`
	Socials       ProjectSocialsOutput       `json:"socials"`
}

type ProjectAddReq struct {
	ProjectInput
}

type ProjectAddRes struct {
	ID primitive.ObjectID `json:"id"`
}

type ProjectUpdateReq struct {
	ID string `json:"id"`
	ProjectInput
}

type ProjectUpdateRes struct {
}

type ProjectSimpleUpdateReq struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	LogoURL         string `json:"logo_url" `
	Description     string `json:"description"`
	OfficialWebsite string `json:"official_website"`
}

type ProjectSimpleUpdateRes struct {
}

type ProjectPublishReq struct {
	ID string `json:"id"`
}

type ProjectPublishRes struct {
}

type ProjectStatusFilter = uint32

const (
	ProjectStatusFilterAll ProjectStatusFilter = iota
	ProjectStatusFilterUnpublished
	ProjectStatusFilterPublished
	ProjectStatusFilterCompleted
	ProjectStatusFilterIncompleted
	ProjectStatusFilterCoreCompleted
)

type ProjectAdminListReq struct {
	Page      uint64              `json:"page"`
	Size      uint64              `json:"size"`
	Query     string              `json:"query"`
	Status    ProjectStatusFilter `json:"status"`
	SortField string              `json:"sort_field"`
	IsAsc     bool                `json:"is_asc"`
}

type ProjectAdminListElementBasicInfo struct {
	Name   string               `json:"name" bson:"name"`
	Chains []primitive.ObjectID `json:"-" bson:"chains"`
	Tracks []primitive.ObjectID `json:"-" bson:"tracks"`
	Tags   []primitive.ObjectID `json:"-" bson:"tags"`
}

type ProjectAdminListElement struct {
	ID                               primitive.ObjectID `json:"id" bson:"_id"`
	ProjectAdminListElementBasicInfo `bson:"basic"`

	ChainObjs        []*model.Chain                `json:"chains" bson:"-"`
	TrackObjs        []*model.Track                `json:"tracks" bson:"-"`
	TagObjs          []*model.Tag                  `json:"tags" bson:"-"`
	Status           model.ProjectStatus           `json:"status" bson:"status"`
	CreateTime       model.JSONTime                `json:"create_time" bson:"create_time"`
	UpdateTime       model.JSONTime                `json:"update_time" bson:"update_time"`
	CompletionStatus model.ProjectCompletionStatus `json:"completion_status" bson:"completion_status"`
	FieldsCompletion []model.FieldCompletion       `json:"fields_completion" bson:"fields_completion"`
	FetchedCrawlers  map[string]model.CrawlerInfo  `json:"fetched_crawlers" bson:"fetched_crawlers"`
}

type ProjectAdminListRes struct {
	List  []*ProjectAdminListElement `json:"list"`
	Total int64                      `json:"total"`
}

type ProjectAdminInfoReq struct {
	ID string `json:"id" form:"id"`
}

type ProjectAdminInfoRes struct {
	ProjectOutput
}

type ProjectAdminSimpleInfoReq struct {
	ID string `json:"id" form:"id"`
}

type ProjectAdminSimpleInfoRes struct {
	ProjectSimpleOutput
}

type ProjectInfoReq struct {
	ID string `json:"id" form:"id"`
}

type ProjectInfoRes struct {
	ProjectOutput
}

type ProjectSimpleInfoReq struct {
	ID string `json:"id" form:"id"`
}

type ProjectSimpleInfoRes struct {
	ProjectSimpleOutput
}

type ProjectDeleteReq struct {
	ID string `json:"id"`
}

type ProjectDeleteRes struct {
}

type ProjectCalculateDerivedDataReq struct {
	IsView bool `json:"is_view" form:"is_view"`
}

type ProjectCalculateDerivedDataRes struct {
}

type ProjectListReqConditionsRange struct {
	Min uint64 `json:"min"`
	Max uint64 `json:"max"`
}

type ProjectListReqConditions struct {
	Chains           []string                       `json:"chains"`
	Tracks           []string                       `json:"tracks"`
	Investors        []string                       `json:"investors"`
	MarketCapRange   *ProjectListReqConditionsRange `json:"market_cap_range"`
	FoundedDateRange *ProjectListReqConditionsRange `json:"founded_date_range"`
}

type ProjectListReq struct {
	Page       uint64                   `json:"page"`
	Size       uint64                   `json:"size"`
	Query      string                   `json:"query"`
	Conditions ProjectListReqConditions `json:"conditions"`
	SortField  string                   `json:"sort_field"`
	IsAsc      bool                     `json:"is_asc"`
}

type ProjectListElementBasicInfo struct {
	Name    string               `json:"name" bson:"name"`
	LogoURL string               `json:"logo_url" bson:"logo_url"`
	Chains  []primitive.ObjectID `json:"-" bson:"chains"`
	Tracks  []primitive.ObjectID `json:"-" bson:"tracks"`
	Tags    []primitive.ObjectID `json:"-" bson:"tags"`
}

type ProjectListElementTokenomicsInfo struct {
	Symbol string `json:"symbol" bson:"token_symbol"`
}

type ProjectListElementFundingInfo struct {
	Highlights struct {
		TotalFundingAmount uint64 `json:"total_funding_amount" bson:"total_funding_amount"`
	} `json:"highlights" bson:"highlights"`
}

type ProjectListElement struct {
	ID                               primitive.ObjectID `json:"id" bson:"_id"`
	ProjectListElementBasicInfo      `bson:"basic"`
	ProjectListElementTokenomicsInfo `bson:"tokenomics"`
	Funding                          ProjectListElementFundingInfo `json:"-" bson:"funding"`
	ChainObjs                        []*model.Chain                `json:"chains" bson:"-"`
	TrackObjs                        []*model.Track                `json:"tracks" bson:"-"`
	TagObjs                          []*model.Tag                  `json:"tags" bson:"-"`
	TotalFunding                     uint64                        `json:"total_funding" bson:"-"`
	Price                            float64                       `json:"price" bson:"-"`
	MarketCap                        uint64                        `json:"market_cap" bson:"-"`
	Rank                             int                           `json:"rank" bson:"-"`
	Last7DaysPictureURL              string                        `json:"last_7_days_picture_url" bson:"-"`
	Status                           model.ProjectStatus           `json:"status" bson:"status"`
	CreateTime                       model.JSONTime                `json:"create_time" bson:"create_time"`
	UpdateTime                       model.JSONTime                `json:"update_time" bson:"update_time"`
}

type ProjectListRes struct {
	List  []*ProjectListElement `json:"list"`
	Total int64                 `json:"total"`
}

type ProjectInfoCompareReq struct {
	ProjectIDs string `form:"project-ids"`

	DecodedProjectIDs []string `form:"-"`
}

type ProjectInfoCompareRes struct {
	Infos []ProjectOutput `json:"infos"`
}

type MetricsType = string

const (
	MetricsTypeCirculatingMarketCap MetricsType = "Circulating-Market-Cap"
	MetricsTypeFullyDilutedValue    MetricsType = "Fully-Diluted-Value"
	MetricsTypeActiveAddresses      MetricsType = "Active-Addresses"
)

type ProjectMetricsCompareReq struct {
	ProjectIDs string      `form:"project-ids"`
	Type       MetricsType `form:"type"`
	StartTime  uint64      `form:"start-time"`
	EndTime    uint64      `form:"end-time"`
	Interval   string      `form:"interval"`

	DecodedProjectIDs []string `form:"-"`
}

type ProjectMetrics struct {
	Timestamp            uint64 `json:"timestamp"`
	CirculatingMarketCap uint64 `json:"circulating_market_cap"`
	FullyDilutedValue    uint64 `json:"fully_diluted_value"`
	ActiveAddresses      uint64 `json:"active_addresses"`
}

type ProjectMetricsCompareRes struct {
	Metrics map[string][]ProjectMetrics `json:"metrics"`
}
