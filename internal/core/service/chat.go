package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/core/datasource"
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/extension"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

const (
	networkErrMsg         = "Oops! Looks like something went wrong. You can try again or modify your question."
	projectNotFoundErrMsg = "Sorry, the project you are looking for cannot be found. Please double-check the project details and try again."
	intentionErrMsg       = "Apologies, we couldn't understand the command. Please ensure the command is valid and try again."
	traderNotFoundErrMsg  = "Trader not found. You can give me a solana address of another trader."
)

// const (
// 	chatSearchDEx                 = "^(WYT Swap|Uniswap|1inch|CoW Swap)$"
// 	chatSearchTokenRegEx0         = "^[A-Za-z\\d]{2,15}$"
// 	chatSearchTokenRegEx1         = "(?i)Search (\\w+)(?: token)? for me"
// 	chatSearchTokenRegEx2         = "(?i)Find (\\w+)(?: token)? for me"
// 	chatSearchTokenPriceRegEx1    = "(?i)What's the price of (\\w+)(?: token)*"
// 	chatSearchTokenPriceRegEx2    = "(?i)What's the market cap of (\\w+)(?: token)*"
// 	chatSearchTokenBaseRegEx      = "(?i)What is the official (?:website|Twitter|github) of (\\w+)(?: token)*"
// 	chatSearchTokenTeamRegEx      = "(?i)Who are (?:the team members|the CEO|the core members) of the (\\w+)?"
// 	chatSearchTokenFundRegEx1     = "(?i)(?:Who are the investors of|How is the financing of) the (\\w+)?"
// 	chatSearchTokenFundRegEx2     = "(?i)Does the (\\w+) have top institutional investment?"
// 	chatSearchTokenExchangesRegEx = "(?i)Which exchanges are the (\\w+) listed on?"
// 	chatSearchTokenEconomicRegEx1 = "(?i)What is the economic model of the (\\w+)?"
// 	chatSearchTokenEconomicRegEx2 = "(?i)How about (\\w+) issuance\\?"
// 	chatSearchTokenEconomicRegEx3 = "(?i)How (\\w+) is circulated and distributed?"
// 	chatCompareTokenRegEx1        = "(?i)Please compare (?:Token )?(\\w+) with (?:Token )?(\\w+)"
// 	chatCompareTokenRegEx2        = "(?i)Which is better, (?:Token )?(\\w+) or (?:Token )?(\\w+)\\?"
// 	chatCompareTokenRegEx3        = "(?i)(?:Token )?(\\w+) PK (?:Token )?(\\w+)"
// 	chatCompareTokenRegEx4        = "(?i)(\\w+) and (\\w+)(?: tokens)? are compared"
// 	chatCompareTokenRegEx5        = "(?i)Which is more valuable, (\\w+) or (\\w+)?"
// )

type ChatService struct {
	baseComponent    *base.Component
	projectDao       *dao.ProjectDao
	miscDao          *dao.MiscDao
	chatDao          *dao.ChatDao
	userPluginDao    *dao.UserPluginDao
	chatgptDriver    *extension.ChatgptDriver
	marketDatasource *datasource.Market
}

func NewChatService(
	baseComponent *base.Component,
	projectDao *dao.ProjectDao,
	miscDao *dao.MiscDao,
	chatDao *dao.ChatDao,
	chatgptDriver *extension.ChatgptDriver,
	marketDatasource *datasource.Market,
	userPluginDao *dao.UserPluginDao,
) (*ChatService, error) {
	return &ChatService{
		baseComponent:    baseComponent,
		projectDao:       projectDao,
		miscDao:          miscDao,
		chatDao:          chatDao,
		chatgptDriver:    chatgptDriver,
		marketDatasource: marketDatasource,
		userPluginDao:    userPluginDao,
	}, nil
}

func (s *ChatService) Create(ctx *reqctx.ReqCtx, req *entity.ChatCreateReq) (*entity.ChatCreateRes, error) {
	w := &model.ChatWindow{
		Title:        fmt.Sprintf("chat(%s)", time.Now().Format("2006-01-02 15:04")),
		TitleIsSet:   false,
		LastUserMsgs: []model.ChatMsg{},
		MsgNum:       0,
		ProjectId:    req.ProjectId,
	}
	if err := s.chatDao.ChatWindowAdd(ctx, w); err != nil {
		return nil, err
	}
	return &entity.ChatCreateRes{
		ID:        w.ID,
		ProjectId: w.ProjectId,
	}, nil
}

func (s *ChatService) List(ctx *reqctx.ReqCtx, req *entity.ChatListReq) (*entity.ChatListRes, error) {
	objID, err := primitive.ObjectIDFromHex(ctx.Caller)
	if err != nil {
		return nil, err
	}

	res, total, err := s.chatDao.ChatWindowList(ctx, req.Page, req.Size, bson.M{
		"is_deleted": false,
		"creator":    objID,
	}, nil)
	if err != nil {
		return nil, err
	}
	for _, element := range res {
		if element.ProjectId == "" {
			element.ProjectId = s.baseComponent.Config.AIBackend.ProjectId
		}
	}
	return &entity.ChatListRes{
		List:  res,
		Total: total,
	}, nil
}

