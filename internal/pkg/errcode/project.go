package errcode

var (
	ErrProjectNotExist      = NewCustomError(10301, "project not exist")
	ErrProjectPublished     = NewCustomError(10302, "project has been published")
	ErrProjectAlreadyExists = NewCustomError(10303, "project already exists")
)
