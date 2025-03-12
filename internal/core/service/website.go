package service

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/pkg/email"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

type WebsiteService struct {
	baseComponent *base.Component
	websiteDao    *dao.WebsiteDao

	helloEmailTemplate string
}

func NewWebsiteService(baseComponent *base.Component, websiteDao *dao.WebsiteDao) *WebsiteService {
	s := &WebsiteService{
		baseComponent: baseComponent,
		websiteDao:    websiteDao,
	}
	baseComponent.RegisterLifecycleHook(s)
	return s
}

func (s *WebsiteService) Start() error {
	helloEmailTemplate, err := os.ReadFile(filepath.Join(s.baseComponent.Config.RootPath, config.WelcomeEmailTemplateName))
	if err != nil {
		return errors.Errorf("failed to read hello email template: %v", err)
	}
	s.helloEmailTemplate = string(helloEmailTemplate)
	return nil
}

func (s *WebsiteService) Stop() error {
	return nil
}

func (s *WebsiteService) Subscribe(ctx *reqctx.ReqCtx, req *entity.WebsiteSubscribeReq) (*entity.WebsiteSubscribeRes, error) {
	if req.Email == "" {
		return &entity.WebsiteSubscribeRes{}, nil
	}

	asyncSendWelcomeEmail := func() {
		s.baseComponent.SafeGo(func() {
			err := util.Retry(s.baseComponent.Config.App.RetryInterval.ToDuration(), s.baseComponent.Config.App.RetryTime, func() (needRetry bool, err error) {
				err = sendWelcomeEmail(email.Cfg{
					SenderAddress:  s.baseComponent.Config.App.Email.SenderAddress,
					SenderName:     s.baseComponent.Config.App.Email.SenderName,
					SenderPwd:      s.baseComponent.Config.App.Email.SenderPwd,
					MailServerHost: s.baseComponent.Config.App.Email.MailServerHost,
					MailServerPort: s.baseComponent.Config.App.Email.MailServerPort,
				}, s.helloEmailTemplate, req.Email, s.baseComponent.Config.App.Email.UnsubscribeFrontendURL)
				if err != nil {
					return true, err
				}
				return false, nil
			})
			if err != nil {
				s.baseComponent.Logger.WithFields(logrus.Fields{
					"err":   err,
					"email": req.Email,
				}).Error("Failed to send welcome email")
			}
		})
	}

	var info *model.Subscribe
	var err error
	info, err = s.websiteDao.SubscribeQueryByEmail(ctx, req.Email)
	if err != nil {
		if err == dao.ErrSubscribeNotExist {
			if err := s.websiteDao.SubscribeAdd(ctx, &model.Subscribe{
				Email:          req.Email,
				IsUnsubscribed: false,
			}); err != nil {
				return nil, err
			}
			asyncSendWelcomeEmail()
			return &entity.WebsiteSubscribeRes{}, nil
		}
		return nil, err
	}

	if info.IsUnsubscribed {
		info.IsUnsubscribed = false
		if err := s.websiteDao.SubscribeUpdate(ctx, info); err != nil {
			return nil, err
		}
	}
	asyncSendWelcomeEmail()
	return &entity.WebsiteSubscribeRes{}, nil
}

func (s *WebsiteService) Unsubscribe(ctx *reqctx.ReqCtx, req *entity.WebsiteUnsubscribeReq) (*entity.WebsiteUnsubscribeRes, error) {
	if req.Email == "" {
		return &entity.WebsiteUnsubscribeRes{}, nil
	}
	var info *model.Subscribe
	var err error
	info, err = s.websiteDao.SubscribeQueryByEmail(ctx, req.Email)
	if err != nil {
		if err == dao.ErrSubscribeNotExist {
			if err := s.websiteDao.SubscribeAdd(ctx, &model.Subscribe{
				Email:          req.Email,
				IsUnsubscribed: true,
			}); err != nil {
				return nil, err
			}
			return &entity.WebsiteUnsubscribeRes{}, nil
		}
		return nil, err
	}

	if !info.IsUnsubscribed {
		info.IsUnsubscribed = true
		if err := s.websiteDao.SubscribeUpdate(ctx, info); err != nil {
			return nil, err
		}
	}

	return &entity.WebsiteUnsubscribeRes{}, nil
}

func sendWelcomeEmail(cfg email.Cfg, welcomeEmailTemplate string, userEmail string, unsubscribeURL string) error {
	helloEmail := strings.ReplaceAll(welcomeEmailTemplate, "{{userEmail}}", userEmail)
	params := url.Values{}
	params.Add("email", userEmail)
	helloEmail = strings.ReplaceAll(helloEmail, "{{unsubscribeURL}}", unsubscribeURL+"?"+params.Encode())

	return email.SendEmail(cfg, email.Msg{
		ToAddress: userEmail,
		Title:     "Welcome to wyt",
		Content:   helloEmail,
	})
}
