package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/core/datasource"
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

const officialWebsiteLinkType = "official website"

type ProjectService struct {
	baseComponent     *base.Component
	projectDao        *dao.ProjectDao
	miscDao           *dao.MiscDao
	marketDatasource  *datasource.Market
	metricsDatasource *datasource.Metrics
	socialDatasource  *datasource.Social
}

func NewProjectService(baseComponent *base.Component, projectDao *dao.ProjectDao, miscDao *dao.MiscDao, marketDatasource *datasource.Market, metricsDatasource *datasource.Metrics, socialDatasource *datasource.Social) *ProjectService {
	return &ProjectService{
		baseComponent:     baseComponent,
		projectDao:        projectDao,
		miscDao:           miscDao,
		marketDatasource:  marketDatasource,
		metricsDatasource: metricsDatasource,
		socialDatasource:  socialDatasource,
	}
}

func (s *ProjectService) parseToProjectBasicOutput(ctx *reqctx.ReqCtx, info *model.Project) (entity.ProjectBasicOutput, error) {
	var projectBasicOutput entity.ProjectBasicOutput
	var err error
	projectBasicOutput.FromModel(ctx.IsZHLang, &info.Basic)
	projectBasicOutput.Chains, err = s.miscDao.ChainBatchQuery(ctx, model.ObjIDsToStrings(info.Basic.Chains))
	if err != nil {
		return entity.ProjectBasicOutput{}, err
	}
	projectBasicOutput.Tracks, err = s.miscDao.TrackBatchQuery(ctx, model.ObjIDsToStrings(info.Basic.Tracks))
	if err != nil {
		return entity.ProjectBasicOutput{}, err
	}
	projectBasicOutput.Tags, err = s.miscDao.TagBatchQuery(ctx, model.ObjIDsToStrings(info.Basic.Tags))
	if err != nil {
		return entity.ProjectBasicOutput{}, err
	}
	return projectBasicOutput, nil
}

func (s *ProjectService) parseToProjectTeamOutput(ctx *reqctx.ReqCtx, info *model.Project) (entity.ProjectTeamOutput, error) {
	var projectTeamOutput entity.ProjectTeamOutput
	var err error
	projectTeamOutput.FromModel(ctx.IsZHLang, &info.Team)
	projectTeamOutput.Impressions, err = s.miscDao.TeamImpressionBatchQuery(ctx, model.ObjIDsToStrings(info.Team.Impressions))
	if err != nil {
		return entity.ProjectTeamOutput{}, err
	}
	return projectTeamOutput, nil
}

func (s *ProjectService) parseToProjectFundingOutput(ctx *reqctx.ReqCtx, info *model.Project) (entity.ProjectFundingOutput, error) {
	var projectFundingOutput entity.ProjectFundingOutput
	var err error
	projectFundingOutput.FromModel(ctx.IsZHLang, &info.Funding)
	projectFundingOutput.TopInvestors, err = s.miscDao.InvestorBatchQuery(ctx, model.ObjIDsToStrings(info.Funding.TopInvestors))
	if err != nil {
		return entity.ProjectFundingOutput{}, err
	}
	for i, e := range projectFundingOutput.FundingDetails {
		e.InvestorsRefactor, err = s.miscDao.InvestorBatchQuery(ctx, model.ObjIDsToStrings(info.Funding.FundingDetails[i].InvestorsRefactor))
		if err != nil {
			return entity.ProjectFundingOutput{}, err
		}
		e.LeadInvestorsRefactor, err = s.miscDao.InvestorBatchQuery(ctx, model.ObjIDsToStrings(info.Funding.FundingDetails[i].LeadInvestorsRefactor))
		if err != nil {
			return entity.ProjectFundingOutput{}, err
		}
	}

	for j, e := range projectFundingOutput.FundingDetails {
		e.InvestorsRefactor, err = s.miscDao.InvestorBatchQuery(ctx, model.ObjIDsToStrings(info.Funding.FundingDetails[j].InvestorsRefactor))
		if err != nil {
			return entity.ProjectFundingOutput{}, err
		}
		e.LeadInvestorsRefactor, err = s.miscDao.InvestorBatchQuery(ctx, model.ObjIDsToStrings(info.Funding.FundingDetails[j].LeadInvestorsRefactor))
		if err != nil {
			return entity.ProjectFundingOutput{}, err
		}
	}

	return projectFundingOutput, nil
}

