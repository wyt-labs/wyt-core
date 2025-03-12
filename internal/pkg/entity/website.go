package entity

type WebsiteSubscribeReq struct {
	Email string `json:"email" form:"email"`
}

type WebsiteSubscribeRes struct {
}

type WebsiteUnsubscribeReq struct {
	Email string `json:"email" form:"email"`
}

type WebsiteUnsubscribeRes struct {
}
