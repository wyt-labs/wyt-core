package model

type Subscribe struct {
	BaseModel      `bson:"inline"`
	Email          string `json:"email" bson:"email"`
	IsUnsubscribed bool   `json:"is_unsubscribed" bson:"is_unsubscribed"`
}
