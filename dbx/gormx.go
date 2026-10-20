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
	// RawOrderFields 原样拼进 ORDER BY 的表达式（不做标识符白名单、不加方言引号）。
	//
	// 安全前提：表达式必须由服务端代码构造，**禁止**直接来自 HTTP query 参数。
	// 需要方言引号 / 保留字列名请继续用 OrderFields。
	//
	// 拼接顺序：RawOrderFields 全部在前，OrderFields 在后——典型用法是先按
	// 表达式分层（CASE ... END），再在同层内按普通列细分。
	RawOrderFields RawOrderFields
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
// 想要错误信号请先调 Validate；想知道具体丢了什么请用 DroppedFields。
// 全部项非法时返回空串——注意空串意味着 "本次查询不带 ORDER BY"，
// 调用方需自行跳过 db.Order("")。
//
// 注意：本方法的白名单只接受标识符，CASE ... END / NULLS LAST / 函数调用
// 这类表达式会被丢弃。服务端构造的表达式请走 QueryOptions.RawOrderFields。
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

// DroppedFields 返回会被 toSQL 丢弃的非法项（按原顺序）。
//
// 仅供调用方自查 / 打日志，不改变 toSQL 的丢弃行为。
//
// 典型用途：OrderFields 直接承载了用户的 ?order_by= 时，handler 应在进入
// 查询前用 Validate() 判 400；DroppedFields 则适合在中间层做一次"有东西被
// 静默丢弃"的告警日志，避免白名单生效后功能悄悄失效而无人察觉。
func (p OrderByParams) DroppedFields() []string {
	var dropped []string
	for _, v := range p {
		if !identPattern.MatchString(strings.TrimSpace(v.Field)) {
			dropped = append(dropped, v.Field)
		}
	}
	return dropped
}

// —— raw ORDER BY ——

// RawOrderField 一条 raw 排序表达式。
//
// Expr 原样拼进 SQL，**无任何校验**。这类表达式（CASE / NULLS LAST /
// 函数调用）本质上无法用标识符正则判定合法性，硬加过滤只会重演"静默丢弃"
// 事故，所以防注入责任明确归属调用方：
//
//   - 必须由服务端代码构造，禁止直接拼接 HTTP query 参数；
//   - 需要列名引号或保留字处理，请改用 OrderByParams（走方言引号）。
//
// Direction 是唯一走白名单的位——引号包不住关键字，非法方向仍降级为 ASC。
type RawOrderField struct {
	// Expr 排序表达式，如 "CASE status WHEN 'a' THEN 0 ELSE 1 END"。
	// 空串（含仅空白）时跳过该项。
	Expr string
	// Direction 排序方向，仍走 normalized() 归一（只允许 ASC/DESC，非法降级 ASC）——
	// 方向位无法用引号保护，这一层白名单必须保留。
	Direction Direction
}

// RawOrderFields 一组 raw 排序表达式。命名的 slice 类型只为挂方法；
// 构造时仍可直接用 []RawOrderField{...} 字面量（两者可互相赋值）。
type RawOrderFields []RawOrderField

// ToSQL 生成单条 raw 排序片段，Expr 为空白时返回空串（表示"跳过该项"）。
func (f RawOrderField) ToSQL() string {
	expr := strings.TrimSpace(f.Expr)
	if expr == "" {
		return ""
	}
	return expr + " " + string(f.Direction.normalized())
}

// toSQL 生成 raw ORDER BY 片段，丢弃 Expr 为空的项。
// 全部为空时返回空串（调用方应先判断再调 db.Order）。
func (p RawOrderFields) toSQL() string {
	if len(p) == 0 {
		return ""
	}
	parts := make([]string, 0, len(p))
	for _, f := range p {
		if s := f.ToSQL(); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, ", ")
}

// orderByClause 合并 raw 与标识符两条通道，raw 在前、OrderFields 在后。
// 两条都为空时返回空串，调用方据此跳过 db.Order()——GORM 会把 "" 当成
// 一列 Raw 排序，产出语法错误的 "ORDER BY "。
func orderByClause(raw RawOrderFields, fields OrderByParams, db *gorm.DB) string {
	var parts []string
	if s := raw.toSQL(); s != "" {
		parts = append(parts, s)
	}
	if s := fields.ToSQLWith(db); s != "" {
		parts = append(parts, s)
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
	if len(opts.OrderFields) > 0 || len(opts.RawOrderFields) > 0 {
		// 全部排序项被丢弃 / 表达式为空时返回空串，此时必须跳过 Order：
		// gorm 会把空串当成一列 Raw 排序，产出语法错误的 "ORDER BY "。
		if s := orderByClause(opts.RawOrderFields, opts.OrderFields, db); s != "" {
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