func (s *ProjectService) parseToProjectEcosystemOutput(ctx *reqctx.ReqCtx, isView bool, info *model.Project) (entity.ProjectEcosystemOutput, error) {
	var projectEcosystemOutput entity.ProjectEcosystemOutput
	var err error
	projectEcosystemOutput.FromModel(ctx.IsZHLang, &info.Ecosystem)

	ids := model.ObjIDsToStrings(info.Ecosystem.TopProjects)
	var list []*entity.ProjectSimpleOutputQueryWrapper
	if err = s.projectDao.CustomBatchQuery(ctx, isView, ids, &list); err != nil {
		return entity.ProjectEcosystemOutput{}, err
	}
	list = dao.CollateBatchQueryResult(ids, list)

	for _, v := range list {
		if v == nil {
			return entity.ProjectEcosystemOutput{}, errcode.ErrProjectNotExist.Wrap(fmt.Sprintf("top project[%s] not found", v.ID.Hex()))
		}

		var officialWebsite string
		for _, link := range v.RelatedLinks {
			if link.Type == officialWebsiteLinkType {
				officialWebsite = link.Link
				break
			}
		}
		e := entity.ProjectSimpleOutput{
			ProjectID:       v.BaseModel.ID,
			Name:            v.Name,
			LogoURL:         v.LogoURL,
			Description:     v.Description,
			OfficialWebsite: officialWebsite,
		}
		if ctx.IsZHLang {
			if v.DescriptionZH != "" {
				e.Description = v.DescriptionZH
			}
		}
		projectEcosystemOutput.TopProjects = append(projectEcosystemOutput.TopProjects, e)
	}

	return projectEcosystemOutput, nil
}

func (s *ProjectService) infoModelToEntity(ctx *reqctx.ReqCtx, isView bool, info *model.Project) (*entity.ProjectOutput, error) {
	var output entity.ProjectOutput
	var err error
	output.BaseModel = info.BaseModel
	output.ProjectInternalInfo = info.ProjectInternalInfo

	output.Basic, err = s.parseToProjectBasicOutput(ctx, info)
	if err != nil {
		return nil, err
	}
	output.RelatedLinks.FromModel(ctx.IsZHLang, &info.RelatedLinks)
	output.Team, err = s.parseToProjectTeamOutput(ctx, info)
	if err != nil {
		return nil, err
	}
	output.Funding, err = s.parseToProjectFundingOutput(ctx, info)
	if err != nil {
		return nil, err
	}
	output.Tokenomics.FromModel(ctx.IsZHLang, &info.Tokenomics)
	output.Ecosystem, err = s.parseToProjectEcosystemOutput(ctx, isView, info)
	if err != nil {
		return nil, err
	}
	output.Profitability.FromModel(ctx.IsZHLang, &info.Profitability)
	output.Exchanges.FromModel(ctx.IsZHLang, &info.Exchanges)
	output.Socials.FromModel(ctx.IsZHLang, &info.Socials)
	return &output, nil
}

func (s *ProjectService) infoEntityToModel(ctx *reqctx.ReqCtx, input *entity.ProjectInput) (*model.Project, error) {
	basic, err := input.Basic.ToModel(ctx.IsZHLang)
	if err != nil {
		return nil, err
	}
	relatedLinks, err := input.RelatedLinks.ToModel(ctx.IsZHLang)
	if err != nil {
		return nil, err
	}
	team, err := input.Team.ToModel(ctx.IsZHLang)
	if err != nil {
		return nil, err
	}
	funding, err := input.Funding.ToModel(s.baseComponent.Config, ctx.IsZHLang)
	if err != nil {
		return nil, err
	}
	tokenomics, err := input.Tokenomics.ToModel(ctx.IsZHLang)
	if err != nil {
		return nil, err
	}
	ecosystem, err := input.Ecosystem.ToModel(ctx.IsZHLang)
	if err != nil {
		return nil, err
	}
	profitability, err := input.Profitability.ToModel(ctx.IsZHLang)
	if err != nil {
		return nil, err
	}
	isOpenSource := false
	for _, link := range *relatedLinks {
		if link.Type == "Github" && link.Link != "" {
			isOpenSource = true
			break
		}
	}
	basic.IsOpenSource = isOpenSource

	return &model.Project{
		Basic:         *basic,
		RelatedLinks:  *relatedLinks,
		Team:          *team,
		Funding:       *funding,
		Tokenomics:    *tokenomics,
		Ecosystem:     *ecosystem,
		Profitability: *profitability,
	}, nil
}

