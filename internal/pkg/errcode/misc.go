package errcode

var (
	ErrChainNotExist              = NewCustomError(10201, "chain not exist")
	ErrChainAlreadyExist          = NewCustomError(10202, "chain already exists")
	ErrTrackNotExist              = NewCustomError(10203, "track not exist")
	ErrTrackAlreadyExist          = NewCustomError(10204, "track already exists")
	ErrTagNotExist                = NewCustomError(10205, "tag not exist")
	ErrTagAlreadyExist            = NewCustomError(10206, "tag already exists")
	ErrTeamImpressionNotExist     = NewCustomError(10207, "team impression not exist")
	ErrTeamImpressionAlreadyExist = NewCustomError(10208, "team impression already exists")
	ErrInvestorNotExist           = NewCustomError(10209, "investor not exist")
)
