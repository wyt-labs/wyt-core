package errcode

var (
	ErrUserNotExist      = NewCustomError(10101, "user not exist")
	ErrUserPinNotExist   = NewCustomError(10102, "user pin agent record not exist")
	ErrUserPinPjNotExist = NewCustomError(10103, "project_id cannot be empty")
)