// Info todo(lrx): add socials data
func (s *ProjectService) Info(ctx *reqctx.ReqCtx, req *entity.ProjectInfoReq) (*entity.ProjectInfoRes, error) {
	info, err := s.projectDao.Query(ctx, true, req.ID)
	if err != nil {
		return nil, err
	}

	infoEntity, err := s.infoModelToEntity(ctx, true, info)
	if err != nil {
		return nil, err
	}
	projectsMarketInfo := s.marketDatasource.FindProjectMarketInfo(req.ID)
	infoEntity.Basic.Coin.CurrentPrice = projectsMarketInfo.Price
	infoEntity.Basic.Coin.MarketCap = projectsMarketInfo.MarketCap
	if projectsMarketInfo.CirculatingSupply != 0 {
		infoEntity.Tokenomics.CirculatingSupply = projectsMarketInfo.CirculatingSupply
	}
	infoEntity.Rank = projectsMarketInfo.Rank
	socialInfo := s.socialDatasource.GetGithubInfo(req.ID)
	infoEntity.Socials.GithubStars = uint64(socialInfo.GithubStars)
	infoEntity.Socials.GithubForks = uint64(socialInfo.GithubForks)
	infoEntity.Socials.GithubCommits = uint64(socialInfo.GithubCommits)
	infoEntity.Socials.GithubContributors = uint64(socialInfo.GithubContributors)
	infoEntity.Socials.GithubFollowers = uint64(socialInfo.GithubFollowers)

	return &entity.ProjectInfoRes{
		ProjectOutput: *infoEntity,
	}, nil
}

func (s *ProjectService) SimpleInfo(ctx *reqctx.ReqCtx, req *entity.ProjectSimpleInfoReq) (*entity.ProjectSimpleInfoRes, error) {
	var res entity.ProjectSimpleOutputQueryWrapper
	err := s.projectDao.CustomQuery(ctx, true, req.ID, &res)
	if err != nil {
		return nil, err
	}

	var officialWebsite string
	for _, link := range res.RelatedLinks {
		if link.Type == officialWebsiteLinkType {
			officialWebsite = link.Link
			break
		}
	}

	e := entity.ProjectSimpleOutput{
		ProjectID:       res.BaseModel.ID,
		Name:            res.Name,
		LogoURL:         res.LogoURL,
		Description:     res.Description,
		OfficialWebsite: officialWebsite,
	}
	if ctx.IsZHLang {
		if res.DescriptionZH != "" {
			e.Description = res.DescriptionZH
		}
	}
	return &entity.ProjectSimpleInfoRes{
		ProjectSimpleOutput: e,
	}, nil
}