func (s *ChatService) History(ctx *reqctx.ReqCtx, req *entity.ChatHistoryReq) (*entity.ChatHistoryRes, error) {
	w, err := s.chatDao.ChatWindowQuery(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if w.Creator.Hex() != ctx.Caller {
		return nil, errcode.ErrAccountPermission
	}

	res, total, err := s.chatDao.ChatHistoryList(ctx, req.Page, req.Size, bson.M{
		"is_deleted": false,
		"window_id":  w.ID,
	}, map[string]bool{"index": false})
	if err != nil {
		return nil, err
	}
	var msgs []*entity.ChatMsg
	for _, element := range res {
		var content any
		switch element.Msg.Role {
		case model.ChatMsgRoleUserBuiltin:
			content = element.Msg.ContentUserBuiltin
		case model.ChatMsgRoleUser:
			content = element.Msg.ContentUser
		case model.ChatMsgRoleSystem:
			content = element.Msg.ContentSystem
		case model.ChatMsgRoleAssistant:
			content = element.Msg.ContentAssistant
		}

		msgs = append(msgs, &entity.ChatMsg{
			Index:     element.Index,
			Timestamp: element.CreateTime,
			Role:      element.Msg.Role,
			Content:   content,
			ProjectId: element.ProjectId,
		})
	}
	pjId := w.ProjectId
	return &entity.ChatHistoryRes{
		Title:      w.Title,
		ProjectId:  pjId,
		CreateTime: w.CreateTime,
		MsgNum:     uint64(total),
		Msgs:       msgs,
	}, nil
}

func (s *ChatService) convertProjectInfo(ctx *reqctx.ReqCtx, projects []*model.Project, view string) ([]model.ChatContentAssistantProjectInfo, string, error) {
	res := lo.Map(projects, func(item *model.Project, index int) model.ChatContentAssistantProjectInfo {
		return model.ChatContentAssistantProjectInfo{
			ID: item.ID,
		}
	})

	tokenomicsSetter := func() {
		for i := range res {
			res[i].Tokenomics = &projects[i].Tokenomics
		}
	}
	profitabilitySetter := func() {
		for i := range res {
			res[i].Profitability = &projects[i].Profitability
		}
	}
	teamSetter := func() {
		for i := range res {
			res[i].Team = &projects[i].Team
		}
	}
	fundingSetter := func() error {
		for i := range res {
			topInverstors, err := s.miscDao.InvestorBatchQuery(ctx, lo.Map(lo.Uniq(projects[i].Funding.TopInvestors), func(item primitive.ObjectID, index int) string {
				return item.Hex()
			}))
			if err != nil {
				return err
			}
			fundingDetails := lo.Map(projects[i].Funding.FundingDetails, func(item model.ProjectFundingDetail, index int) *model.ChatContentAssistantProjectFundingDetail {
				return &model.ChatContentAssistantProjectFundingDetail{
					Round:                 item.Round,
					Date:                  item.Date,
					Amount:                item.Amount,
					Valuation:             item.Valuation,
					Investors:             item.Investors,
					LeadInvestors:         item.LeadInvestors,
					InvestorsRefactor:     []model.ChatContentAssistantProjectInfoInvestor{},
					LeadInvestorsRefactor: []model.ChatContentAssistantProjectInfoInvestor{},
				}
			})

			for j, e := range fundingDetails {
				investorsRefactor, err := s.miscDao.InvestorBatchQuery(ctx, model.ObjIDsToStrings(projects[i].Funding.FundingDetails[j].InvestorsRefactor))
				if err != nil {
					return err
				}
				leadInvestorsRefactor, err := s.miscDao.InvestorBatchQuery(ctx, model.ObjIDsToStrings(projects[i].Funding.FundingDetails[j].LeadInvestorsRefactor))
				if err != nil {
					return err
				}
				e.InvestorsRefactor = lo.Map(investorsRefactor, func(item *model.Investor, index int) model.ChatContentAssistantProjectInfoInvestor {
					return model.ChatContentAssistantProjectInfoInvestor{
						ID:               item.ID,
						Name:             item.Name,
						Description:      item.Description,
						AvatarURL:        item.AvatarURL,
						Subject:          item.Subject,
						Type:             item.Type,
						SocialMediaLinks: item.SocialMediaLinks,
					}
				})

				e.LeadInvestorsRefactor = lo.Map(leadInvestorsRefactor, func(item *model.Investor, index int) model.ChatContentAssistantProjectInfoInvestor {
					return model.ChatContentAssistantProjectInfoInvestor{
						ID:               item.ID,
						Name:             item.Name,
						Description:      item.Description,
						AvatarURL:        item.AvatarURL,
						Subject:          item.Subject,
						Type:             item.Type,
						SocialMediaLinks: item.SocialMediaLinks,
					}
				})
			}

			res[i].Funding = &model.ChatContentAssistantProjectInfoFunding{
				TopInvestors: lo.Map(topInverstors, func(item *model.Investor, index int) model.ChatContentAssistantProjectInfoInvestor {
					return model.ChatContentAssistantProjectInfoInvestor{
						ID:               item.ID,
						Name:             item.Name,
						Description:      item.Description,
						AvatarURL:        item.AvatarURL,
						Subject:          item.Subject,
						Type:             item.Type,
						SocialMediaLinks: item.SocialMediaLinks,
					}
				}),
				FundingDetails: fundingDetails,
				Reference:      projects[i].Funding.Reference,
				Highlights:     projects[i].Funding.Highlights,
			}
		}
		return nil
	}
	exchangesSetter := func() {
		for i := range res {
			res[i].Exchanges = &projects[i].Exchanges
		}
	}
	overviewSetter := func() error {
		for i := range res {
			marketInfo := s.marketDatasource.FindProjectMarketInfo(projects[i].ID.Hex())

			tracks, err := s.miscDao.TrackBatchQuery(ctx, model.ObjIDsToStrings(projects[i].Basic.Tracks))
			if err != nil {
				return err
			}
			tags, err := s.miscDao.TagBatchQuery(ctx, model.ObjIDsToStrings(projects[i].Basic.Tags))
			if err != nil {
				return err
			}
			teamImpressions, err := s.miscDao.TeamImpressionBatchQuery(ctx, model.ObjIDsToStrings(projects[i].Team.Impressions))
			if err != nil {
				return err
			}

			res[i].Overview = &model.ChatContentAssistantProjectOverview{
				Name:           projects[i].Basic.Name,
				Description:    projects[i].Basic.Description,
				LogoURL:        projects[i].Basic.LogoURL,
				TokenSymbol:    projects[i].Tokenomics.TokenSymbol,
				TokenPrice:     marketInfo.Price,
				TokenMarketCap: marketInfo.MarketCap,
				TeamImpressions: lo.Map(teamImpressions, func(item *model.TeamImpression, index int) model.ChatContentAssistantProjectInfoTeamImpression {
					return model.ChatContentAssistantProjectInfoTeamImpression{
						ID:          item.ID,
						Name:        item.Name,
						Description: item.Description,
					}
				}),
				Tracks: lo.Map(tracks, func(item *model.Track, index int) model.ChatContentAssistantProjectInfoTrack {
					return model.ChatContentAssistantProjectInfoTrack{
						ID:          item.ID,
						Name:        item.Name,
						Description: item.Description,
					}
				}),
				Tags: lo.Map(tags, func(item *model.Tag, index int) model.ChatContentAssistantProjectInfoTag {
					return model.ChatContentAssistantProjectInfoTag{
						ID:          item.ID,
						Name:        item.Name,
						Description: item.Description,
					}
				}),
				RelatedLinks: projects[i].RelatedLinks,
			}
		}
		return nil
	}
	switch view {
	case "tokenomics":
		tokenomicsSetter()
	case "profitability":
		profitabilitySetter()
	case "team":
		teamSetter()
	case "funding":
		if err := fundingSetter(); err != nil {
			return nil, "", err
		}
	case "exchanges":
		exchangesSetter()
	case "all":
		tokenomicsSetter()
		profitabilitySetter()
		teamSetter()
		if err := fundingSetter(); err != nil {
			return nil, "", err
		}
		exchangesSetter()
		if err := overviewSetter(); err != nil {
			return nil, "", err
		}
	default:
		view = "overview"
	}
	if err := overviewSetter(); err != nil {
		return nil, "", err
	}
	return res, view, nil
}

func (s *ChatService) fetchProjectData(ctx *reqctx.ReqCtx, projectIDs []string, view string) ([]model.ChatContentAssistantProjectInfo, string, error) {
	projects, err := s.projectDao.BatchQuery(ctx, true, projectIDs)
	if err != nil {
		return nil, "", err
	}
	return s.convertProjectInfo(ctx, projects, view)
}

func (s *ChatService) fetchProjectDataByKey(ctx *reqctx.ReqCtx, projectKeys []string, view string) ([]model.ChatContentAssistantProjectInfo, string, error) {
	projects, err := s.projectDao.BatchSearch(ctx, true, projectKeys)
	if err != nil {
		return nil, "", err
	}
	return s.convertProjectInfo(ctx, projects, view)
}

func (s *ChatService) Completions(ctx *reqctx.ReqCtx, req *entity.ChatCompletionsReq, useStructuredOutput bool) (*entity.ChatCompletionsRes, error) {
	w, err := s.chatDao.ChatWindowQuery(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	pjId := w.ProjectId
	s.baseComponent.Logger.WithFields(logrus.Fields{"window id:": req.ID}).Info("project id:", pjId)
	if req.ProjectId != "" && pjId != "" && req.ProjectId != pjId {
		return nil, errcode.ErrPjWinIdErr
	}
	if req.ProjectId == "" && pjId != "" {
		req.ProjectId = pjId
	}
	if pjId == "" && req.ProjectId != "" {
		pjId = req.ProjectId
	}
	if pjId == "" {
		// 默认是WYT的项目
		pjId = s.baseComponent.Config.AIBackend.ProjectId
	}

	if w.Creator.Hex() != ctx.Caller {
		return nil, errcode.ErrAccountPermission
	}

	now := time.Now()
	// user msg time and role
	userMsg := &model.ChatMsg{
		Timestamp: model.JSONTime(now),
		Role:      req.Type,
	}
	// ai msg
	aiMsg := &model.ChatMsg{
		Timestamp: model.JSONTime(now),
	}
	switch req.Type {
	case model.ChatMsgRoleUser:
		// ai msg
		aiMsg.ContentAssistant = &model.ChatContentAssistant{}
		aiMsg.Role = model.ChatMsgRoleAssistant
		// 本次问答结果
		var chatAIAnalyticalResult model.ChatAIAnalyticalResult
		// user msg content
		userMsg.ContentUser = &model.ChatContentUser{
			Content: req.Msg,
		}
		// 将本次user msg添加到user msg list
		w.LastUserMsgs = append(w.LastUserMsgs, *userMsg)
		if len(w.LastUserMsgs) >= 1 {
			w.LastUserMsgs = w.LastUserMsgs[len(w.LastUserMsgs)-1:]
		}

		err = func() error {
			questMsgs := make([]*extension.ChatgptMsg, 0)
			for _, msg := range w.LastUserMsgs {
				questMsgs = append(questMsgs, &extension.ChatgptMsg{
					Role:    msg.Role,
					Content: msg.ContentUser.Content,
				})
			}
			var retStr string
			var fcRet *model.FuncCallingRet

			if useStructuredOutput {
				// always call this
				retStr, fcRet, err = s.chatgptDriver.DocsFaq(ctx.Ctx, questMsgs, pjId)
			} else {
				retStr, fcRet, err = s.chatgptDriver.ChatWithStructuredOutput(questMsgs)
			}
			if err != nil {
				ctx.AddCustomLogField("ai_err", err)
				if strings.Contains(err.Error(), "not found") {
					return errors.New(traderNotFoundErrMsg)
				}
				return errors.New(networkErrMsg)
			}
			if retStr != "" {
				if err := json.Unmarshal([]byte(retStr), &chatAIAnalyticalResult); err != nil {
					ctx.AddCustomLogField("ai_err", err)
					return errors.New(intentionErrMsg)
				}
				ctx.AddCustomLogField("ai_analytical", retStr)
				if (chatAIAnalyticalResult.Intention != model.ChatAIAnalyticalIntentionSearch) &&
					(chatAIAnalyticalResult.Intention != model.ChatAIAnalyticalIntentionCompare) {
					aiMsg.ContentAssistant.Type = model.ChatContentAssistantTypeGeneral
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
					aiMsg.ContentAssistant.Tips = "Here is the answer:"
					aiMsg.ContentAssistant.Type = model.ChatContentAssistantTypeGeneral
					aiMsg.ContentAssistant.GeneralAnswer = &model.ChatContentAssistantGeneralAnswerRes{
						View: model.ChatContentAssistantProjectInfoViewGeneralAnswer,
						GeneralAnswer: model.ChatContentAssistantGeneralAnswer{
							ID:      primitive.NewObjectID(),
							Content: chatAIAnalyticalResult.Content,
						},
					}
				} else {
					isGeneralFaq := false
					datas, view, err := s.fetchProjectDataByKey(ctx, chatAIAnalyticalResult.IntentKeys, chatAIAnalyticalResult.View)
					if err != nil {
						ctx.AddCustomLogField("ai_err", err)
						if chatAIAnalyticalResult.Content != "" {
							isGeneralFaq = true
						}
					}
					chatAIAnalyticalResult.View = view
					aiMsg.ContentAssistant.ProjectKeys = chatAIAnalyticalResult.IntentKeys

					switch chatAIAnalyticalResult.Intention {
					case model.ChatAIAnalyticalIntentionSearch:
						if len(chatAIAnalyticalResult.IntentKeys) == 0 {
							return errors.New("the intention is not clear, and AI cannot analyze it. Please provide a more detailed description")
						}
						part := view
						if view == "overview" {
							part = "some info"
						}
						aiMsg.ContentAssistant.Tips = fmt.Sprintf("Here is %s about %s:", part, chatAIAnalyticalResult.IntentKeys[0])
						if isGeneralFaq {
							aiMsg.ContentAssistant.Type = model.ChatContentAssistantTypeGeneral
							aiMsg.ContentAssistant.GeneralAnswer = &model.ChatContentAssistantGeneralAnswerRes{
								View: model.ChatContentAssistantProjectInfoViewGeneralAnswer,
								GeneralAnswer: model.ChatContentAssistantGeneralAnswer{
									ID:      primitive.NewObjectID(),
									Content: chatAIAnalyticalResult.Content,
								},
							}
						} else {
							aiMsg.ContentAssistant.Type = model.ChatContentAssistantTypeProjectInfo
							aiMsg.ContentAssistant.ProjectInfo = &model.ChatContentAssistantProjectInfoRes{
								View:    chatAIAnalyticalResult.View,
								Project: datas[0],
							}
						}
					case model.ChatAIAnalyticalIntentionCompare:
						if len(chatAIAnalyticalResult.IntentKeys) < 2 {
							return errors.New("Sorry, I didn't understand what you meant. Can you provide me with more information?")
						}
						if view == "overview" {
							aiMsg.ContentAssistant.Tips = fmt.Sprintf("Here is a brief comparison about %s and %s:", chatAIAnalyticalResult.IntentKeys[0], chatAIAnalyticalResult.IntentKeys[1])
						} else {
							aiMsg.ContentAssistant.Tips = fmt.Sprintf("Here is %s about comparison of %s and %s:", view, chatAIAnalyticalResult.IntentKeys[0], chatAIAnalyticalResult.IntentKeys[1])
						}
						aiMsg.ContentAssistant.Type = model.ChatContentAssistantTypeProjectCompare
						aiMsg.ContentAssistant.ProjectCompare = &model.ChatContentAssistantProjectCompareRes{
							View:     chatAIAnalyticalResult.View,
							Projects: datas,
						}
					default:
						aiMsg.ContentAssistant.Type = model.ChatContentAssistantTypeGeneral
						aiMsg.ContentAssistant.GeneralAnswer = &model.ChatContentAssistantGeneralAnswerRes{
							View: model.ChatContentAssistantProjectInfoViewGeneralAnswer,
							GeneralAnswer: model.ChatContentAssistantGeneralAnswer{
								ID:      primitive.NewObjectID(),
								Content: chatAIAnalyticalResult.Content,
							},
						}
					}
					chatAIAnalyticalResult.ProjectIDs = lo.Map(datas, func(item model.ChatContentAssistantProjectInfo, index int) primitive.ObjectID {
						return item.ID
					})
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
				}
				return nil
			}

			if fcRet != nil {
				if fcRet.FCType == model.FCSwap {
					chatAIAnalyticalResult.Intention = model.ChatAIAnalyticalIntentionSwap
					chatAIAnalyticalResult.IntentKeys = []string{fcRet.FCSwapResult.SwapInToken, fcRet.FCSwapResult.SwapOutToken}
					jsonStr, _ := json.Marshal(fcRet.FCSwapResult)
					chatAIAnalyticalResult.Content = string(jsonStr)
					chatAIAnalyticalResult.View = string(fcRet.FCType)
					chatAIAnalyticalResult.Fill = ""
					chatAIAnalyticalResult.ProjectIDs = []primitive.ObjectID{}

					aiMsg.ContentAssistant.Type = model.ChatAIAnalyticalIntentionSwap
					aiMsg.ContentAssistant.Fill = ""
					aiMsg.ContentAssistant.ProjectKeys = chatAIAnalyticalResult.IntentKeys
					aiMsg.ContentAssistant.Tips = "Sure, you can swap here. If it doesn't meet your requirements, you can directly operate in the panel or modify your prompt."
					swap := &model.ChatContentAssistantInfo{
						ID:             primitive.NewObjectID(),
						FuncCallingRet: *fcRet,
					}
					aiMsg.ContentAssistant.SwapInfo = &model.ChatContentAssistantSwapRes{
						View: chatAIAnalyticalResult.View,
						Swap: *swap,
					}
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
					return nil
				} else if fcRet.FCType == model.FCUniswap { // uniswap project
					chatAIAnalyticalResult.Intention = model.ChatAIAnalyticalIntentionUniswap
					chatAIAnalyticalResult.IntentKeys = []string{model.ChatAIAnalyticalIntentionUniswap}
					jsonStr, _ := json.Marshal(fcRet.FCUniswapResult)
					chatAIAnalyticalResult.Content = string(jsonStr)
					chatAIAnalyticalResult.View = string(fcRet.FCType)
					chatAIAnalyticalResult.Fill = ""
					chatAIAnalyticalResult.ProjectIDs = []primitive.ObjectID{}

					aiMsg.ContentAssistant.Type = model.ChatAIAnalyticalIntentionUniswap
					aiMsg.ContentAssistant.Fill = ""
					aiMsg.ContentAssistant.ProjectKeys = chatAIAnalyticalResult.IntentKeys
					aiMsg.ContentAssistant.Tips = "Sure, you can swap here. If it doesn't meet your requirements, you can directly operate in the panel or modify your prompt."
					uniswap := model.ChatContentAssistantInfo{
						ID:             primitive.NewObjectID(),
						FuncCallingRet: *fcRet,
					}
					aiMsg.ContentAssistant.Uniswap = &model.ChatContentAssistantUniswapRes{
						View:    model.ChatContentAssistantSwapViewUniswap,
						Uniswap: uniswap,
					}
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
					return nil
				} else if fcRet.FCType == model.FCDailyNewToken {
					chatAIAnalyticalResult.Intention = model.ChatAIDailyNewToken
					chatAIAnalyticalResult.IntentKeys = []string{model.ChatAIDailyNewToken}
					// data result
					jsonStr, _ := json.Marshal(fcRet.NewTokenResult)
					chatAIAnalyticalResult.Content = string(jsonStr)
					// 视图
					chatAIAnalyticalResult.View = string(fcRet.FCType)
					chatAIAnalyticalResult.Fill = ""
					chatAIAnalyticalResult.ProjectIDs = []primitive.ObjectID{}
					// build ai assistant message
					aiMsg.ContentAssistant.Type = model.ChatAIDailyNewToken
					aiMsg.ContentAssistant.Fill = ""
					aiMsg.ContentAssistant.ProjectKeys = chatAIAnalyticalResult.IntentKeys
					aiMsg.ContentAssistant.Tips = "here is daily new tokens"
					newToken := model.ChatContentAssistantInfo{
						ID:             primitive.NewObjectID(),
						FuncCallingRet: *fcRet,
					}
					// pump.fun new token
					aiMsg.ContentAssistant.DailyNewToken = &model.ChatContentAssistantNewTokenRes{
						View:     model.ChatContentAssistantNewTokenView,
						NewToken: newToken,
					}
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
					return nil
				} else if fcRet.FCType == model.FCTokenLaunchedTimeDistribution {
					chatAIAnalyticalResult.Intention = model.ChatAITokenLaunchedTimeDistribution
					chatAIAnalyticalResult.IntentKeys = []string{model.ChatAITokenLaunchedTimeDistribution}
					// data
					jsonStr, _ := json.Marshal(fcRet.TokenLTD)
					chatAIAnalyticalResult.Content = string(jsonStr)
					chatAIAnalyticalResult.View = string(fcRet.FCType)
					chatAIAnalyticalResult.Fill = ""
					chatAIAnalyticalResult.ProjectIDs = []primitive.ObjectID{}

					aiMsg.ContentAssistant.Type = model.ChatAITokenLaunchedTimeDistribution
					aiMsg.ContentAssistant.Fill = ""
					aiMsg.ContentAssistant.ProjectKeys = chatAIAnalyticalResult.IntentKeys
					aiMsg.ContentAssistant.Tips = "here is token launched time distribution "
					tokenltd := model.ChatContentAssistantInfo{
						ID:             primitive.NewObjectID(),
						FuncCallingRet: *fcRet,
					}
					aiMsg.ContentAssistant.TokenLaunchedTime = &model.ChatContentAssistantTokenLtDistRes{
						View:                     model.ChatContentAssistantTokenLTView,
						LaunchedTimeDistribution: tokenltd,
					}
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
					return nil
				} else if fcRet.FCType == model.FCDailyTokenSwapCount {
					chatAIAnalyticalResult.Intention = model.ChatAIDailyTokenSwapCount
					chatAIAnalyticalResult.IntentKeys = []string{model.ChatAIDailyTokenSwapCount}
					// data
					jsonStr, _ := json.Marshal(fcRet.TokenSwapCount)
					chatAIAnalyticalResult.Content = string(jsonStr)
					chatAIAnalyticalResult.View = string(fcRet.FCType)
					chatAIAnalyticalResult.Fill = ""
					chatAIAnalyticalResult.ProjectIDs = []primitive.ObjectID{}

					aiMsg.ContentAssistant.Type = model.ChatAIDailyTokenSwapCount
					aiMsg.ContentAssistant.Fill = ""
					aiMsg.ContentAssistant.ProjectKeys = chatAIAnalyticalResult.IntentKeys
					aiMsg.ContentAssistant.Tips = "Here is daily token swap count"
					swap := model.ChatContentAssistantInfo{
						ID:             primitive.NewObjectID(),
						FuncCallingRet: *fcRet,
					}
					aiMsg.ContentAssistant.TokenSwapCount = &model.ChatContentAssistantTokenSwapCountRes{
						View:           model.ChatContentAssistantTokenSwapCountView,
						TokenSwapCount: swap,
					}
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
					return nil
				} else if fcRet.FCType == model.FCTopTrader {
					chatAIAnalyticalResult.Intention = model.ChatAITopTrader
					chatAIAnalyticalResult.IntentKeys = []string{model.ChatAITopTrader}
					// data
					jsonStr, _ := json.Marshal(fcRet.TopTrader)
					chatAIAnalyticalResult.Content = string(jsonStr)
					chatAIAnalyticalResult.View = string(fcRet.FCType)
					chatAIAnalyticalResult.Fill = ""
					chatAIAnalyticalResult.ProjectIDs = []primitive.ObjectID{}

					aiMsg.ContentAssistant.Type = model.ChatAITopTrader
					aiMsg.ContentAssistant.Fill = ""
					aiMsg.ContentAssistant.ProjectKeys = chatAIAnalyticalResult.IntentKeys
					aiMsg.ContentAssistant.Tips = "Here is top trader"
					tt := model.ChatContentAssistantInfo{
						ID:             primitive.NewObjectID(),
						FuncCallingRet: *fcRet,
					}
					aiMsg.ContentAssistant.TopTrader = &model.ChatContentAssistantTopTraderRes{
						View:      model.ChatContentAssistantTopTrader,
						TopTrader: tt,
					}
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
					return nil
				} else if fcRet.FCType == model.FCTraderOverview {
					chatAIAnalyticalResult.Intention = model.ChatAITraderOverview
					chatAIAnalyticalResult.IntentKeys = []string{model.ChatAITraderOverview}
					// data
					jsonStr, _ := json.Marshal(fcRet.TraderOverview)
					chatAIAnalyticalResult.Content = string(jsonStr)
					chatAIAnalyticalResult.View = string(fcRet.FCType)
					chatAIAnalyticalResult.Fill = ""
					chatAIAnalyticalResult.ProjectIDs = []primitive.ObjectID{}

					aiMsg.ContentAssistant.Type = model.ChatAITraderOverview
					aiMsg.ContentAssistant.Fill = ""
					aiMsg.ContentAssistant.ProjectKeys = chatAIAnalyticalResult.IntentKeys
					aiMsg.ContentAssistant.Tips = "Here is trader overview of the past 7 days"
					to := model.ChatContentAssistantInfo{
						ID:             primitive.NewObjectID(),
						FuncCallingRet: *fcRet,
					}
					aiMsg.ContentAssistant.TraderOverview = &model.ChatContentAssistantTraderOverviewRes{
						View:           model.ChatContentAssistantTraderOverview,
						TraderOverview: to,
					}
					aiMsg.ContentAssistant.Fill = chatAIAnalyticalResult.Fill
					return nil
				}
			}

			return nil
		}()
		if err != nil {
			aiMsg.Role = model.ChatMsgRoleSystem
			aiMsg.ContentSystem = &model.ChatContentSystem{
				Content: err.Error(),
			}
		}
	case model.ChatMsgRoleUserBuiltin:
		err := func() error {
			relatedProjectInfo, err := s.chatDao.ChatHistoryQuery(ctx, w.ID, uint64(req.RelatedMsgIndex))
			if err != nil {
				return errors.Wrap(err, "internal error: failed to get last ai intention")
			}
			if relatedProjectInfo.Msg.Role != model.ChatMsgRoleAssistant {
				return errors.New("internal error: failed to get last ai intention")
			}

			userMsg.ContentUserBuiltin = &model.ChatContentUserBuiltin{
				Content: req.Msg,
			}
			aiMsg.Role = model.ChatMsgRoleAssistant
			aiMsg.ContentAssistant = &model.ChatContentAssistant{
				ProjectKeys: relatedProjectInfo.Msg.ContentAssistant.ProjectKeys,
			}

			switch relatedProjectInfo.Msg.ContentAssistant.Type {
			case model.ChatContentAssistantTypeFill:
				return errors.New("internal error: failed to get last ai intention")
			case model.ChatContentAssistantTypeProjectInfo:
				datas, view, err := s.fetchProjectData(ctx, []string{relatedProjectInfo.Msg.ContentAssistant.ProjectInfo.Project.ID.Hex()}, req.Msg)
				if err != nil {
					return errors.New(projectNotFoundErrMsg)
				}
				part := view
				if view == "overview" {
					part = "some info"
				}
				aiMsg.ContentAssistant.Tips = fmt.Sprintf("Here is %s about %s:", part, relatedProjectInfo.Msg.ContentAssistant.ProjectKeys[0])
				aiMsg.ContentAssistant.Type = model.ChatContentAssistantTypeProjectInfo
				aiMsg.ContentAssistant.ProjectInfo = &model.ChatContentAssistantProjectInfoRes{
					View:    view,
					Project: datas[0],
				}
			case model.ChatContentAssistantTypeProjectCompare:
				datas, view, err := s.fetchProjectData(ctx, []string{
					relatedProjectInfo.Msg.ContentAssistant.ProjectCompare.Projects[0].ID.Hex(),
					relatedProjectInfo.Msg.ContentAssistant.ProjectCompare.Projects[1].ID.Hex(),
				}, req.Msg)
				if err != nil {
					return errors.New(projectNotFoundErrMsg)
				}
				if view == "overview" {
					aiMsg.ContentAssistant.Tips = fmt.Sprintf("Here is a brief comparison about %s and %s:", relatedProjectInfo.Msg.ContentAssistant.ProjectKeys[0], relatedProjectInfo.Msg.ContentAssistant.ProjectKeys[1])
				} else {
					aiMsg.ContentAssistant.Tips = fmt.Sprintf("Here is %s about comparison of %s and %s:", view, relatedProjectInfo.Msg.ContentAssistant.ProjectKeys[0], relatedProjectInfo.Msg.ContentAssistant.ProjectKeys[1])
				}
				aiMsg.ContentAssistant.Type = model.ChatContentAssistantTypeProjectCompare
				aiMsg.ContentAssistant.ProjectCompare = &model.ChatContentAssistantProjectCompareRes{
					View:     view,
					Projects: datas,
				}
			}

			return nil
		}()
		if err != nil {
			aiMsg.Role = model.ChatMsgRoleSystem
			aiMsg.ContentSystem = &model.ChatContentSystem{
				Content: err.Error(),
			}
		}
	default:
		aiMsg.Role = model.ChatMsgRoleSystem
		aiMsg.ContentSystem = &model.ChatContentSystem{
			Content: "internal error, unsupported msg type: " + req.Type,
		}
	}

	if err := s.chatDao.ChatHistoryAdd(ctx, &model.ChatHistory{
		WindowID:  w.ID,
		Index:     w.MsgNum,
		Msg:       *userMsg,
		ProjectId: pjId,
	}); err != nil {
		return nil, err
	}
	if err := s.chatDao.ChatHistoryAdd(ctx, &model.ChatHistory{
		WindowID:  w.ID,
		Index:     w.MsgNum + 1,
		Msg:       *aiMsg,
		ProjectId: pjId,
	}); err != nil {
		return nil, err
	}
	w.MsgNum += 2
	if err := s.chatDao.ChatWindowUpdate(ctx, w); err != nil {
		return nil, err
	}

	var content any
	if aiMsg.Role == model.ChatMsgRoleSystem {
		content = aiMsg.ContentSystem
	} else if aiMsg.Role == model.ChatMsgRoleAssistant {
		content = aiMsg.ContentAssistant
	}
	return &entity.ChatCompletionsRes{
		Msg: &entity.ChatMsg{
			Timestamp: aiMsg.Timestamp,
			Role:      aiMsg.Role,
			Content:   content,
			ProjectId: pjId,
		},
	}, nil
}

func (s *ChatService) Update(ctx *reqctx.ReqCtx, req *entity.ChatUpdateReq) (*entity.ChatUpdateRes, error) {
	w, err := s.chatDao.ChatWindowQuery(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if w.Creator.Hex() != ctx.Caller {
		return nil, errcode.ErrAccountPermission
	}
	w.TitleIsSet = true
	w.Title = req.Title
	if err := s.chatDao.ChatWindowUpdate(ctx, w); err != nil {
		return nil, err
	}
	return &entity.ChatUpdateRes{}, nil
}

func (s *ChatService) Delete(ctx *reqctx.ReqCtx, req *entity.ChatDeleteReq) (*entity.ChatDeleteRes, error) {
	w, err := s.chatDao.ChatWindowQuery(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if w.Creator.Hex() != ctx.Caller {
		return nil, errcode.ErrAccountPermission
	}
	if err := s.chatDao.ChatWindowDelete(ctx, req.ID); err != nil {
		return nil, err
	}
	return &entity.ChatDeleteRes{}, nil
}

func (s *ChatService) DeleteAll(ctx *reqctx.ReqCtx, req *entity.ChatDeleteAllReq) (*entity.ChatDeleteAllRes, error) {
	objID, err := primitive.ObjectIDFromHex(ctx.Caller)
	if err != nil {
		return nil, err
	}

	res, _, err := s.chatDao.ChatWindowList(ctx, 0, 0, bson.M{
		"is_deleted": false,
		"creator":    objID,
	}, nil)
	if err != nil {
		return nil, err
	}

	for _, w := range res {
		if err := s.chatDao.ChatWindowDelete(ctx, w.ID.Hex()); err != nil {
			return nil, err
		}
	}

	return &entity.ChatDeleteAllRes{}, nil
}

func (s *ChatService) InsertAgents(ctx *reqctx.ReqCtx, userId string) error {
	pin, err := s.userPluginDao.QueryByUserIdAndPj(ctx, userId, s.baseComponent.Config.AIBackend.UniProjectId)
	if err != nil {
		return err
	}
	if pin == nil {
		uni := &model.UserPlugin{
			UserId:      userId,
			ProjectId:   s.baseComponent.Config.AIBackend.UniProjectId,
			ProjectName: "Uniswap",
			PinStatus:   0,
		}
		bmuni, err := model.NewBaseModel(userId)
		if err != nil {
			return err
		}
		uni.BaseModel = bmuni
		err = s.userPluginDao.InsertUserPin(ctx, uni)
		if err != nil {
			return err
		}
	}

	pinwyt, err := s.userPluginDao.QueryByUserIdAndPj(ctx, userId, s.baseComponent.Config.AIBackend.ProjectId)
	if pinwyt == nil {
		// wyt
		wyt := &model.UserPlugin{
			UserId:      userId,
			ProjectId:   s.baseComponent.Config.AIBackend.ProjectId,
			ProjectName: "WYT Agent",
			PinStatus:   1,
		}
		wytbm, err := model.NewBaseModel(userId)
		if err != nil {
			return err
		}
		wyt.BaseModel = wytbm
		err = s.userPluginDao.InsertUserPin(ctx, wyt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ChatService) AgentList(ctx *reqctx.ReqCtx, req *entity.AgentListReq) (*entity.AgentListRes, error) {
	objID, err := primitive.ObjectIDFromHex(ctx.Caller)
	if err != nil {
		return nil, err
	}
	rets := make([]*model.UserPluginDto, 0)
	pinList, total, err := s.userPluginDao.UserPluginPinList(ctx, 0, 0, bson.M{
		"is_deleted": false,
		"user_id":    objID.Hex(),
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(pinList) == 0 || total == 0 {
		err := s.InsertAgents(ctx, ctx.Caller)
		if err != nil {
			s.baseComponent.Logger.Info("insert user associated agent error", err)
			return nil, err
		}
		rets = append(rets, &model.UserPluginDto{
			UserId:      ctx.Caller,
			ProjectId:   s.baseComponent.Config.AIBackend.ProjectId,
			ProjectName: "WYT Agent",
			PinStatus:   1,
		}, &model.UserPluginDto{
			UserId:      ctx.Caller,
			ProjectId:   s.baseComponent.Config.AIBackend.UniProjectId,
			ProjectName: "Uniswap",
			PinStatus:   0,
		})
	} else {
		for _, p := range pinList {
			dto := &model.UserPluginDto{
				UserId:      p.UserId,
				ProjectId:   p.ProjectId,
				ProjectName: p.ProjectName,
				PinStatus:   p.PinStatus,
				PinTime:     time.Time(p.UpdateTime).Unix(),
			}
			rets = append(rets, dto)
		}
	}
	resp := &entity.AgentListRes{
		Total: total,
		List:  rets,
	}
	return resp, nil
}

func (s *ChatService) PinAgent(ctx *reqctx.ReqCtx, req *entity.AgentPinReq) error {
	objID, err := primitive.ObjectIDFromHex(ctx.Caller)
	if err != nil {
		return err
	}
	up, err := s.userPluginDao.QueryByUserIdAndPj(ctx, objID.Hex(), req.ProjectId)
	if err != nil {
		if err == errcode.ErrUserPinNotExist {
			up = nil
		} else {
			return err
		}
	}
	if up == nil {
		var pjName string
		if req.ProjectId == s.baseComponent.Config.AIBackend.ProjectId {
			pjName = "WYT Agent"
		} else if req.ProjectId == s.baseComponent.Config.AIBackend.UniProjectId {
			pjName = "Uniswap"
		}
		up = &model.UserPlugin{
			UserId:      objID.Hex(),
			ProjectId:   req.ProjectId,
			ProjectName: pjName,
			PinStatus:   1,
		}
		bm, err := model.NewBaseModel(ctx.Caller)
		if err != nil {
			return err
		}
		up.BaseModel = bm
		err = s.userPluginDao.InsertUserPin(ctx, up)
		if err != nil {
			return err
		}
		return nil
	} else {
		up.PinStatus = 1
	}

	err = s.userPluginDao.UpdateUserPin(ctx, up)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatService) UnpinAgent(ctx *reqctx.ReqCtx, req *entity.AgentUnPinReq) error {
	objID, err := primitive.ObjectIDFromHex(ctx.Caller)
	if err != nil {
		return err
	}
	up, err := s.userPluginDao.QueryByUserIdAndPj(ctx, objID.Hex(), req.ProjectId)
	if err != nil {
		return err
	}
	if up != nil {
		up.PinStatus = 0
	}
	err = s.userPluginDao.UpdateUserPin(ctx, up)
	if err != nil {
		return err
	}
	return nil
}

// func patternMatchTransfer(msg string) *model.ChatAIAnalyticalResult {
// 	res := model.ChatAIAnalyticalResult{}
// 	res.Intention = model.ChatAIAnalyticalIntentionDex

// 	if regexp.MustCompile(chatSearchDEx).MatchString(msg) {
// 		res.IntentKeys = []string{msg}
// 		res.View = "overview"
// 		// contentMap := make(map[string]interface{}, 0)
// 		// contentMap
// 		// res.Content =
// 		return &res
// 	}

// 	if regexp.MustCompile(chatSearchTokenRegEx0).MatchString(msg) {
// 		res.IntentKeys = []string{msg}
// 		res.View = "overview"
// 		return &res
// 	}

// 	var match []string
// 	var pattern string
// 	var view string

// 	switch {
// 	case regexp.MustCompile(chatSearchTokenRegEx1).MatchString(msg):
// 		pattern = chatSearchTokenRegEx1
// 		view = "overview"
// 	case regexp.MustCompile(chatSearchTokenRegEx2).MatchString(msg):
// 		pattern = chatSearchTokenRegEx2
// 		view = "overview"
// 	case regexp.MustCompile(chatSearchTokenPriceRegEx1).MatchString(msg):
// 		pattern = chatSearchTokenPriceRegEx1
// 		view = "overview"
// 	case regexp.MustCompile(chatSearchTokenPriceRegEx2).MatchString(msg):
// 		pattern = chatSearchTokenPriceRegEx2
// 		view = "overview"
// 	case regexp.MustCompile(chatSearchTokenBaseRegEx).MatchString(msg):
// 		pattern = chatSearchTokenBaseRegEx
// 		view = "overview"
// 	case regexp.MustCompile(chatSearchTokenTeamRegEx).MatchString(msg):
// 		pattern = chatSearchTokenTeamRegEx
// 		view = "team"
// 	case regexp.MustCompile(chatSearchTokenFundRegEx1).MatchString(msg):
// 		pattern = chatSearchTokenFundRegEx1
// 		view = "funding"
// 	case regexp.MustCompile(chatSearchTokenFundRegEx2).MatchString(msg):
// 		pattern = chatSearchTokenFundRegEx2
// 		view = "funding"
// 	case regexp.MustCompile(chatSearchTokenExchangesRegEx).MatchString(msg):
// 		pattern = chatSearchTokenExchangesRegEx
// 		view = "exchanges"
// 	case regexp.MustCompile(chatSearchTokenEconomicRegEx1).MatchString(msg):
// 		pattern = chatSearchTokenEconomicRegEx1
// 		view = "tokenomics"
// 	case regexp.MustCompile(chatSearchTokenEconomicRegEx2).MatchString(msg):
// 		pattern = chatSearchTokenEconomicRegEx2
// 		view = "tokenomics"
// 	case regexp.MustCompile(chatSearchTokenEconomicRegEx3).MatchString(msg):
// 		pattern = chatSearchTokenEconomicRegEx3
// 		view = "tokenomics"
// 	case regexp.MustCompile(chatCompareTokenRegEx1).MatchString(msg):
// 		res.Intention = model.ChatAIAnalyticalIntentionCompare
// 		pattern = chatCompareTokenRegEx1
// 		view = "overview"
// 	case regexp.MustCompile(chatCompareTokenRegEx2).MatchString(msg):
// 		res.Intention = model.ChatAIAnalyticalIntentionCompare
// 		pattern = chatCompareTokenRegEx2
// 		view = "overview"
// 	case regexp.MustCompile(chatCompareTokenRegEx3).MatchString(msg):
// 		res.Intention = model.ChatAIAnalyticalIntentionCompare
// 		pattern = chatCompareTokenRegEx3
// 		view = "overview"
// 	case regexp.MustCompile(chatCompareTokenRegEx4).MatchString(msg):
// 		res.Intention = model.ChatAIAnalyticalIntentionCompare
// 		pattern = chatCompareTokenRegEx4
// 		view = "overview"
// 	case regexp.MustCompile(chatCompareTokenRegEx5).MatchString(msg):
// 		res.Intention = model.ChatAIAnalyticalIntentionCompare
// 		pattern = chatCompareTokenRegEx5
// 		view = "overview"
// 	default:
// 		return nil
// 	}

// 	match = regexp.MustCompile(pattern).FindStringSubmatch(msg)
// 	if len(match) <= 1 {
// 		return nil
// 	}

// 	res.IntentKeys = match[1:]
// 	res.View = view

// 	return &res
// }
