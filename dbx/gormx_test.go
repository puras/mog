package dbx

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/puras/mog/contextx"
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

		if err := tx.Create(&testItem{ID: "c", Name: "c"}).Error; err == nil {
			t.Error("ctx 已取消，裸取的事务仍写入成功")
		} else if !errors.Is(err, context.Canceled) {
			t.Errorf("期望 context.Canceled，得到 %v", err)
		} else {
			sawCancelErr = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Exec: %v", err)
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