func (s *ProjectService) List(ctx *reqctx.ReqCtx, req *entity.ProjectListReq) (*entity.ProjectListRes, error) {
	var list []*entity.ProjectListElement

	var conditions bson.A
	conditions = append(conditions, bson.M{
		"is_deleted": false,
	})
	if req.Query != "" {
		conditions = append(conditions, bson.M{"$or": bson.A{
			bson.M{
				"basic.name": bson.M{"$regex": req.Query, "$options": "i"},
			},
			bson.M{
				"tokenomics.token_symbol": bson.M{"$regex": req.Query, "$options": "i"},
			},
		}})
	}
	if len(req.Conditions.Chains) != 0 {
		objIDs, err := dao.IDsToObjectIDs(req.Conditions.Chains)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, bson.M{
			"basic.chains": bson.M{"$in": objIDs},
		})
	}
	if len(req.Conditions.Tracks) != 0 {
		objIDs, err := dao.IDsToObjectIDs(req.Conditions.Tracks)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, bson.M{
			"basic.tracks": bson.M{"$in": objIDs},
		})
	}
	if len(req.Conditions.Investors) != 0 {
		objIDs, err := dao.IDsToObjectIDs(req.Conditions.Investors)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, bson.M{
			"funding.top_investors": bson.M{"$in": objIDs},
		})
	}
	if req.Conditions.FoundedDateRange != nil {
		if req.Conditions.FoundedDateRange.Min != 0 && req.Conditions.FoundedDateRange.Max != 0 {
			now := time.Now()
			if req.Conditions.FoundedDateRange.Min != 0 {
				minTime := now.AddDate(0, int(-req.Conditions.FoundedDateRange.Min), 0)
				conditions = append(conditions, bson.M{
					"basic.internal_founded_date": bson.M{"$lte": minTime},
				})
			}
			if req.Conditions.FoundedDateRange.Min != 0 {
				maxTime := now.AddDate(0, int(-req.Conditions.FoundedDateRange.Max), 0)
				conditions = append(conditions, bson.M{
					"basic.internal_founded_date": bson.M{"$gte": maxTime},
				})
			}
		}
	}
	// filter marketcap
	if req.Conditions.MarketCapRange != nil {
		if req.Conditions.MarketCapRange.Min != 0 && req.Conditions.MarketCapRange.Max != 0 {
			marketInfos, err := s.marketDatasource.FindProjectsByMarketCapSort(req.Conditions.MarketCapRange.Min, req.Conditions.MarketCapRange.Max)
			if err != nil {
				return nil, err
			}
			var a bson.A
			for _, marketInfo := range marketInfos {
				objID, err := primitive.ObjectIDFromHex(marketInfo.ID)
				if err != nil {
					return nil, err
				}
				a = append(a, objID)
			}
			conditions = append(conditions, bson.M{"_id": bson.M{"$in": a}})
		}
	}

	page := req.Page
	size := req.Size
	sortFields := map[string]bool{}
	if req.SortField != "" {
		if req.SortField == "marketcap" || req.SortField == "price" {
			//	load all and paging in memory
			page = 0
			size = 0
		} else if req.SortField == "total_funding" {
			sortFields["funding.highlights.total_funding_amount"] = req.IsAsc
		} else if req.SortField == "create_time" || req.SortField == "update_time" {
			sortFields[req.SortField] = req.IsAsc
		} else {
			req.SortField = "marketcap"
			req.IsAsc = false
		}
	} else {
		page = 0
		size = 0
		req.SortField = "marketcap"
		req.IsAsc = false
	}

	total, err := s.projectDao.CustomList(ctx, true, page, size, bson.M{"$and": conditions}, sortFields, &list)
	if err != nil {
		return nil, err
	}

	idToIndex := map[string]int{}
	ids := lo.Map(list, func(item *entity.ProjectListElement, index int) string {
		id := item.ID.Hex()
		idToIndex[id] = index
		return id
	})

	projectsMarketInfos := s.marketDatasource.FindProjectsMarketInfo(ids)
	if req.SortField == "marketcap" || req.SortField == "price" {
		var sortType datasource.SortType
		if req.SortField == "marketcap" {
			sortType = datasource.ProjectMarketInfosSortByMarketcap
		}
		if req.SortField == "price" {
			sortType = datasource.ProjectMarketInfosSortByPrice
		}

		datasource.ProjectMarketInfosSort{
			List:     projectsMarketInfos,
			SortType: sortType,
			IsAsc:    req.IsAsc,
		}.Sort()
		if req.Page != 0 && req.Size != 0 {
			start := int((req.Page - 1) * req.Size)
			end := int(req.Page * req.Size)
			if start > len(projectsMarketInfos)-1 {
				return &entity.ProjectListRes{
					List:  []*entity.ProjectListElement{},
					Total: total,
				}, nil
			}
			if end > len(projectsMarketInfos) {
				end = len(projectsMarketInfos)
			}

			var pageList []*entity.ProjectListElement
			for _, marketInfo := range projectsMarketInfos[start:end] {
				info := list[idToIndex[marketInfo.ID]]
				info.MarketCap = marketInfo.MarketCap
				info.Rank = marketInfo.Rank
				info.Price = marketInfo.Price
				info.Last7DaysPictureURL = marketInfo.Last7DaysKlinesDataPictureURL
				pageList = append(pageList, info)
			}
			list = pageList
		}
	} else {
		for idx, element := range list {
			marketInfo := projectsMarketInfos[idx]
			element.MarketCap = marketInfo.MarketCap
			element.Rank = marketInfo.Rank
			element.Price = marketInfo.Price
			element.Last7DaysPictureURL = marketInfo.Last7DaysKlinesDataPictureURL
		}
	}

	for _, element := range list {
		element.ChainObjs, err = s.miscDao.ChainBatchQuery(ctx, model.ObjIDsToStrings(element.Chains))
		if err != nil {
			return nil, err
		}
		element.TrackObjs, err = s.miscDao.TrackBatchQuery(ctx, model.ObjIDsToStrings(element.Tracks))
		if err != nil {
			return nil, err
		}
		element.TagObjs, err = s.miscDao.TagBatchQuery(ctx, model.ObjIDsToStrings(element.Tags))
		if err != nil {
			return nil, err
		}
		element.TotalFunding = element.Funding.Highlights.TotalFundingAmount
	}

	return &entity.ProjectListRes{
		List:  list,
		Total: total,
	}, nil
}

