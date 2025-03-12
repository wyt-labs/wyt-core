package rest

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/coreapi"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/auth/jwt"
	"github.com/wyt-labs/wyt-core/pkg/basic"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func init() {
	basic.RegisterComponents(New)
}

type Server struct {
	baseComponent *base.Component
	router        *gin.Engine
	listener      net.Listener
	hs            *http.Server
	*coreapi.CoreAPI
}

func New(baseComponent *base.Component, api *coreapi.CoreAPI) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	s := &Server{
		baseComponent: baseComponent,
		router:        router,
		hs: &http.Server{
			Addr:           fmt.Sprintf(":%d", baseComponent.Config.HTTP.Port),
			Handler:        router,
			ReadTimeout:    baseComponent.Config.HTTP.ReadTimeout.ToDuration(),
			WriteTimeout:   baseComponent.Config.HTTP.WriteTimeout.ToDuration(),
			MaxHeaderBytes: 1 << 20,
		},
		CoreAPI: api,
	}
	baseComponent.RegisterLifecycleHook(s)
	return s
}

func (s *Server) Start() error {
	err := s.init()
	if err != nil {
		return errors.Wrap(err, "register router failed")
	}

	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.baseComponent.Config.HTTP.Port))
	if err != nil {
		return err
	}
	printServerInfo := func() {
		s.baseComponent.Logger.Infof("Http server listen on: %d", s.baseComponent.Config.HTTP.Port)
	}

	s.baseComponent.SafeGoPersistentTask(func() {
		err := func() error {
			if s.baseComponent.Config.HTTP.TLSEnable {
				if _, err := os.Stat(s.baseComponent.Config.HTTP.TLSCertFilePath); err != nil {
					return errors.Wrapf(err, "tls_cert_file_path [%s] is invalid path", s.baseComponent.Config.HTTP.TLSCertFilePath)
				}
				if _, err := os.Stat(s.baseComponent.Config.HTTP.TLSKeyFilePath); err != nil {
					return errors.Wrapf(err, "tls_key_file_path [%s] is invalid path", s.baseComponent.Config.HTTP.TLSKeyFilePath)
				}
				printServerInfo()

				if err := s.hs.ServeTLS(s.listener, s.baseComponent.Config.HTTP.TLSCertFilePath, s.baseComponent.Config.HTTP.TLSKeyFilePath); err != nil {
					return err
				}
			} else {
				printServerInfo()
				if err := s.hs.Serve(s.listener); err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil && err != http.ErrServerClosed {
			s.baseComponent.Logger.WithFields(logrus.Fields{"err": err, "port": s.baseComponent.Config.HTTP.Port}).Warn("Failed to start http server")
			s.baseComponent.ComponentShutdown()
			return
		}
		s.baseComponent.Logger.Info("Http server shutdown")
	})

	return nil
}

func (s *Server) Stop() error {
	return s.hs.Close()
}

func (s *Server) init() error {
	s.router.MaxMultipartMemory = s.baseComponent.Config.HTTP.MultipartMemory
	s.router.Use(s.crossOriginMiddleware)

	{
		v := s.router.Group("/api/v1")
		{
			g := v.Group("/user")
			g.GET("/nonce", s.apiHandlerWrap(s.userNonce))
			g.POST("/signin", s.apiHandlerWrap(s.userSignin))
			g.POST("/login", s.apiHandlerWrap(s.userLogin))
			g.GET("/refresh-token", s.apiHandlerWrap(s.userRefreshToken, apiNeedAuth()))
			g.GET("/info", s.apiHandlerWrap(s.userInfo, apiNeedAuth()))
			g.POST("/update", s.apiHandlerWrap(s.userUpdate, apiNeedAuth()))
			g.POST("/set-role", s.apiHandlerWrap(s.adminUserSetRole, apiNeedAdmin()))
			g.POST("/set-is-lock", s.apiHandlerWrap(s.adminUserSetIsLock, apiNeedAdmin()))

			// dev api
			if s.baseComponent.IsDevVersion() {
				// not check signature and nonce
				g.POST("/dev-login", s.apiHandlerWrap(s.userDevLogin))
			}
		}
		{
			g := v.Group("/fs")
			g.POST("/upload", s.apiHandlerWrap(s.fsUpload, apiNeedAuth()))
			g.GET("/files/:bucket/:id", s.fsContent)
		}
		{
			g := v.Group("/misc")
			g.POST("/chain/add", s.apiHandlerWrap(s.adminChainAdd, apiNeedAdmin()))
			g.GET("/chain/list", s.apiHandlerWrap(s.chainList))
			g.POST("/track/add", s.apiHandlerWrap(s.adminTrackAdd, apiNeedAdmin()))
			g.GET("/track/list", s.apiHandlerWrap(s.trackList))
			g.POST("/tag/add", s.apiHandlerWrap(s.tagAdd, apiNeedAuth()))
			g.GET("/tag/list", s.apiHandlerWrap(s.tagList))
			g.POST("/team-impression/add", s.apiHandlerWrap(s.adminTeamImpressionAdd, apiNeedAdmin()))
			g.GET("/team-impression/list", s.apiHandlerWrap(s.teamImpressionList))
			g.POST("/investor/add", s.apiHandlerWrap(s.investorAdd, apiNeedAuth()))
			g.POST("/investor/update", s.apiHandlerWrap(s.investorUpdate, apiNeedAuth()))
			g.GET("/investor/list", s.apiHandlerWrap(s.investorList))
		}

		{
			g := v.Group("/project")
			g.POST("/add", s.apiHandlerWrap(s.adminProjectAdd, apiNeedAdmin()))
			g.GET("/info-edit", s.apiHandlerWrap(s.adminProjectInfo, apiNeedAdmin()))
			g.GET("/simple-info-edit", s.apiHandlerWrap(s.adminProjectSimpleInfo, apiNeedAdmin()))
			g.POST("/list-edit", s.apiHandlerWrap(s.adminProjectList, apiNeedAdmin()))
			g.POST("/update", s.apiHandlerWrap(s.adminProjectUpdate, apiNeedAdmin()))
			g.POST("/simple-update", s.apiHandlerWrap(s.adminProjectSimpleUpdate, apiNeedAdmin()))
			g.POST("/publish", s.apiHandlerWrap(s.adminProjectPublish, apiNeedAdmin()))
			g.POST("/delete", s.apiHandlerWrap(s.adminProjectDelete, apiNeedAdmin()))
			g.POST("/calculate-derived-data", s.apiHandlerWrap(s.adminProjectCalculateDerivedData, apiNeedAdmin()))

			g.GET("/info-view", s.apiHandlerWrap(s.projectInfo))
			g.GET("/simple-info-view", s.apiHandlerWrap(s.projectSimpleInfo))
			g.POST("/list-view", s.apiHandlerWrap(s.projectList))
			g.GET("/info-compare", s.apiHandlerWrap(s.projectInfoCompare))
			g.GET("/metrics-compare", s.apiHandlerWrap(s.projectMetricsCompare))
		}

		{
			g := v.Group("/agent")
			g.GET("/list", s.apiHandlerWrap(s.agentList, apiNeedAuth()))
			g.POST("/pin", s.apiHandlerWrap(s.pinAgent, apiNeedAuth()))
			g.POST("/unpin", s.apiHandlerWrap(s.unpinAgent, apiNeedAuth()))
		}

		{
			g := v.Group("/chat")
			g.POST("/create", s.apiHandlerWrap(s.chatCreate, apiNeedAuth()))
			g.GET("/list", s.apiHandlerWrap(s.chatList, apiNeedAuth()))
			g.GET("/history", s.apiHandlerWrap(s.chatHistory, apiNeedAuth()))
			// chat
			g.POST("/completions", s.apiHandlerWrap(s.chatCompletionsJsonMode, apiNeedAuth()))
			g.POST("/update", s.apiHandlerWrap(s.chatUpdate, apiNeedAuth()))
			g.POST("/delete", s.apiHandlerWrap(s.chatDelete, apiNeedAuth()))
			g.POST("/delete-all", s.apiHandlerWrap(s.chatDeleteAll, apiNeedAuth()))
		}

		{
			g := v.Group("/data/pump")
			g.GET("/new-tokens", s.apiHandlerWrap(s.dataPumpNewTokens, apiNeedAuth()))
			g.GET("/launch-time", s.apiHandlerWrap(s.dataPumpLaunchTime, apiNeedAuth()))
			g.GET("/transactions", s.apiHandlerWrap(s.dataPumpTransactions, apiNeedAuth()))
			g.GET("/top-traders", s.apiHandlerWrap(s.dataPumpTopTraders, apiNeedAuth()))
			g.GET("/trader/info", s.apiHandlerWrap(s.dataPumpTraderInfo, apiNeedAuth()))
			g.GET("/trader/overview", s.apiHandlerWrap(s.dataPumpTraderOverview, apiNeedAuth()))
			g.GET("/trader/profit", s.apiHandlerWrap(s.dataPumpTraderProfit, apiNeedAuth()))
			g.GET("/trader/profit-distribution", s.apiHandlerWrap(s.dataPumpTraderProfitDistribution, apiNeedAuth()))
			g.GET("/trader/trades", s.apiHandlerWrap(s.dataPumpTraderTrades, apiNeedAuth()))
			g.GET("/trader/detail", s.apiHandlerWrap(s.dataPumpTraderDetail, apiNeedAuth()))
		}

		{
			g := v.Group("/website")
			g.POST("/subscribe", s.apiHandlerWrap(s.websiteSubscribe))
			g.POST("/unsubscribe", s.apiHandlerWrap(s.websiteUnsubscribe))
		}

		{
			v.GET("/dex/aggregator/supported/chain", s.apiHandlerWrap(s.dexSupportedChain))
			v.GET("/dex/aggregator/all-tokens", s.apiHandlerWrap(s.dexSupportedAllTokens))
			v.GET("/dex/aggregator/token", s.apiHandlerWrap(s.tokenInfo))
			v.GET("/dex/aggregator/quote", s.apiHandlerWrap(s.getQuote))
			v.GET("/dex/aggregator/approve-transaction", s.apiHandlerWrap(s.approveTransaction))
			v.GET("/dex/aggregator/swap", s.apiHandlerWrap(s.swap))
			v.GET("/dex/cross-chain/supported/bridge-tokens-pairs", s.apiHandlerWrap(s.bridgeTokensPairs))
			v.GET("/dex/cross-chain/quote", s.apiHandlerWrap(s.crossChainQuote))
		}
	}

	// dev enable pprof
	if s.baseComponent.IsDevVersion() {
		s.router.GET("/debug/pprof/", IndexHandler())
		s.router.GET("/debug/pprof/heap", HeapHandler())
		s.router.GET("/debug/pprof/goroutine", GoroutineHandler())
		s.router.GET("/debug/pprof/allocs", AllocsHandler())
		s.router.GET("/debug/pprof/block", BlockHandler())
		s.router.GET("/debug/pprof/threadcreate", ThreadCreateHandler())
		s.router.GET("/debug/pprof/cmdline", CmdlineHandler())
		s.router.GET("/debug/pprof/profile", ProfileHandler())
		s.router.GET("/debug/pprof/symbol", SymbolHandler())
		s.router.GET("/debug/pprof/trace", TraceHandler())
		s.router.GET("/debug/pprof/mutex", MutexHandler())
	}

	return nil
}

func (s *Server) crossOriginMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "token, origin, content-type, accept, is_zh")
	c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}

