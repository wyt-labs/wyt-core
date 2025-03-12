package reqctx

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/wyt-labs/wyt-core/internal/core/model"
)

type ReqCtx struct {
	Ctx                    context.Context
	Logger                 logrus.FieldLogger
	RequestID              int64
	Caller                 string
	CallerRole             model.UserRole
	CallerStatus           model.UserStatus
	IsZHLang               bool
	Lock                   *sync.RWMutex
	values                 map[any]any
	customLogFields        map[string]any
	customLogFieldsOnError map[string]any
}

func NewReqCtx(ctx context.Context, logger logrus.FieldLogger, requestID int64, caller string) *ReqCtx {
	return &ReqCtx{
		Ctx: ctx,
		Logger: logger.WithFields(logrus.Fields{
			"req_id": requestID,
		}),
		RequestID:              requestID,
		Caller:                 caller,
		Lock:                   new(sync.RWMutex),
		values:                 map[any]any{},
		customLogFields:        map[string]any{},
		customLogFieldsOnError: map[string]any{},
	}
}

func (ctx *ReqCtx) AddCustomLogField(key string, value any) {
	ctx.customLogFields[key] = value
}

func (ctx *ReqCtx) AddCustomLogFields(fields map[string]any) {
	for key, value := range fields {
		ctx.customLogFields[key] = value
	}
}

func (ctx *ReqCtx) AddCustomLogFieldOnError(key string, value any) {
	ctx.customLogFieldsOnError[key] = value
}

func (ctx *ReqCtx) AddCustomLogFieldsOnError(fields map[string]any) {
	for key, value := range fields {
		ctx.customLogFieldsOnError[key] = value
	}
}

func (ctx *ReqCtx) PutValue(key any, value any) {
	ctx.values[key] = value
}

func (ctx *ReqCtx) GetValue(key any) any {
	return ctx.values[key]
}

func (ctx *ReqCtx) Clone() *ReqCtx {
	c := make(map[any]any, len(ctx.values))
	for s, i := range ctx.values {
		c[s] = i
	}
	return &ReqCtx{
		Ctx:                    ctx.Ctx,
		Logger:                 ctx.Logger,
		RequestID:              ctx.RequestID,
		Caller:                 ctx.Caller,
		Lock:                   new(sync.RWMutex),
		values:                 c,
		customLogFields:        ctx.customLogFields,
		customLogFieldsOnError: ctx.customLogFieldsOnError,
	}
}

func (ctx *ReqCtx) CombineCustomLogFields(target map[string]any) {
	for key, value := range ctx.customLogFields {
		target[key] = value
	}
}

func (ctx *ReqCtx) CombineCustomLogFieldsOnError(target map[string]any) {
	for key, value := range ctx.customLogFieldsOnError {
		target[key] = value
	}
}
