package crud

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/puras/mog/dbx"
	"gorm.io/gorm"
)

// crudItem 覆盖 CrudRepo 用到的基础列：id / created_at / updated_at /
// deleted / deleted_at（Delete 走逻辑删除，直接写这几列）。
type crudItem struct {
	ID        string `gorm:"primaryKey;size:64"`
	Name      string `gorm:"size:64"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Deleted   bool
	DeletedAt time.Time
}

func (m crudItem) GetID() string { return m.ID }

func newRepo(t *testing.T) (*CrudRepo[crudItem], *gorm.DB) {
	t.Helper()
	db, err := dbx.NewDB(dbx.Config{
		DBType: "sqlite3",
		DSN:    filepath.Join(t.TempDir(), "crud.db"),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&crudItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewRepo[crudItem](db, nil), db
}

func seed(t *testing.T, db *gorm.DB, ids ...string) {
	t.Helper()
	for _, id := range ids {
		if err := db.Create(&crudItem{ID: id, Name: id + "-name"}).Error; err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}
}

func countAll(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&crudItem{}).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// Get 必须真的按 id 过滤。此前 dbx.Where 丢弃返回值，在 Session/根 DB 等
// clone 链上条件会静默失效，Get 会退化成「取第一行」。
func TestCrudRepo_GetFiltersById(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, "a", "b")

	got, err := repo.Get(context.Background(), "b")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("Get 返回 nil，期望命中 b")
	}
	if got.ID != "b" {
		t.Errorf("Get 命中 %q，want b（id 条件未生效）", got.ID)
	}
}

// 逻辑删除的行不应被 Get 命中。
func TestCrudRepo_GetIgnoresDeleted(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, "a")
	if err := db.Model(&crudItem{}).Where("id = ?", "a").
		Updates(map[string]any{"deleted": true}).Error; err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get(context.Background(), "a")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != nil {
		t.Errorf("已逻辑删除的行不该被 Get 命中：%+v", got)
	}
}

// Update 只能影响目标行。
func TestCrudRepo_UpdateFiltersById(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, "a", "b")

	// 与 CrudBiz.Update 的真实路径一致：item 由 Get 取出（带正确 ID），
	// 再 FillTo 后回写。Repo.Update 用 Select("*") 会连 id 一起更新，
	// 传空 ID 的 item 会把主键覆盖成空串。
	if err := repo.Update(context.Background(), "a", &crudItem{ID: "a", Name: "renamed"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	var rows []crudItem
	if err := db.Order("id").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("行数 = %d, want 2", len(rows))
	}
	if rows[0].ID != "a" || rows[0].Name != "renamed" {
		t.Errorf("a 行未被正确更新: %+v", rows[0])
	}
	if rows[1].ID != "b" || rows[1].Name != "b-name" {
		t.Errorf("b 行被误改: %+v（id 条件未生效）", rows[1])
	}
}

// Delete 是逻辑删除：只置位目标行，不删数据。
func TestCrudRepo_DeleteIsLogical(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, "a", "b")

	if err := repo.Delete(context.Background(), "a"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if n := countAll(t, db); n != 2 {
		t.Fatalf("逻辑删除不应减少行数，实际 %d", n)
	}
	var a, b crudItem
	if err := db.Where("id = ?", "a").First(&a).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Where("id = ?", "b").First(&b).Error; err != nil {
		t.Fatal(err)
	}
	if !a.Deleted {
		t.Error("a 未被标记删除")
	}
	if b.Deleted {
		t.Error("b 被误标记删除（id 条件未生效）")
	}
}

// HardDelete 是物理删除：只移除目标行。
func TestCrudRepo_HardDeleteFiltersById(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, "a", "b")

	if err := repo.HardDelete(context.Background(), "a"); err != nil {
		t.Fatalf("HardDelete: %v", err)
	}

	if n := countAll(t, db); n != 1 {
		t.Fatalf("物理删除后行数 = %d, want 1", n)
	}
	var left crudItem
	if err := db.First(&left).Error; err != nil {
		t.Fatal(err)
	}
	if left.ID != "b" {
		t.Errorf("留下的是 %q，want b（id 条件未生效）", left.ID)
	}
}

// Query 应过滤逻辑删除的行。
func TestCrudRepo_QueryExcludesDeleted(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, "a", "b")
	if err := db.Model(&crudItem{}).Where("id = ?", "a").
		Updates(map[string]any{"deleted": true}).Error; err != nil {
		t.Fatal(err)
	}

	pr, err := repo.Query(context.Background(), PageParams{}, dbx.PaginationParam{}, dbx.QueryOptions{})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	items, ok := pr.Items.([]*crudItem)
	if !ok {
		t.Fatalf("Items 类型 = %T", pr.Items)
	}
	if len(items) != 1 || items[0].ID != "b" {
		t.Errorf("Query 结果 = %+v，期望仅 b", items)
	}
}
