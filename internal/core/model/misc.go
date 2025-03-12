package model

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntityType = uint32

const (
	EntityTypeUnknow  EntityType = iota
	EntityTypeProject EntityType = iota
	EntityTypePeople  EntityType = iota
	EntityTypeOrg     EntityType = iota
)

type EntityAlias struct {
	Alias      string             `json:"alias" bson:"_id"`
	EntityType EntityType         `json:"entity_type" bson:"entity_type"`
	EntityName string             `json:"entity_name" bson:"entity_name"`
	EntityID   primitive.ObjectID `json:"entity_id" bson:"entity_id"`
	CreateTime JSONTime           `json:"create_time" bson:"create_time"`
}

type Chain struct {
	BaseModel `bson:"inline"`

	// index
	Name string `json:"name" bson:"name,omitempty"`

	Description string `json:"description" bson:"description"`
	LogoURL     string `json:"logo_url" bson:"logo_url"`
	Base64Icon  string `json:"base_64_icon" bson:"base_64_icon"`
	Score       uint64 `json:"score" bson:"score"`

	// zh
	NameZH string `json:"-" bson:"name_zh,omitempty"`

	DescriptionZH string `json:"-" bson:"description_zh"`
}

func (m *Chain) Translate(isZHLang bool) {
	if isZHLang {
		if m.NameZH != "" {
			m.Name = m.NameZH
		}
		if m.DescriptionZH != "" {
			m.Description = m.DescriptionZH
		}
	}
}

type Track struct {
	BaseModel `bson:"inline"`

	// index
	Name string `json:"name" bson:"name,omitempty"`

	Description string `json:"description" bson:"description"`
	Score       uint64 `json:"score" bson:"score"`

	// zh
	NameZH string `json:"-" bson:"name_zh,omitempty"`

	DescriptionZH string `json:"-" bson:"description_zh"`
}

func (m *Track) Translate(isZHLang bool) {
	if isZHLang {
		if m.NameZH != "" {
			m.Name = m.NameZH
		}
		if m.DescriptionZH != "" {
			m.Description = m.DescriptionZH
		}
	}
}

type Tag struct {
	BaseModel `bson:"inline"`

	// index
	Name string `json:"name" bson:"name,omitempty"`

	Description string `json:"description" bson:"description"`
	Score       uint64 `json:"score" bson:"score"`

	// zh
	NameZH string `json:"-" bson:"name_zh,omitempty"`

	DescriptionZH string `json:"-" bson:"description_zh"`
}

func (m *Tag) Translate(isZHLang bool) {
	if isZHLang {
		if m.NameZH != "" {
			m.Name = m.NameZH
		}
		if m.DescriptionZH != "" {
			m.Description = m.DescriptionZH
		}
	}
}

type TeamImpression struct {
	BaseModel `bson:"inline"`

	// index
	Name string `json:"name" bson:"name,omitempty"`

	Description string `json:"description" bson:"description"`

	// zh
	NameZH string `json:"-" bson:"name_zh,omitempty"`

	DescriptionZH string `json:"-" bson:"description_zh"`
}

func (m *TeamImpression) Translate(isZHLang bool) {
	if isZHLang {
		if m.NameZH != "" {
			m.Name = m.NameZH
		}
		if m.DescriptionZH != "" {
			m.Description = m.DescriptionZH
		}
	}
}

type InvestorSubject uint32

const (
	InvestorSubjectInstitution InvestorSubject = iota
	InvestorSubjectOrgIndividual
	InvestorSubjectUnknown
)

func (s InvestorSubject) Validate() error {
	if uint32(s) > uint32(InvestorSubjectUnknown) {
		return errors.New("invalid investor subject")
	}
	return nil
}

type InvestorType uint32

const (
	InvestorTypeKOL InvestorType = iota
	InvestorTypeTop
	InvestorTypeUnknown
)

func (t InvestorType) Validate() error {
	if uint32(t) > uint32(InvestorTypeUnknown) {
		return errors.New("invalid investor type")
	}
	return nil
}

type Investor struct {
	BaseModel        `bson:"inline"`
	Name             string          `json:"name" bson:"name"`
	Description      string          `json:"description" bson:"description"`
	AvatarURL        string          `json:"avatar_url" bson:"avatar_url"`
	Subject          InvestorSubject `json:"subject" bson:"subject"`
	Type             InvestorType    `json:"type" bson:"type"`
	SocialMediaLinks []LinkInfo      `json:"social_media_links" bson:"social_media_links"`
	IsTop            bool            `json:"is_top" bson:"is_top"`

	// zh
	DescriptionZH string `json:"-" bson:"description_zh"`
}

func (m *Investor) Translate(isZHLang bool) {
	if isZHLang {
		if m.DescriptionZH != "" {
			m.Description = m.DescriptionZH
		}
	}
}