func (s *ProjectService) InfoCompare(ctx *reqctx.ReqCtx, req *entity.ProjectInfoCompareReq) (*entity.ProjectInfoCompareRes, error) {
	infos, err := s.projectDao.BatchQuery(ctx, true, req.DecodedProjectIDs)
	if err != nil {
		return nil, err
	}

	var res []entity.ProjectOutput
	for _, info := range infos {
		infoEntity, err := s.infoModelToEntity(ctx, true, info)
		if err != nil {
			return nil, err
		}
		projectsMarketInfo := s.marketDatasource.FindProjectMarketInfo(info.ID.Hex())
		infoEntity.Basic.Coin.CurrentPrice = projectsMarketInfo.Price
		infoEntity.Basic.Coin.MarketCap = projectsMarketInfo.MarketCap
		if projectsMarketInfo.CirculatingSupply != 0 {
			infoEntity.Tokenomics.CirculatingSupply = projectsMarketInfo.CirculatingSupply
		}
		infoEntity.Rank = projectsMarketInfo.Rank
		res = append(res, *infoEntity)
	}

	return &entity.ProjectInfoCompareRes{
		Infos: res,
	}, nil
}

func (s *ProjectService) MetricsCompare(ctx *reqctx.ReqCtx, req *entity.ProjectMetricsCompareReq) (*entity.ProjectMetricsCompareRes, error) {
	infos, err := s.projectDao.BatchQuery(ctx, true, req.DecodedProjectIDs)
	if err != nil {
		return nil, err
	}

	res := make(map[string][]entity.ProjectMetrics)
	for _, info := range infos {
		res[info.ID.Hex()] = []entity.ProjectMetrics{}
	}
	switch req.Type {
	case entity.MetricsTypeCirculatingMarketCap:
		for _, info := range infos {
			if info.Tokenomics.TokenSymbol == "" {
				continue
			}
			marketInfo := s.marketDatasource.FindProjectMarketInfo(info.ID.Hex())
			if marketInfo.CirculatingSupply == 0 || marketInfo.Price == 0 {
				continue
			}
			prices, dates, err := s.marketDatasource.FetchKlinesData(info.ID.Hex(), req.Interval, req.StartTime, req.EndTime)
			if err != nil {
				ctx.Logger.WithField("err", err).Warnf("Failed to fetch klines data for: %s", info.ID.Hex())
				continue
			}
			var metrics []entity.ProjectMetrics
			for i := range prices {
				metrics = append(metrics, entity.ProjectMetrics{
					Timestamp:            uint64(dates[i].Unix()),
					CirculatingMarketCap: uint64(prices[i] * marketInfo.CirculatingSupply),
				})
			}
			res[info.ID.Hex()] = metrics
		}
	case entity.MetricsTypeFullyDilutedValue:
		for _, info := range infos {
			if info.Tokenomics.TokenSymbol == "" {
				continue
			}
			marketInfo := s.marketDatasource.FindProjectMarketInfo(info.ID.Hex())
			if marketInfo.TotalSupply == 0 || marketInfo.Price == 0 {
				continue
			}
			prices, dates, err := s.marketDatasource.FetchKlinesData(info.ID.Hex(), req.Interval, req.StartTime, req.EndTime)
			if err != nil {
				ctx.Logger.WithField("err", err).Warnf("Failed to fetch klines data for: %s", info.ID.Hex())
				continue
			}
			var metrics []entity.ProjectMetrics
			for i := range prices {
				metrics = append(metrics, entity.ProjectMetrics{
					Timestamp:         uint64(dates[i].Unix()),
					FullyDilutedValue: uint64(prices[i] * marketInfo.TotalSupply),
				})
			}
			res[info.ID.Hex()] = metrics
		}

	case entity.MetricsTypeActiveAddresses:
		for _, info := range infos {
			metricsInfo := s.metricsDatasource.FindProjectMetricsInfo(info.Basic.Name)
			if len(metricsInfo.ActiveUsers) == 0 {
				ctx.Logger.Warnf("Failed to fetch active user data for: %s", info.ID.Hex())
				continue
			}
			nums, dates, err := metricsInfo.FetchActiveUsersByTimeRange(req.Interval, req.StartTime, req.EndTime)
			if err != nil {
				ctx.Logger.WithField("err", err).Warnf("Failed to fetch active user data for: %s", info.ID.Hex())
				continue
			}
			var metrics []entity.ProjectMetrics
			for i := range nums {
				metrics = append(metrics, entity.ProjectMetrics{
					Timestamp:       uint64(dates[i].Unix()),
					ActiveAddresses: uint64(nums[i]),
				})
			}
			res[info.ID.Hex()] = metrics
		}
	default:
		return nil, errcode.ErrRequestParameter.Wrap("unsupported metrics type: " + req.Type)
	}

	var availableRes []entity.ProjectMetrics
	for _, v := range res {
		if len(v) != 0 {
			availableRes = v
			break
		}
	}
	if len(availableRes) != 0 {
		for k, v := range res {
			if len(v) == 0 {
				res[k] = lo.Map(availableRes, func(item entity.ProjectMetrics, index int) entity.ProjectMetrics {
					return entity.ProjectMetrics{
						Timestamp: item.Timestamp,
					}
				})
			}
		}
	} else {
		var extractionInterval int64 = 24 * 60 * 60
		switch req.Interval {
		case "1d":
			extractionInterval = extractionInterval * 1
		case "1w":
			extractionInterval = extractionInterval * 7
		case "1M":
			extractionInterval = extractionInterval * 30
		default:
			return nil, errcode.ErrRequestParameter.Wrap("unsupported interval")
		}
		startTime := time.Unix(int64(req.StartTime), 0)
		if startTime.Hour() != 0 || startTime.Minute() != 0 || startTime.Second() != 0 {
			startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day()+1, 0, 0, 0, 0, startTime.Location())
		}
		endTime := time.Unix(int64(req.EndTime), 0)
		if endTime.Hour() != 0 || endTime.Minute() != 0 || endTime.Second() != 0 {
			endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day()-1, 0, 0, 0, 0, endTime.Location())
		}

		var emptyRes []entity.ProjectMetrics
		for i := startTime.Unix(); i <= endTime.Unix(); i += extractionInterval {
			emptyRes = append(emptyRes, entity.ProjectMetrics{
				Timestamp: uint64(i),
			})
		}
		for k := range res {
			res[k] = emptyRes
		}
	}

	return &entity.ProjectMetricsCompareRes{
		Metrics: res,
	}, nil
}

