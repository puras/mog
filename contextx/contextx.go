package contextx

import (
	"context"

	"gorm.io/gorm"
)

type (
	traceIdCtx struct{}
	transCtx   struct{}
	rowLockCtx struct{}
	userIdCtx  struct{}
)

func NewTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, traceIdCtx{}, traceId)
}

func FromTracdId(ctx context.Context) string {
	v := ctx.Value(traceIdCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

func NewTrans(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, transCtx{}, db)
}

func FromTrans(ctx context.Context) (*gorm.DB, bool) {
	v := ctx.Value(transCtx{})
	if v != nil {
		return v.(*gorm.DB), true
	}
	return nil, false
}

func NewRowLock(ctx context.Context) context.Context {
	return context.WithValue(ctx, rowLockCtx{}, true)
}

func FromRowLock(ctx context.Context) bool {
	v := ctx.Value(rowLockCtx{})
	return v != nil && v.(bool)
}

func NewUserId(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, userIdCtx{}, userId)
}

func FromUserId(ctx context.Context) string {
	v := ctx.Value(userIdCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}
