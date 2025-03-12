package errcode

var (
	ErrRequestParameter  = NewCustomError(10002, "error request parameter")
	ErrAuthCode          = NewCustomError(10003, "error auth token")
	ErrAccountStatus     = NewCustomError(10004, "abnormal account status")
	ErrAccountPermission = NewCustomError(10005, "no permission")
	ErrPjWinIdErr        = NewCustomError(10006, "req's conversation id not constant with the window id")
)
