package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wyt-labs/wyt-core/internal/core/model"
)

// ----- chain -----

type ChainAddReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
	Base64Icon  string `json:"base_64_icon"`
}

type ChainAddRes struct {
	ID primitive.ObjectID `json:"id"`
}

type ChainUpdateReq struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
}

type ChainUpdateRes struct {
}

type ChainBatchQueryReq struct {
	IDs []string `json:"id"`
}

type ChainBatchQueryRes struct {
	Chains map[string]*model.Chain `json:"chains"`
}

type ChainListReq struct {
	Page uint64 `json:"page" form:"page"`
	Size uint64 `json:"size" form:"size"`
}

type ChainListRes struct {
	List  []*model.Chain `json:"list"`
	Total int64          `json:"total"`
}

// ----- track -----

type TrackAddReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TrackAddRes struct {
	ID primitive.ObjectID `json:"id"`
}

type TrackUpdateReq struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TrackUpdateRes struct {
}

type TrackBatchQueryReq struct {
	IDs []string `json:"id"`
}

type TrackBatchQueryRes struct {
	Tracks map[string]*model.Track `json:"tracks"`
}

type TrackListReq struct {
	Page  uint64 `json:"page" form:"page"`
	Size  uint64 `json:"size" form:"size"`
	IsHot bool   `json:"is_hot" form:"is_hot"`
}

type TrackListRes struct {
	List  []*model.Track `json:"list"`
	Total int64          `json:"total"`
}

// ----- tag -----

type TagAddReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Score       uint64 `json:"score"`
}

type TagAddRes struct {
	ID primitive.ObjectID `json:"id"`
}

type TagUpdateReq struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Score       uint64 `json:"score"`
}

type TagUpdateRes struct {
}

type TagBatchQueryReq struct {
	IDs []string `json:"id"`
}

type TagBatchQueryRes struct {
	Tags map[string]*model.Tag `json:"tags"`
}

type TagListReq struct {
	Page  uint64 `json:"page" form:"page"`
	Size  uint64 `json:"size" form:"size"`
	IsHot bool   `json:"is_hot" form:"is_hot"`
	Query string `json:"query" form:"query"`
}

type TagListRes struct {
	List  []*model.Tag `json:"list"`
	Total int64        `json:"total"`
}

// ----- team impressions -----

type TeamImpressionAddReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TeamImpressionAddRes struct {
	ID primitive.ObjectID `json:"id"`
}

type TeamImpressionUpdateReq struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TeamImpressionUpdateRes struct {
}

type TeamImpressionBatchQueryReq struct {
	IDs []string `json:"id"`
}

type TeamImpressionBatchQueryRes struct {
	TeamImpressions map[string]*model.TeamImpression `json:"team_impressions"`
}

type TeamImpressionListReq struct {
	Page uint64 `json:"page" form:"page"`
	Size uint64 `json:"size" form:"size"`
}

type TeamImpressionListRes struct {
	List  []*model.TeamImpression `json:"list"`
	Total int64                   `json:"total"`
}

// ----- investor -----

type InvestorAddReq struct {
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	AvatarURL        string                `json:"avatar_url"`
	Subject          model.InvestorSubject `json:"subject"`
	Type             model.InvestorType    `json:"type"`
	SocialMediaLinks []model.LinkInfo      `json:"social_media_links"`
}

type InvestorAddRes struct {
	ID primitive.ObjectID `json:"id"`
}

type InvestorUpdateReq struct {
	ID               string                `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	AvatarURL        string                `json:"avatar_url"`
	Subject          model.InvestorSubject `json:"subject"`
	Type             model.InvestorType    `json:"type"`
	SocialMediaLinks []model.LinkInfo      `json:"social_media_links"`
}

type InvestorUpdateRes struct {
}

type InvestorBatchQueryReq struct {
	IDs []string `json:"id"`
}

type InvestorBatchQueryRes struct {
	Investors map[string]*model.Investor `json:"investors"`
}

type InvestorListReq struct {
	Page  uint64 `json:"page" form:"page"`
	Size  uint64 `json:"size" form:"size"`
	Query string `json:"query" form:"query"`
}

type InvestorListRes struct {
	List  []*model.Investor `json:"list"`
	Total int64             `json:"total"`
}
