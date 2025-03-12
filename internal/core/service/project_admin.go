package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *ProjectService) AdminAdd(ctx *reqctx.ReqCtx, req *entity.ProjectAddReq) (*entity.ProjectAddRes, error) {
	info, err := s.infoEntityToModel(ctx, &req.ProjectInput)
	if err != nil {
		return nil, err
	}

	caller, err := primitive.ObjectIDFromHex(ctx.Caller)
	if err != nil {
		return nil, err
	}
	info.ProjectInternalInfo = model.ProjectInternalInfo{
		Status:    model.ProjectStatusIndexed,
		Completer: caller,
	}

	// check chain,track,tag,team-impressions,top-investor ids, top-project ids
	if err := s.miscDao.ChainBatchCheck(ctx, req.Basic.Chains); err != nil {
		return nil, err
	}
	if err := s.miscDao.TrackBatchCheck(ctx, req.Basic.Tracks); err != nil {
		return nil, err
	}
	if err := s.miscDao.TagBatchCheck(ctx, req.Basic.Tags); err != nil {
		return nil, err
	}
	if err := s.miscDao.TeamImpressionBatchCheck(ctx, req.Team.Impressions); err != nil {
		return nil, err
	}
	if err := s.miscDao.InvestorBatchCheck(ctx, req.Funding.TopInvestors); err != nil {
		return nil, err
	}
	if err := s.projectDao.BatchCheck(ctx, false, req.Ecosystem.TopProjects); err != nil {
		return nil, err
	}

	if err := s.projectDao.Add(ctx, false, info); err != nil {
		return nil, err
	}
	return &entity.ProjectAddRes{
		ID: info.ID,
	}, nil
}

func (s *ProjectService) AdminUpdate(ctx *reqctx.ReqCtx, req *entity.ProjectUpdateReq) (*entity.ProjectUpdateRes, error) {
	info, err := s.projectDao.Query(ctx, false, req.ID)
	if err != nil {
		return nil, err
	}

	if info.Basic.Name != req.Basic.Name {
		exist, err := s.projectDao.ExistsByProjectName(ctx, false, req.Basic.Name)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, errcode.ErrProjectAlreadyExists
		}
	}

	newInfo, err := s.infoEntityToModel(ctx, &req.ProjectInput)
	if err != nil {
		return nil, err
	}
	info.Basic = newInfo.Basic
	info.RelatedLinks = newInfo.RelatedLinks
	info.Team = newInfo.Team
	info.Funding = newInfo.Funding
	info.Tokenomics = newInfo.Tokenomics
	info.Ecosystem = newInfo.Ecosystem
	info.Profitability = newInfo.Profitability

	// check chain,track,tag,team-impressions,top-investor ids, top-project ids
	if err := s.miscDao.ChainBatchCheck(ctx, req.Basic.Chains); err != nil {
		return nil, err
	}
	if err := s.miscDao.TrackBatchCheck(ctx, req.Basic.Tracks); err != nil {
		return nil, err
	}
	if err := s.miscDao.TagBatchCheck(ctx, req.Basic.Tags); err != nil {
		return nil, err
	}
	if err := s.miscDao.TeamImpressionBatchCheck(ctx, req.Team.Impressions); err != nil {
		return nil, err
	}
	if err := s.miscDao.InvestorBatchCheck(ctx, req.Funding.TopInvestors); err != nil {
		return nil, err
	}
	if err := s.projectDao.BatchCheck(ctx, false, req.Ecosystem.TopProjects); err != nil {
		return nil, err
	}

	if info.Status == model.ProjectStatusPublished {
		info.Status = model.ProjectStatusReedited
	}
	info.CompletionStatus = req.CompletionStatus
	if err := s.projectDao.Update(ctx, false, info); err != nil {
		return nil, err
	}
	return &entity.ProjectUpdateRes{}, nil
}

func (s *ProjectService) AdminSimpleUpdate(ctx *reqctx.ReqCtx, req *entity.ProjectSimpleUpdateReq) (*entity.ProjectSimpleUpdateRes, error) {
	info, err := s.projectDao.Query(ctx, false, req.ID)
	if err != nil {
		return nil, err
	}
	if !ctx.IsZHLang {
		info.Basic.Description = req.Description
	} else {
		info.Basic.DescriptionZH = req.Description
	}
	info.Basic.Name = req.Name
	info.Basic.LogoURL = req.LogoURL
	needAdd := true
	for i, link := range info.RelatedLinks {
		if link.Type == officialWebsiteLinkType {
			needAdd = false
			info.RelatedLinks[i].Link = req.OfficialWebsite
			break
		}
	}
	if needAdd {
		info.RelatedLinks = append(info.RelatedLinks, model.LinkInfo{
			Type: officialWebsiteLinkType,
			Link: req.OfficialWebsite,
		})
	}

	if info.Status == model.ProjectStatusPublished {
		info.Status = model.ProjectStatusReedited
	}
	if err := s.projectDao.Update(ctx, false, info); err != nil {
		return nil, err
	}
	return &entity.ProjectSimpleUpdateRes{}, nil
}