func (s *Server) generateRequestContext(c *gin.Context) *reqctx.ReqCtx {
	reqID := s.baseComponent.UUIDGenerator.Generate()
	ctx := reqctx.NewReqCtx(c.Request.Context(), s.baseComponent.Logger, int64(reqID), "")
	return ctx
}

type apiConfig struct {
	needAuth       bool
	needAdmin      bool
	customResponse func(c *gin.Context)
}

type apiConfigOption func(*apiConfig)

func apiNeedAuth() apiConfigOption {
	return func(c *apiConfig) {
		c.needAuth = true
	}
}

func apiNeedAdmin() apiConfigOption {
	return func(c *apiConfig) {
		c.needAdmin = true
	}
}

// nolint
func apiCustomResponse(customResponse func(c *gin.Context)) apiConfigOption {
	return func(c *apiConfig) {
		c.customResponse = customResponse
	}
}

func newAPIConfig(opts ...apiConfigOption) apiConfig {
	apiCfg := &apiConfig{
		needAuth:  false,
		needAdmin: false,
	}
	for _, opt := range opts {
		opt(apiCfg)
	}
	return *apiCfg
}

func (s *Server) apiHandlerWrap(handler func(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error), opts ...apiConfigOption) func(c *gin.Context) {
	cfg := newAPIConfig(opts...)
	return func(c *gin.Context) {
		ctx := s.generateRequestContext(c)
		startTime := time.Now()
		reqURI := c.Request.URL.Path
		var res any
		err := s.baseComponent.RecoverExecute(func() error {
			if cfg.needAuth || cfg.needAdmin {
				token := c.GetHeader(config.JWTTokenHeaderKey)
				if token == "" {
					return errcode.ErrAuthCode.Wrap("token is empty")
				}

				var customClaims entity.CustomClaims
				id, err := jwt.ParseWithHMACKey(s.baseComponent.Config.HTTP.JWTTokenHMACKey, token, &customClaims)
				if err != nil {
					return errcode.ErrAuthCode.Wrap(err.Error())
				}
				if id == "" {
					return errcode.ErrAuthCode.Wrap("internal error: token data invalid: id is empty")
				}

				ctx.Caller = id
				ctx.CallerRole = customClaims.CallerRole
				ctx.CallerStatus = customClaims.CallerStatus

				if ctx.Caller == s.baseComponent.Config.App.AdminAddr {
					ctx.CallerRole = model.UserRoleAdmin
				}

				if ctx.CallerStatus != model.UserStatusNormal {
					return errcode.ErrAccountStatus
				}
				if cfg.needAdmin {
					if ctx.CallerRole == model.UserRoleMember {
						return errcode.ErrAccountPermission
					}
				}
			}

			isZhLangStr := c.GetHeader(config.IsZHLangHeaderKey)
			if isZhLangStr != "" {
				isZhLang, err := strconv.ParseBool(isZhLangStr)
				if err == nil {
					ctx.IsZHLang = isZhLang
				}
			}

			var err error
			res, err = handler(ctx, c)
			return err
		})
		endTime := time.Now()

		latencyTime := fmt.Sprintf("%6v", endTime.Sub(startTime))
		reqMethod := c.Request.Method

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		logFields := logrus.Fields{
			"http_code": statusCode,
			"time_cost": latencyTime,
			"ip":        clientIP,
			"method":    reqMethod,
			"uri":       reqURI,
		}
		if ctx.Caller != "" {
			logFields["caller"] = ctx.Caller
		}

		if err != nil {
			if cfg.customResponse == nil {
				s.failResponseWithErr(ctx, c, err)
			} else {
				cfg.customResponse(c)
			}
			ctx.CombineCustomLogFields(logFields)
			ctx.CombineCustomLogFieldsOnError(logFields)
			ctx.Logger.WithFields(logFields).Error("API request failed")
			return
		}
		ctx.CombineCustomLogFields(logFields)
		ctx.Logger.WithFields(logFields).Info("API request")

		if cfg.customResponse == nil {
			s.successResponseWithData(c, res)
		} else {
			cfg.customResponse(c)
		}
	}
}

func (s *Server) failResponseWithErr(ctx *reqctx.ReqCtx, c *gin.Context, err error) {
	code := errcode.DecodeError(err)
	msg := err.Error()

	ctx.AddCustomLogField("err_code", code)
	ctx.AddCustomLogField("err_msg", msg)

	httpCode := http.StatusOK
	if strings.Contains(config.Version, "test") {
		httpCode = http.StatusInternalServerError
	}

	c.JSON(httpCode, gin.H{
		"code":    code,
		"message": msg,
	})
}

func (s *Server) successResponseWithData(c *gin.Context, data any) {
	res := gin.H{
		"code": 0,
	}
	if data != nil {
		res["data"] = data
	}
	c.JSON(http.StatusOK, res)
}