// nolint
func (s *ProjectService) getProjectCoin(tokenSymbol string) (*model.ProjectCoin, error) {
	if tokenSymbol == "" {
		return nil, errors.New("tokenSymbol can not be empty")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", s.baseComponent.Config.Datasource.Market.Cmc.APIEndpoint+"/cryptocurrency/quotes/latest", nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("symbol", tokenSymbol)

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", s.baseComponent.Config.Datasource.Market.Cmc.ApiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("error sending request to server")
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("error geting response from server")
	}

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	data, ok := result["data"].(map[string]any)
	if !ok {
		return nil, errors.New("unexpected JSON response format: missing 'data' field")
	}
	tokenData, ok := data[tokenSymbol].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("token data not found for symbol '%s'", tokenSymbol)
	}
	marketCapRankValue, ok := tokenData["cmc_rank"]
	if !ok {
		return nil, fmt.Errorf("marketCapRank not found or has invalid type for symbol '%s'", tokenSymbol)
	}
	marketCapRank, ok := marketCapRankValue.(float64)
	if !ok {
		return nil, fmt.Errorf("marketCapRank has invalid type for symbol '%s'", tokenSymbol)
	}
	marketCapValue, ok := tokenData["quote"].(map[string]any)["USD"].(map[string]any)["market_cap"]
	if !ok {
		return nil, fmt.Errorf("marketCap not found or has invalid type for symbol '%s'", tokenSymbol)
	}
	marketCap, ok := marketCapValue.(float64)
	if !ok {
		return nil, fmt.Errorf("marketCap has invalid type for symbol '%s'", tokenSymbol)
	}
	projectCoin := &model.ProjectCoin{
		MarketCapRank: uint(marketCapRank),
		MarketCap:     uint64(marketCap),
	}
	return projectCoin, nil
}
