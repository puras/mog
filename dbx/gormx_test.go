package dbx

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/puras/mog/config"
	"github.com/puras/mog/contextx"

	"gorm.io/gorm"
)

type testItem struct {
	ID   string `gorm:"primaryKey;size:64"`
	Name string `gorm:"size:64"`
}

func newTestDB(t *testing.T) *Trans {
	t.Helper()
	db, err := NewDB(Config{DBType: "sqlite3", DSN: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&testItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &Trans{DB: db}
}

func countItems(t *testing.T, tr *Trans, names ...string) int64 {
	t.Helper()
	var n int64
	if err := tr.DB.Model(&testItem{}).Where("name IN ?", names).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// 事务注入 ctx 的 tx 必须继承调用方 ctx —— 这正是 GORM Begin() 从 db.Statement.Context
// 取事务 ctx 的行为。缺了 Trans.Exec 里的 WithContext，root DB 的 Statement.Context
// 是 context.Background()，事务会失去 deadline / cancel。
func TestTransExec_TxInheritsCallerContext(t *testing.T) {
	tr := newTestDB(t)

	type ctxKey struct{}
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), ctxKey{}, "v"), time.Minute)
	defer cancel()

	err := tr.Exec(ctx, func(c context.Context) error {
		tx, ok := contextx.FromTrans(c)
		if !ok {
			t.Fatal("ctx 中未找到事务")
		}
		if tx.Statement.Context != ctx {
			t.Errorf("事务 ctx 未继承调用方 ctx：got %v", tx.Statement.Context)
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("测试前提错误：调用方 ctx 应带 deadline")
		}
		if _, ok := tx.Statement.Context.Deadline(); !ok {
			t.Error("事务 ctx 无 deadline，调用方 deadline 未生效")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Exec: %v", err)
	}
}

// ctx 取消后，事务内裸取的 tx 也必须立即失败并回滚。
//
// 这里刻意不用 GetDB(ctx, ...)（它自身带 WithContext），直接用 FromTrans 拿到的
// tx 调 Create —— 与 billing 里 9 处 contextx.FromTrans(ctx) 裸取点的用法一致。
// 修复前该用例会看到 c 行写成功、事务不报错。
func TestTransExec_RollsBackAfterContextCancel(t *testing.T) {
	tr := newTestDB(t)
	if err := tr.DB.Create(&testItem{ID: "seed", Name: "seed"}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var sawCancelErr bool
	var createErr error
	err := tr.Exec(ctx, func(c context.Context) error {
		tx, ok := contextx.FromTrans(c)
		if !ok {
			t.Fatal("ctx 中未找到事务")
		}
		if err := tx.Create(&testItem{ID: "b", Name: "b"}).Error; err != nil {
			return err
		}

		cancel()
		time.Sleep(100 * time.Millisecond)

		createErr = tx.Create(&testItem{ID: "c", Name: "c"}).Error
		if createErr == nil {
			t.Error("ctx 已取消，裸取的事务仍写入成功")
		} else if !errors.Is(createErr, context.Canceled) {
			t.Errorf("期望 context.Canceled，得到 %v", createErr)
		} else {
			sawCancelErr = true
		}
		// 必须把 ctx 取消后的错误向上抛，让 GORM 走 Rollback 而不是 Commit。
		// 修复前 WithContext(ctx) 不存在，tx 走 background，c 行照样写成功、Exec 返回 nil。
		return createErr
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Exec: 期望 context.Canceled，得到 %v", err)
	}
	if !sawCancelErr {
		t.Error("未观察到 context.Canceled")
	}
	if n := countItems(t, tr, "b", "c"); n != 0 {
		t.Fatalf("事务未回滚，残留 %d 行", n)
	}
}

// 已存在事务时嵌套 Exec 应直接复用，不再开新事务。
func TestTransExec_ReusesExistingTrans(t *testing.T) {
	tr := newTestDB(t)

	outer, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := tr.Exec(outer, func(c context.Context) error {
		outerTx, _ := contextx.FromTrans(c)
		return tr.Exec(c, func(inner context.Context) error {
			innerTx, _ := contextx.FromTrans(inner)
			if innerTx != outerTx {
				t.Error("嵌套 Exec 开了新事务")
			}
			return innerTx.Create(&testItem{ID: "n", Name: "n"}).Error
		})
	}); err != nil {
		t.Fatalf("Exec: %v", err)
	}
	if n := countItems(t, tr, "n"); n != 1 {
		t.Fatalf("嵌套写入未提交，得到 %d 行", n)
	}
}

// —— OrderByParams.ToSQL ——

func TestOrderByParams_ToSQL(t *testing.T) {
	tests := []struct {
		name string
		in   OrderByParams
		want string
	}{
		{"空切片", OrderByParams{}, ""},
		{"单条", OrderByParams{{Field: "created_at", Direction: DESC}}, "created_at DESC"},
		{"多条按序拼接", OrderByParams{
			{Field: "created_at", Direction: DESC},
			{Field: "id", Direction: ASC},
		}, "created_at DESC, id ASC"},
		{"空方向补 ASC", OrderByParams{{Field: "name"}}, "name ASC"},
		{"方向大小写归一", OrderByParams{{Field: "name", Direction: "desc"}}, "name DESC"},
		{"字段前后空白裁剪", OrderByParams{{Field: "  name  ", Direction: " Desc "}}, "name DESC"},
		{"table.column 形式", OrderByParams{{Field: "user.name", Direction: ASC}}, "user.name ASC"},
		// 非法项被丢弃，保留下来的项顺序不变
		{"丢弃非法字段", OrderByParams{
			{Field: "id; DROP TABLE users--", Direction: ASC},
			{Field: "name", Direction: ASC},
		}, "name ASC"},
		{"全部非法返回空串", OrderByParams{{Field: "1=1", Direction: ASC}}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.ToSQL(); got != tt.want {
				t.Errorf("ToSQL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ToSQL 是 query 参数的唯一入口，必须保证注入载荷不会出现在输出里。
// 这里用一组真实攻击串覆盖：注释截断、恒真条件、UNION 探测、函数调用。
func TestOrderByParams_ToSQL_RejectsInjection(t *testing.T) {
	payloads := []string{
		"id; DROP TABLE users--",
		"id = 1 OR 1=1",
		"id, (SELECT password FROM users)",
		"id`; DROP TABLE users; --",
		"id/**/DESC",
		"id ASC, SLEEP(5)",
		`id" , (SELECT 1) --`,
		"1",
		".",
		"a..b",
		"a.b.c",
		"-id",
		"id DESC --",
	}

	for _, payload := range payloads {
		out := OrderByParams{{Field: payload, Direction: ASC}}.ToSQL()
		if out != "" {
			t.Errorf("载荷 %q 泄漏进 SQL: %q", payload, out)
		}
	}

	// 方向位同样不能泄漏：引号保护不了关键字，必须由 normalized 白名单兜底。
	out := OrderByParams{{Field: "id", Direction: "; DROP TABLE users--"}}.ToSQL()
	if out != "id ASC" {
		t.Errorf("方向位未归一: %q", out)
	}
}

func TestOrderByParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		in      OrderByParams
		wantErr bool
	}{
		{"空切片合法", OrderByParams{}, false},
		{"常规字段", OrderByParams{{Field: "created_at", Direction: DESC}}, false},
		{"空方向合法", OrderByParams{{Field: "created_at"}}, false},
		{"大小写方向合法", OrderByParams{{Field: "id", Direction: "desc"}}, false},
		{"table.column 合法", OrderByParams{{Field: "user.name", Direction: ASC}}, false},
		{"注入字段报错", OrderByParams{{Field: "id; DROP TABLE users--"}}, true},
		{"非法方向报错", OrderByParams{{Field: "id", Direction: "asc, (select 1)"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ToSQLWith 按方言加引号：sqlite 驱动走反引号，PG 走双引号，
// 且 table.column 会被逐段包裹。
func TestOrderByParams_ToSQLWith(t *testing.T) {
	tr := newTestDB(t)

	got := OrderByParams{{Field: "test_item.name", Direction: DESC}}.ToSQLWith(tr.DB)
	if want := "`test_item`.`name` DESC"; got != want {
		t.Errorf("ToSQLWith(sqlite) = %q, want %q", got, want)
	}

	// nil db 退化为不引用的安全版本
	if got := (OrderByParams{{Field: "name", Direction: ASC}}).ToSQLWith(nil); got != "name ASC" {
		t.Errorf("ToSQLWith(nil) = %q, want %q", got, "name ASC")
	}

	// 保留字列名加引号后才可用
	if got := (OrderByParams{{Field: "order", Direction: ASC}}).ToSQLWith(tr.DB); got != "`order` ASC" {
		t.Errorf("保留字未被引用: %q", got)
	}
}

// 端到端：QueryOptions.OrderFields 真的能驱动查询，且注入载荷退化为无排序
// （而不是语法错误或全表无序执行）。
func TestWrapQueryOptions_OrderFields(t *testing.T) {
	tr := newTestDB(t)
	for _, n := range []string{"c", "a", "b"} {
		if err := tr.DB.Create(&testItem{ID: n, Name: n}).Error; err != nil {
			t.Fatal(n, err)
		}
	}

	// 正常排序
	db := tr.DB.Model(&testItem{})
	var got []testItem
	if err := wrapQueryOptions(db, QueryOptions{
		OrderFields: OrderByParams{{Field: "name", Direction: ASC}},
	}).Find(&got).Error; err != nil {
		t.Fatalf("升序查询: %v", err)
	}
	if len(got) != 3 || got[0].Name != "a" || got[2].Name != "c" {
		t.Fatalf("升序结果不对: %+v", got)
	}

	// 注入载荷 → 全部项非法 → 跳过 Order，查询仍能正常执行
	db2 := tr.DB.Model(&testItem{})
	got = nil
	if err := wrapQueryOptions(db2, QueryOptions{
		OrderFields: OrderByParams{{Field: "id; DROP TABLE users--"}},
	}).Find(&got).Error; err != nil {
		t.Fatalf("非法排序项应被跳过而非让查询失败: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("期望返回 3 行，得到 %d", len(got))
	}
}

// InitDB 返回的 cleanup 必须真正关掉连接池。
//
// 修复前这里是 `if err != nil { sqlDB.Close() }`，判断写反了：而 db.DB()
// 只在 ConnPool 不是 *sql.DB 时才报错，正常路径恒为 (db, nil) —— 于是
// cleanup 永远不关池（静默泄漏，且 injector 的初始化失败回滚路径也跟着失效），
// 反倒是唯一会执行到 Close() 的分支里 sqlDB 为 nil，直接 panic。
func TestInitDB_CleanupClosesPool(t *testing.T) {
	old := config.C.Storage.DataBase
	defer func() { config.C.Storage.DataBase = old }()

	// MaxIdleConns 必须 > 0，否则 database/sql 会立刻回收空闲连接，测不出差异
	config.C.Storage.DataBase.Enable = true
	config.C.Storage.DataBase.Type = "sqlite3"
	config.C.Storage.DataBase.DSN = filepath.Join(t.TempDir(), "cleanup.db")
	config.C.Storage.DataBase.MaxIdleConns = 5
	config.C.Storage.DataBase.MaxOpenConns = 5

	db, cleanup, err := InitDB(context.Background())
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	if cleanup == nil {
		t.Fatal("cleanup 为 nil，无法验证")
	}

	// 建一张表，确保确实持有空闲连接
	if err := db.AutoMigrate(&testItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if before := sqlDB.Stats().OpenConnections; before == 0 {
		t.Fatalf("测试前提不成立：cleanup 前应有空闲连接，实际 %d", before)
	}

	cleanup()

	if after := sqlDB.Stats().OpenConnections; after != 0 {
		t.Errorf("cleanup 未关闭连接池：OpenConnections = %d, want 0", after)
	}
}

// Enable=false 时不应建库，cleanup 也保持 nil。
func TestInitDB_Disabled(t *testing.T) {
	old := config.C.Storage.DataBase
	defer func() { config.C.Storage.DataBase = old }()

	config.C.Storage.DataBase.Enable = false

	db, cleanup, err := InitDB(context.Background())
	if err != nil {
		t.Fatalf("Enable=false 不该报错: %v", err)
	}
	if db != nil || cleanup != nil {
		t.Errorf("Enable=false 时应返回 (nil, nil, nil)，实际 db=%v, cleanup 非 nil=%v",
			db, cleanup != nil)
	}
}

// —— RawOrderField / RawOrderFields ——

func TestRawOrderField_ToSQL(t *testing.T) {
	tests := []struct {
		name string
		in   RawOrderField
		want string
	}{
		{"普通表达式", RawOrderField{Expr: "priority", Direction: DESC}, "priority DESC"},
		{"CASE 表达式", RawOrderField{
			Expr:      "CASE status WHEN 'enabled' THEN 0 ELSE 1 END",
			Direction: ASC,
		}, "CASE status WHEN 'enabled' THEN 0 ELSE 1 END ASC"},
		{"空 Expr 跳过", RawOrderField{Expr: "", Direction: ASC}, ""},
		{"仅空白 Expr 跳过", RawOrderField{Expr: "   \t\n ", Direction: DESC}, ""},
		{"表达式前后空白裁剪", RawOrderField{Expr: "  priority  ", Direction: ASC}, "priority ASC"},
		{"空方向补 ASC", RawOrderField{Expr: "priority"}, "priority ASC"},
		{"小写方向归一", RawOrderField{Expr: "priority", Direction: "desc"}, "priority DESC"},
		// 方向位不走引号，必须靠归一白名单挡住
		{"非法方向降级 ASC", RawOrderField{Expr: "priority", Direction: "; DROP TABLE t--"}, "priority ASC"},
		// 单引号是 SQL 字面量的一部分，必须原样保留，不能做任何转义或改写
		{"单引号原样保留", RawOrderField{
			Expr:      "CASE WHEN name = 'O''Brien' THEN 1 ELSE 0 END",
			Direction: DESC,
		}, "CASE WHEN name = 'O''Brien' THEN 1 ELSE 0 END DESC"},
		{"函数调用原样保留", RawOrderField{Expr: "LENGTH(name)", Direction: DESC}, "LENGTH(name) DESC"},
		{"NULLS LAST 原样保留", RawOrderField{Expr: "created_at NULLS LAST", Direction: ASC}, "created_at NULLS LAST ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.ToSQL(); got != tt.want {
				t.Errorf("ToSQL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRawOrderFields_SkipsEmpty(t *testing.T) {
	in := RawOrderFields{
		{Expr: "priority", Direction: DESC},
		{Expr: "", Direction: ASC},
		{Expr: "   ", Direction: ASC},
		{Expr: "weight", Direction: DESC},
	}
	if got, want := in.toSQL(), "priority DESC, weight DESC"; got != want {
		t.Errorf("toSQL() = %q, want %q", got, want)
	}

	if got := (RawOrderFields{}).toSQL(); got != "" {
		t.Errorf("空切片应返回空串，得到 %q", got)
	}
	if got := (RawOrderFields{{Expr: "  "}}).toSQL(); got != "" {
		t.Errorf("全空白应返回空串，得到 %q", got)
	}
}

// raw 与 OrderFields 合并成一次 Order，raw 在前。
func TestWrapQueryOptions_RawOrderFields(t *testing.T) {
	tr := newTestDB(t)

	// 不引号化的 raw 通道
	if got, want := orderByClause(
		RawOrderFields{{Expr: "CASE status WHEN 'enabled' THEN 0 ELSE 1 END", Direction: ASC}},
		nil, tr.DB,
	), "CASE status WHEN 'enabled' THEN 0 ELSE 1 END ASC"; got != want {
		t.Errorf("raw only = %q, want %q", got, want)
	}

	// raw 在前、OrderFields 在后，后者按方言加引号
	got := orderByClause(
		RawOrderFields{{Expr: "CASE status WHEN 'enabled' THEN 0 ELSE 1 END", Direction: ASC}},
		OrderByParams{{Field: "priority", Direction: DESC}, {Field: "created_at", Direction: DESC}},
		tr.DB,
	)
	want := "CASE status WHEN 'enabled' THEN 0 ELSE 1 END ASC, `priority` DESC, `created_at` DESC"
	if got != want {
		t.Errorf("merged = %q\n     want %q", got, want)
	}

	// 两边都空 → 空串，wrapQueryOptions 据此跳过 Order
	if got := orderByClause(RawOrderFields{{Expr: " "}}, OrderByParams{{Field: "1=1"}}, tr.DB); got != "" {
		t.Errorf("全部无效时应返回空串，得到 %q", got)
	}

	// 断言真的只调了一次 Order：用 DryRun 走完整 SQL 渲染，ORDER BY 只该出现一次。
	// 若把 raw 与 OrderFields 分成两次 db.Order()，GORM 会用逗号拼接成
	// "ORDER BY a, ORDER BY b"，这里就会暴露。
	dry := tr.DB.Session(&gorm.Session{DryRun: true}).Model(&testItem{})
	var dryOut []testItem
	wrapQueryOptions(dry, QueryOptions{
		RawOrderFields: RawOrderFields{{Expr: "name", Direction: ASC}},
		OrderFields:    OrderByParams{{Field: "id", Direction: DESC}},
	}).Find(&dryOut)

	sql := dry.Statement.SQL.String()
	if n := strings.Count(sql, "ORDER BY"); n != 1 {
		t.Errorf("ORDER BY 出现 %d 次，want 1: %q", n, sql)
	}
	wantSQL := "SELECT * FROM `test_item` ORDER BY name ASC, `id` DESC"
	if sql != wantSQL {
		t.Errorf("SQL = %q\n  want %q", sql, wantSQL)
	}
}

// 端到端：mogo admin 渠道列表的状态分层排序。
// 修复前 CASE 项被 identPattern 静默丢弃，auto_disabled 会排到最前面。
func TestWrapQueryOptions_RawOrderFields_AppliesCASE(t *testing.T) {
	tr := newTestDB(t)
	if err := tr.DB.AutoMigrate(&statusItem{}); err != nil {
		t.Fatal(err)
	}

	// 故意按"错误顺序"插入，确保结果不是靠插入顺序蒙对
	rows := []statusItem{
		{ID: "1", Status: "disabled", Name: "d"},
		{ID: "2", Status: "auto_disabled", Name: "ad"},
		{ID: "3", Status: "enabled", Name: "e"},
		{ID: "4", Status: "auto_disabled", Name: "ad2"},
	}
	for i := range rows {
		if err := tr.DB.Create(&rows[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	var got []statusItem
	db := tr.DB.Model(&statusItem{})
	if err := wrapQueryOptions(db, QueryOptions{
		RawOrderFields: RawOrderFields{{
			Expr:      "CASE status WHEN 'enabled' THEN 0 WHEN 'auto_disabled' THEN 1 WHEN 'disabled' THEN 2 ELSE 3 END",
			Direction: ASC,
		}},
	}).Find(&got).Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	// enabled 必须全部在最前，其次 auto_disabled，最后 disabled
	wantOrder := []string{"enabled", "auto_disabled", "auto_disabled", "disabled"}
	if len(got) != len(wantOrder) {
		t.Fatalf("返回 %d 行，want %d", len(got), len(wantOrder))
	}
	for i, want := range wantOrder {
		if got[i].Status != want {
			statuses := make([]string, len(got))
			for j := range got {
				statuses[j] = got[j].Status
			}
			t.Fatalf("第 %d 行状态 = %q, want %q\n实际顺序: %v", i, got[i].Status, want, statuses)
		}
	}
}

// 对照组：同样的数据走 OrderFields 会被整条丢弃（证明白名单缺口真实存在，
// 也是本次新增 RawOrderFields 的动机）。
func TestWrapQueryOptions_OrderFieldsDropsCASE(t *testing.T) {
	tr := newTestDB(t)
	if err := tr.DB.AutoMigrate(&statusItem{}); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"disabled", "enabled"} {
		if err := tr.DB.Create(&statusItem{ID: s, Status: s}).Error; err != nil {
			t.Fatal(err)
		}
	}

	caseExpr := "CASE status WHEN 'enabled' THEN 0 ELSE 1 END"
	if dropped := (OrderByParams{{Field: caseExpr, Direction: ASC}}).DroppedFields(); len(dropped) != 1 || dropped[0] != caseExpr {
		t.Fatalf("DroppedFields() = %v, want [%q]", dropped, caseExpr)
	}

	db := tr.DB.Model(&statusItem{})
	if s := orderByClause(nil, OrderByParams{{Field: caseExpr, Direction: ASC}}, db); s != "" {
		t.Fatalf("CASE 表达式本应被 OrderFields 丢弃，实际生成 %q", s)
	}
}

// DroppedFields 保序报告非法项，合法项不出现。
func TestOrderByParams_DroppedFields(t *testing.T) {
	p := OrderByParams{
		{Field: "name", Direction: ASC},
		{Field: "CASE status WHEN 'a' THEN 0 END", Direction: ASC},
		{Field: "id", Direction: DESC},
		{Field: "1=1", Direction: ASC},
	}
	got := p.DroppedFields()
	want := []string{"CASE status WHEN 'a' THEN 0 END", "1=1"}
	if len(got) != len(want) {
		t.Fatalf("DroppedFields() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("DroppedFields() = %v, want %v", got, want)
		}
	}

	if got := (OrderByParams{{Field: "name"}}).DroppedFields(); got != nil {
		t.Errorf("全部合法时应返回 nil，得到 %v", got)
	}
}

// statusItem 带 status 列，用于验证 CASE 分层排序（testItem 没有该字段）。
type statusItem struct {
	ID     string `gorm:"primaryKey;size:64"`
	Status string `gorm:"size:32"`
	Name   string `gorm:"size:64"`
}
