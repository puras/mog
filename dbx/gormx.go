package dbx

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/puras/mog/contextx"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaginationResult struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
}

type PaginationParam struct {
	Pagination bool `form:"-" default:"true"`
	OnlyCount  bool `form:"-"`
	PageNum    int  `form:"page_num"`
	PageSize   int  `form:"page_size" binding:"max=100"`
}

func (self PaginationParam) GetPaginationParam() PaginationParam {
	return self
}

type QueryOptions struct {
	SelectFields []string
	OmitFields   []string
	OrderFields  OrderByParams
}

type Direction string

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

type OrderByParam struct {
	Field     string
	Direction Direction
}

type OrderByParams []OrderByParam

// identPattern 限定可参与排序的列名：字母或下划线开头，后接字母数字下划线；
// 最多允许 `.` 分隔成 table.column 两段。
//
// 这是 ORDER BY 拼接的第一道防线——OrderFields 常直接来自 HTTP query，
// 走白名单比事后过滤可靠：被拒的字段在拼接前就出局，根本进不了 SQL 文本。
var identPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`)

// normalized 把方向归一成 ASC / DESC：空值补 ASC（`ORDER BY col` 与
// `ORDER BY col ASC` 等价），大小写不敏感，非法值降级为 ASC。
//
// 方向位是排序片段里唯一无法靠引号保护的位置——引号只能包标识符，
// `DESC` 是关键字，必须走白名单，所以归一放在这里而不是交给调用方。
func (d Direction) normalized() Direction {
	switch strings.ToUpper(strings.TrimSpace(string(d))) {
	case string(DESC):
		return DESC
	default:
		return ASC
	}
}

// Validate 严格校验排序参数：字段名不在 identPattern 白名单内、或方向
// 不是 ASC/DESC/空，返回 error 描述第一个非法项。
//
// 与 ToSQL 的"丢弃非法项"策略互补——ToSQL 保证输出安全（宁可少排一列），
// Validate 让 handler 能把用户的 ?order_by=xxx 判成 400 而不是静默忽略。
func (p OrderByParams) Validate() error {
	for _, v := range p {
		if !identPattern.MatchString(strings.TrimSpace(v.Field)) {
			return fmt.Errorf("dbx: 非法的排序字段 %q", v.Field)
		}
		switch d := strings.ToUpper(strings.TrimSpace(string(v.Direction))); d {
		case "", string(ASC), string(DESC):
		default:
			return fmt.Errorf("dbx: 非法的排序方向 %q（仅支持 ASC/DESC）", v.Direction)
		}
	}
	return nil
}

// ToSQL 生成 ORDER BY 片段，多项以 ", " 连接。
//
// 无 db 上下文，因此字段名只走白名单校验、不加引号：白名单内只可能出现
// [A-Za-z0-9_.]，拼接结果天然无注入面。需要方言引号（保留字、显式表名）
// 请用 ToSQLWith。
//
// 保底策略是丢弃非法项而非报错：这是排序，排序不了不该让查询失败。
// 想要错误信号请先调 Validate。全部项非法时返回空串——注意空串意味着
// "本次查询不带 ORDER BY"，调用方需自行跳过 db.Order("")。
func (p OrderByParams) ToSQL() string {
	return p.toSQL(nil)
}

// ToSQLWith 是 ToSQL 的方言感知版本，字段名交给 db.Dialector 加引号：
// MySQL/SQLite 产出 `user`.`name`，PostgreSQL 产出 "user"."name"。
//
// 引号除了防注入，还能让 order / key 这类保留字合法列名正常工作。
// wrapQueryOptions 走的就是这条路径——它手上有 *gorm.DB，没有理由不用。
func (p OrderByParams) ToSQLWith(db *gorm.DB) string {
	return p.toSQL(db)
}

func (p OrderByParams) toSQL(db *gorm.DB) string {
	if len(p) == 0 {
		return ""
	}

	// stmt 只在需要方言引号时建一次，避免逐字段分配。
	var stmt *gorm.Statement
	if db != nil && db.Dialector != nil {
		stmt = &gorm.Statement{DB: db}
	}

	parts := make([]string, 0, len(p))
	for _, v := range p {
		field := strings.TrimSpace(v.Field)
		if !identPattern.MatchString(field) {
			continue
		}
		if stmt != nil {
			field = stmt.Quote(field)
		}
		parts = append(parts, field+" "+string(v.Direction.normalized()))
	}
	return strings.Join(parts, ", ")
}

func (t *Trans) Exec(ctx context.Context, fn TransFunc) error {
	if _, ok := contextx.FromTrans(ctx); ok {
		return fn(ctx)
	}
	// WithContext 必须显式带上：GORM 的 Begin() 用 db.Statement.Context 作为事务 ctx，
	// 而 root DB 的 Statement.Context 是 context.Background()。漏掉它会让事务失去
	// deadline / cancel，调用方 contextx.NewTrans(ctx, ...) 拿到的 tx 也不认 ctx，
	// ctx 过期后语句照跑、事务无法自行收敛（连接带着未关闭的 tx 回到池里）。
	return t.DB.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		return fn(contextx.NewTrans(ctx, db))
	})
}

func GetDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	db := defDB

	if tdb, ok := contextx.FromTrans(ctx); ok {
		db = tdb
	}
	if contextx.FromRowLock(ctx) {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	return db.WithContext(ctx)
}

func wrapQueryOptions(db *gorm.DB, opts QueryOptions) *gorm.DB {
	if len(opts.SelectFields) > 0 {
		db = db.Select(opts.SelectFields)
	}
	if len(opts.OmitFields) > 0 {
		db = db.Omit(opts.OmitFields...)
	}
	if len(opts.OrderFields) > 0 {
		// 全部排序项非法时 ToSQLWith 返回空串，此时必须跳过 Order：
		// gorm 会把空串当成一列 Raw 排序，产出语法错误的 "ORDER BY "。
		if s := opts.OrderFields.ToSQLWith(db); s != "" {
			db = db.Order(s)
		}
	}
	return db
}

func WrapPageQuery(ctx context.Context, db *gorm.DB, pp PaginationParam, opts QueryOptions, out any) (*PaginationResult, error) {
	if pp.OnlyCount {
		var count int64
		err := db.Count(&count).Error
		if err != nil {
			return nil, err
		}
		return &PaginationResult{Total: count}, nil
	} else if !pp.Pagination {
		pageSize := pp.PageSize
		if pageSize > 0 {
			db = db.Limit(pageSize)
		}
		db = wrapQueryOptions(db, opts)
		err := db.Find(out).Error
		return nil, err
	}
	total, err := FindPage(ctx, db, pp, opts, out)
	if err != nil {
		return nil, err
	}
	return &PaginationResult{
		Items:    out,
		Total:    total,
		PageNum:  pp.PageNum,
		PageSize: pp.PageSize,
	}, nil
}

func FindPage(ctx context.Context, db *gorm.DB, pp PaginationParam, opts QueryOptions, out any) (int64, error) {
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return 0, err
	} else if count == 0 {
		return count, nil
	}

	pageNum, pageSize := pp.PageNum, pp.PageSize
	if pageNum > 0 && pageSize > 0 {
		db = db.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	} else if pageSize > 0 {
		db = db.Limit(pageSize)
	}

	db = wrapQueryOptions(db, opts)
	err = db.Find(out).Error
	return count, err
}

func FindOne(ctx context.Context, db *gorm.DB, opts QueryOptions, out any) (bool, error) {
	db = wrapQueryOptions(db, opts)
	result := db.Limit(1).Scan(out)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func Exists(ctx context.Context, db *gorm.DB) (bool, error) {
	var count int64
	result := db.Count(&count)
	if err := result.Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
