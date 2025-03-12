package model

type UserPlugin struct {
	BaseModel `bson:"inline"`

	UserId      string `json:"user_id" bson:"user_id"`
	ProjectId   string `json:"project_id" bson:"project_id"`
	ProjectName string `json:"project_name" bson:"project_name"`

	PinStatus int `json:"pin_status" bson:"pin_status"`
}

type UserPluginDto struct {
	UserId      string `json:"user_id" bson:"user_id"`
	ProjectId   string `json:"project_id" bson:"project_id"`
	ProjectName string `json:"project_name" bson:"project_name"`
	PinStatus   int    `json:"pin_status" bson:"pin_status"`
	PinTime     int64  `json:"pin_time" bson:"pin_time"`
}
