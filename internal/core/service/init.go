package service

import (
	"github.com/wyt-labs/wyt-core/pkg/basic"
)

func init() {
	basic.RegisterComponents(
		NewUserService,
		NewProjectService,
		NewMiscService,
		NewFileSystemService,
		NewChatService,
		NewWebsiteService,
	)
}
