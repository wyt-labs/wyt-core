package coreapi

import (
	"github.com/wyt-labs/wyt-core/internal/core/component/datapuller"
	"github.com/wyt-labs/wyt-core/internal/core/component/okxswap"
	"github.com/wyt-labs/wyt-core/internal/core/service"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/pkg/basic"
	"github.com/wyt-labs/wyt-core/pkg/extension"
	"github.com/wyt-labs/wyt-core/pkg/mutex"
)

func init() {
	basic.RegisterComponents(NewCoreAPI, mutex.NewKeyMutex, NewChatgptDriver)
}

type CoreAPI struct {
	UserService       *service.UserService
	ProjectService    *service.ProjectService
	MiscService       *service.MiscService
	FileSystemService *service.FileSystemService
	ChatService       *service.ChatService
	WebsiteService    *service.WebsiteService
	PumpDataService   *datapuller.PumpDataService
	OkxDexServiceApi  *okxswap.OkxSwapApi
}

func NewCoreAPI(
	baseComponent *base.Component,
	userService *service.UserService,
	projectService *service.ProjectService,
	miscService *service.MiscService,
	fileSystemService *service.FileSystemService,
	chatService *service.ChatService,
	websiteService *service.WebsiteService,
	pumpDataService *datapuller.PumpDataService,
	okxDexServiceApi *okxswap.OkxSwapApi,
) (*CoreAPI, error) {
	baseComponent.Logger.Info("core api init")
	return &CoreAPI{
		UserService:       userService,
		ProjectService:    projectService,
		MiscService:       miscService,
		FileSystemService: fileSystemService,
		ChatService:       chatService,
		WebsiteService:    websiteService,
		PumpDataService:   pumpDataService,
		OkxDexServiceApi:  okxDexServiceApi,
	}, nil
}

func NewChatgptDriver(
	baseComponent *base.Component,
	pumpDataService *datapuller.PumpDataService,
) (*extension.ChatgptDriver, error) {
	baseComponent.Logger.Info("chatgpt driver init")
	cfg := baseComponent.Config.Extension.Chatgpt
	return extension.NewChatgptDriver(&extension.ChatgptConfig{
		Endpoint:        cfg.Endpoint,
		EndpointFull:    cfg.EndpointFull,
		APIKey:          cfg.APIKey,
		Model:           cfg.Model,
		Temperature:     cfg.Temperature,
		PresencePenalty: cfg.PresencePenalty,
	}, baseComponent, pumpDataService)
}
