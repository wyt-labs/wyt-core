package entity

import (
	"io"
	"time"
)

type File struct {
	ID         string
	Name       string
	Length     int64
	UploadDate time.Time
	Metadata   string
	Data       io.Reader
}

type BucketType uint32

const (
	BucketTypeMisc BucketType = iota
	BucketTypeAvatar
	BucketTypeProject
)

func (t BucketType) String() string {
	switch t {
	case BucketTypeMisc:
		return "misc"
	case BucketTypeAvatar:
		return "avatar"
	case BucketTypeProject:
		return "project"
	default:
		return "misc"
	}
}

type FileUploadReq struct {
	Type BucketType `json:"type" form:"type"`
}

type FileUploadRes struct {
	URL string `json:"url"`
}
