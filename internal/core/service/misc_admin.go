package service

import (
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *MiscService) AdminChainAdd(ctx *reqctx.ReqCtx, req *entity.ChainAddReq) (*entity.ChainAddRes, error) {
	chain := &model.Chain{
		LogoURL:    req.LogoURL,
		Base64Icon: req.Base64Icon,
	}
	if !ctx.IsZHLang {
		chain.Name = req.Name
		chain.Description = req.Description
	} else {
		chain.NameZH = req.Name
		chain.DescriptionZH = req.Description
	}

	if err := s.miscDao.ChainAdd(ctx, chain); err != nil {
		return nil, err
	}
	return &entity.ChainAddRes{
		ID: chain.ID,
	}, nil
}

func (s *MiscService) AdminTrackAdd(ctx *reqctx.ReqCtx, req *entity.TrackAddReq) (*entity.TrackAddRes, error) {
	track := &model.Track{}
	if !ctx.IsZHLang {
		track.Name = req.Name
		track.Description = req.Description
	} else {
		track.NameZH = req.Name
		track.DescriptionZH = req.Description
	}

	if err := s.miscDao.TrackAdd(ctx, track); err != nil {
		return nil, err
	}
	return &entity.TrackAddRes{
		ID: track.ID,
	}, nil
}

func (s *MiscService) AdminTeamImpressionAdd(ctx *reqctx.ReqCtx, req *entity.TeamImpressionAddReq) (*entity.TeamImpressionAddRes, error) {
	t := &model.TeamImpression{}
	if !ctx.IsZHLang {
		t.Name = req.Name
		t.Description = req.Description
	} else {
		t.NameZH = req.Name
		t.DescriptionZH = req.Description
	}

	if err := s.miscDao.TeamImpressionAdd(ctx, t); err != nil {
		return nil, err
	}
	return &entity.TeamImpressionAddRes{
		ID: t.ID,
	}, nil
}
