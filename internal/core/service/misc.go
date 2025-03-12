package service

import (
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

type MiscService struct {
	baseComponent *base.Component
	miscDao       *dao.MiscDao
}

func NewMiscService(baseComponent *base.Component, miscDao *dao.MiscDao) *MiscService {
	return &MiscService{
		baseComponent: baseComponent,
		miscDao:       miscDao,
	}
}

func (s *MiscService) ChainList(ctx *reqctx.ReqCtx, req *entity.ChainListReq) (*entity.ChainListRes, error) {
	res, total, err := s.miscDao.ChainList(ctx, req.Page, req.Size, bson.M{"is_deleted": false}, map[string]bool{"create_time": false})
	if err != nil {
		return nil, err
	}
	lo.ForEach(res, func(m *model.Chain, index int) {
		m.Translate(ctx.IsZHLang)
	})
	return &entity.ChainListRes{
		List:  res,
		Total: total,
	}, nil
}

func (s *MiscService) TrackList(ctx *reqctx.ReqCtx, req *entity.TrackListReq) (*entity.TrackListRes, error) {
	if req.IsHot {
		req.Page = 1
		req.Size = s.baseComponent.Config.App.HotTrackNum
	}
	sort := map[string]bool{"score": false}

	res, total, err := s.miscDao.TrackList(ctx, req.Page, req.Size, bson.M{"is_deleted": false}, sort)
	if err != nil {
		return nil, err
	}
	lo.ForEach(res, func(m *model.Track, index int) {
		m.Translate(ctx.IsZHLang)
	})
	return &entity.TrackListRes{
		List:  res,
		Total: total,
	}, nil
}

func (s *MiscService) TagList(ctx *reqctx.ReqCtx, req *entity.TagListReq) (*entity.TagListRes, error) {
	if req.IsHot {
		req.Page = 1
		req.Size = s.baseComponent.Config.App.HotTagNum
	}
	sort := map[string]bool{"score": false}
	var conditions bson.A
	conditions = append(conditions, bson.M{
		"is_deleted": false,
	})
	if req.Query != "" {
		conditions = append(conditions, bson.M{
			"name": bson.M{"$regex": req.Query, "$options": "i"},
		})
	}
	res, total, err := s.miscDao.TagList(ctx, req.Page, req.Size, bson.M{"$and": conditions}, sort)
	if err != nil {
		return nil, err
	}
	lo.ForEach(res, func(m *model.Tag, index int) {
		m.Translate(ctx.IsZHLang)
	})
	return &entity.TagListRes{
		List:  res,
		Total: total,
	}, nil
}

func (s *MiscService) TagAdd(ctx *reqctx.ReqCtx, req *entity.TagAddReq) (*entity.TagAddRes, error) {
	tag := &model.Tag{}
	if !ctx.IsZHLang {
		tag.Name = req.Name
		tag.Description = req.Description
	} else {
		tag.NameZH = req.Name
		tag.DescriptionZH = req.Description
	}
	if ctx.CallerRole != model.UserRoleMember {
		tag.Score = req.Score
	}

	if err := s.miscDao.TagAdd(ctx, tag); err != nil {
		return nil, err
	}
	return &entity.TagAddRes{
		ID: tag.ID,
	}, nil
}

func (s *MiscService) TeamImpressionList(ctx *reqctx.ReqCtx, req *entity.TeamImpressionListReq) (*entity.TeamImpressionListRes, error) {
	res, total, err := s.miscDao.TeamImpressionList(ctx, req.Page, req.Size, bson.M{"is_deleted": false}, map[string]bool{"create_time": false})
	if err != nil {
		return nil, err
	}
	lo.ForEach(res, func(m *model.TeamImpression, index int) {
		m.Translate(ctx.IsZHLang)
	})
	return &entity.TeamImpressionListRes{
		List:  res,
		Total: total,
	}, nil
}

func (s *MiscService) InvestorAdd(ctx *reqctx.ReqCtx, req *entity.InvestorAddReq) (*entity.InvestorAddRes, error) {
	investor := &model.Investor{
		Name:             req.Name,
		AvatarURL:        req.AvatarURL,
		Subject:          req.Subject,
		Type:             req.Type,
		SocialMediaLinks: req.SocialMediaLinks,
	}
	if !ctx.IsZHLang {
		investor.Description = req.Description
	} else {
		investor.DescriptionZH = req.Description
	}

	if err := s.miscDao.InvestorAdd(ctx, investor); err != nil {
		return nil, err
	}
	return &entity.InvestorAddRes{
		ID: investor.ID,
	}, nil
}

func (s *MiscService) InvestorUpdate(ctx *reqctx.ReqCtx, req *entity.InvestorUpdateReq) (*entity.InvestorUpdateRes, error) {
	investor, err := s.miscDao.InvestorQuery(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	investor.Name = req.Name
	investor.AvatarURL = req.AvatarURL
	investor.Subject = req.Subject
	investor.Type = req.Type
	investor.SocialMediaLinks = req.SocialMediaLinks
	if !ctx.IsZHLang {
		investor.Description = req.Description
	} else {
		investor.DescriptionZH = req.Description
	}

	if err := s.miscDao.InvestorUpdate(ctx, investor); err != nil {
		return nil, err
	}
	return &entity.InvestorUpdateRes{}, nil
}

func (s *MiscService) InvestorList(ctx *reqctx.ReqCtx, req *entity.InvestorListReq) (*entity.InvestorListRes, error) {
	var conditions bson.A
	conditions = append(conditions, bson.M{
		"is_deleted": false,
	})
	if req.Query != "" {
		conditions = append(conditions, bson.M{
			"name": bson.M{"$regex": req.Query, "$options": "i"},
		})
	}
	res, total, err := s.miscDao.InvestorList(ctx, req.Page, req.Size, bson.M{"$and": conditions}, map[string]bool{"is_top": false})
	if err != nil {
		return nil, err
	}
	lo.ForEach(res, func(m *model.Investor, index int) {
		m.Translate(ctx.IsZHLang)
	})
	return &entity.InvestorListRes{
		List:  res,
		Total: total,
	}, nil
}