func (s *ProjectService) AdminPublish(ctx *reqctx.ReqCtx, req *entity.ProjectPublishReq) (*entity.ProjectPublishRes, error) {
	info, err := s.projectDao.Query(ctx, false, req.ID)
	if err != nil {
		return nil, err
	}

	info.Status = model.ProjectStatusPublished

	if err := s.projectDao.Upsert(ctx, true, info); err != nil {
		return nil, err
	}

	if err := s.projectDao.Update(ctx, false, info); err != nil {
		return nil, err
	}

	s.baseComponent.SafeGo(func() {
		if err := s.marketDatasource.AddProject(info.ID.Hex(), info.Tokenomics.TokenSymbol, info.Tokenomics.CirculatingSupply); err != nil {
			s.baseComponent.Logger.WithFields(logrus.Fields{
				"err":    err,
				"symbol": info.Tokenomics.TokenSymbol,
			}).Error("Failed to update subscription for new project")
		}
	})
	return &entity.ProjectPublishRes{}, nil
}

func (s *ProjectService) AdminInfo(ctx *reqctx.ReqCtx, req *entity.ProjectAdminInfoReq) (*entity.ProjectAdminInfoRes, error) {
	info, err := s.projectDao.Query(ctx, false, req.ID)
	if err != nil {
		return nil, err
	}

	infoEntity, err := s.infoModelToEntity(ctx, false, info)
	if err != nil {
		return nil, err
	}

	return &entity.ProjectAdminInfoRes{
		ProjectOutput: *infoEntity,
	}, nil
}

func (s *ProjectService) AdminSimpleInfo(ctx *reqctx.ReqCtx, req *entity.ProjectAdminSimpleInfoReq) (*entity.ProjectAdminSimpleInfoRes, error) {
	var res entity.ProjectSimpleOutputQueryWrapper
	err := s.projectDao.CustomQuery(ctx, false, req.ID, &res)
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
	return &entity.ProjectAdminSimpleInfoRes{
		ProjectSimpleOutput: e,
	}, nil
}

func (s *ProjectService) AdminList(ctx *reqctx.ReqCtx, req *entity.ProjectAdminListReq) (*entity.ProjectAdminListRes, error) {
	var list []*entity.ProjectAdminListElement

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
	if req.Status != entity.ProjectStatusFilterAll {
		switch req.Status {
		case entity.ProjectStatusFilterPublished:
			conditions = append(conditions, bson.M{
				"$or": bson.A{
					bson.M{"status": model.ProjectStatusPublished},
					bson.M{"status": model.ProjectStatusReedited},
				},
			})
		case entity.ProjectStatusFilterUnpublished:
			conditions = append(conditions, bson.M{
				"status": model.ProjectStatusIndexed,
			})
		case entity.ProjectStatusFilterCompleted:
			conditions = append(conditions, bson.M{
				"completion_status": model.ProjectCompletionStatusComplete,
			})
		case entity.ProjectStatusFilterIncompleted:
			conditions = append(conditions, bson.M{
				"completion_status": model.ProjectCompletionStatusIncomplete,
			})
		case entity.ProjectStatusFilterCoreCompleted:
			conditions = append(conditions, bson.M{
				"completion_status": model.ProjectCompletionStatusCoreDataComplete,
			})
		}
	}

	sortFields := map[string]bool{}
	if req.SortField != "" {
		sortFields[req.SortField] = req.IsAsc
	}
	total, err := s.projectDao.CustomList(ctx, false, req.Page, req.Size, bson.M{"$and": conditions}, sortFields, &list)
	if err != nil {
		return nil, err
	}
	fmt.Println(len(list))
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
	}

	return &entity.ProjectAdminListRes{
		List:  list,
		Total: total,
	}, nil
}

func (s *ProjectService) AdminDelete(ctx *reqctx.ReqCtx, req *entity.ProjectDeleteReq) (*entity.ProjectDeleteRes, error) {
	info, err := s.projectDao.Query(ctx, false, req.ID)
	if err != nil {
		return nil, err
	}
	if info.Status == model.ProjectStatusPublished || info.Status == model.ProjectStatusReedited {
		return nil, errcode.ErrProjectPublished
	}
	if err := s.projectDao.Delete(ctx, false, req.ID); err != nil {
		return nil, err
	}
	return &entity.ProjectDeleteRes{}, nil
}

func (s *ProjectService) AdminCalculateDerivedData(ctx *reqctx.ReqCtx, req *entity.ProjectCalculateDerivedDataReq) (*entity.ProjectCalculateDerivedDataRes, error) {
	ctx.Ctx = context.Background()
	s.baseComponent.SafeGo(func() {
		res, _, err := s.projectDao.List(ctx, req.IsView, 0, 0, nil, nil)
		if err != nil {
			ctx.Logger.WithField("err", err).Warn("Failed to load all project")
			return
		}

		parallelNum := 10
		if len(res) < parallelNum {
			parallelNum = len(res)
		}
		chunkSize := len(res) / parallelNum
		remainder := len(res) % parallelNum
		wg := &sync.WaitGroup{}
		for i := 0; i < parallelNum; i++ {
			chunkEnd := chunkSize
			if i < remainder {
				chunkEnd++
			}
			chunk := res[0:chunkEnd]
			res = res[chunkEnd:]
			wg.Add(1)
			index := i
			s.baseComponent.SafeGo(func() {
				defer wg.Done()
				for j, e := range chunk {
					err = s.projectDao.UpdateInternalData(ctx, req.IsView, e)
					if err != nil {
						ctx.Logger.WithField("err", err).Warnf("Failed to calculate project derived data for %s", e.Basic.Name)
					}
					if j%100 == 0 && j > 0 {
						ctx.Logger.Infof("[%d] Calculate project derived data progress: %v", index, j)
					}
				}
			})
		}
		wg.Wait()
		ctx.Logger.Infof("Calculate project derived data finished")
	})

	return &entity.ProjectCalculateDerivedDataRes{}, nil
}
