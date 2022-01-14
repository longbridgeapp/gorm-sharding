package gormsharding

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/longbridgeapp/gorm-sharding/keygen"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/hints"
)

type Order struct {
	ID      int64 `gorm:"primarykey"`
	UserID  int64
	Product string
}

type Category struct {
	ID   int64 `gorm:"primarykey"`
	Name string
}

func databaseURL() string {
	databaseURL := os.Getenv("DATABASE_URL")
	if len(databaseURL) == 0 {
		databaseURL = "postgres://localhost:5432/sharding-test?sslmode=disable"
	}
	return databaseURL
}

var (
	dbConfig = postgres.Config{
		DSN:                  databaseURL(),
		PreferSimpleProtocol: true,
	}
	db, _ = gorm.Open(postgres.New(dbConfig), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	sharding = Register(map[string]Resolver{
		"orders": {
			EnableFullTable: true,
			ShardingColumn:  "user_id",
			ShardingAlgorithm: func(value interface{}) (suffix string, err error) {
				userId := 0
				switch value := value.(type) {
				case int:
					userId = value
				case int64:
					userId = int(value)
				case string:
					userId, err = strconv.Atoi(value)
					if err != nil {
						return "", err
					}
				default:
					return "", err
				}
				return fmt.Sprintf("_%03d", userId%8), nil
			},
			ShardingAlgorithmByPrimaryKey: func(id int64) (suffix string) {
				return fmt.Sprintf("_%03d", keygen.TableIdx(id))
			},
			PrimaryKeyGenerate: func(tableIdx int64) int64 {
				return keygen.Next(tableIdx)
			},
		},
	})
)

var tables = []string{"orders", "categories"}

func init() {
	dropTables()
	err := db.AutoMigrate(&Order{}, &Category{})
	if err != nil {
		panic(err)
	}
	db.Use(&sharding)
}

func dropTables() {
	for _, tableName := range tables {
		db.Exec("DROP TABLE IF EXISTS " + tableName)
	}
}

func TestInsert(t *testing.T) {
	tx := db.Create(&Order{ID: 100, UserID: 100, Product: "iPhone"})
	assertQueryResult(t, `INSERT INTO "orders_004" ("user_id", "product", "id") VALUES ($1, $2, $3) RETURNING "id"`, tx)
}

func TestFillID(t *testing.T) {
	db.Create(&Order{UserID: 100, Product: "iPhone"})
	lastQuery := sharding.LastQuery()
	assert.Equal(t, `INSERT INTO "orders_004" ("user_id", "product", "id") VALUES`, lastQuery[0:57])
}

func TestSelect1(t *testing.T) {
	tx := db.Model(&Order{}).Where("user_id", 101).Where("id", keygen.Next(24)).Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_024" WHERE "user_id" = $1 AND "id" = $2`, tx)
}

func TestSelect2(t *testing.T) {
	tx := db.Model(&Order{}).Where("id", keygen.Next(24)).Where("user_id", 101).Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_005" WHERE "id" = $1 AND "user_id" = $2`, tx)
}

func TestSelect3(t *testing.T) {
	tx := db.Model(&Order{}).Where("id", keygen.Next(24)).Where("user_id = 101").Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_005" WHERE "id" = $1 AND "user_id" = 101`, tx)
}

func TestSelect4(t *testing.T) {
	tx := db.Model(&Order{}).Where("product", "iPad").Where("user_id", 100).Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_004" WHERE "product" = $1 AND "user_id" = $2`, tx)
}

func TestSelect5(t *testing.T) {
	tx := db.Model(&Order{}).Where("user_id = 101").Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_005" WHERE "user_id" = 101`, tx)
}

func TestSelect6(t *testing.T) {
	tx := db.Model(&Order{}).Where("id", keygen.Next(24)).Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_024" WHERE "id" = $1`, tx)
}

func TestSelect7(t *testing.T) {
	tx := db.Model(&Order{}).Where("user_id", 101).Where("id > ?", keygen.Next(24)).Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_005" WHERE "user_id" = $1 AND "id" > $2`, tx)
}

func TestSelect8(t *testing.T) {
	tx := db.Model(&Order{}).Where("id > ?", keygen.Next(24)).Where("user_id", 101).Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_005" WHERE "id" > $1 AND "user_id" = $2`, tx)
}

func TestSelect9(t *testing.T) {
	tx := db.Model(&Order{}).Where("user_id = 101").First(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "orders_005" WHERE "user_id" = 101 ORDER BY "orders_005"."id" LIMIT 1`, tx)
}

func TestSelect10(t *testing.T) {
	tx := db.Clauses(hints.Comment("select", "nosharding")).Model(&Order{}).Find(&[]Order{})
	assertQueryResult(t, `SELECT /* nosharding */ * FROM "orders"`, tx)
}

func TestSelect11(t *testing.T) {
	tx := db.Clauses(hints.Comment("select", "nosharding")).Model(&Order{}).Where("user_id", 101).Find(&[]Order{})
	assertQueryResult(t, `SELECT /* nosharding */ * FROM "orders" WHERE "user_id" = $1`, tx)
}

func TestSelect12(t *testing.T) {
	tx := db.Raw(`SELECT * FROM "public"."orders" WHERE "user_id" = 101`).Find(&[]Order{})
	assertQueryResult(t, `SELECT * FROM "public"."orders" WHERE "user_id" = 101`, tx)
}

func TestSelect13(t *testing.T) {
	tx := db.Raw("SELECT 1").Find(&[]Order{})
	assertQueryResult(t, `SELECT 1`, tx)
}

func TestUpdate(t *testing.T) {
	tx := db.Model(&Order{}).Where("user_id = ?", 100).Update("product", "new title")
	assertQueryResult(t, `UPDATE "orders_004" SET "product" = $1 WHERE "user_id" = $2`, tx)
}

func TestDelete(t *testing.T) {
	tx := db.Where("user_id = ?", 100).Delete(&Order{})
	assertQueryResult(t, `DELETE FROM "orders_004" WHERE "user_id" = $1`, tx)
}

func TestInsertMissingShardingKey(t *testing.T) {
	err := db.Exec(`INSERT INTO "orders" ("id", "product") VALUES(1, 'iPad')`).Error
	assert.Equal(t, ErrMissingShardingKey, err)
}

func TestSelectMissingShardingKey(t *testing.T) {
	err := db.Exec(`SELECT * FROM "orders" WHERE "product" = 'iPad'`).Error
	assert.Equal(t, ErrMissingShardingKey, err)
}

func TestSelectNoSharding(t *testing.T) {
	err := db.Exec(`SELECT /* nosharding */ * FROM "orders" WHERE "product" = 'iPad'`).Error
	assert.Equal(t, nil, err)
}

func TestNoEq(t *testing.T) {
	err := db.Model(&Order{}).Where("user_id <> ?", 101).Find([]Order{}).Error
	assert.Equal(t, ErrMissingShardingKey, err)
}

func TestShardingKeyOK(t *testing.T) {
	err := db.Model(&Order{}).Where("user_id = ? and id > ?", 101, int64(100)).Find(&[]Order{}).Error
	assert.Equal(t, nil, err)
}

func TestShardingKeyNotOK(t *testing.T) {
	err := db.Model(&Order{}).Where("user_id > ? and id > ?", 101, int64(100)).Find(&[]Order{}).Error
	assert.Equal(t, ErrMissingShardingKey, err)
}

func TestShardingIdOK(t *testing.T) {
	err := db.Model(&Order{}).Where("id = ? and user_id > ?", int64(101), 100).Find(&[]Order{}).Error
	assert.Equal(t, nil, err)
}

func TestNoSharding(t *testing.T) {
	categories := []Category{}
	tx := db.Model(&Category{}).Where("id = ?", 1).Find(&categories)
	assertQueryResult(t, `SELECT * FROM "categories" WHERE id = $1`, tx)
}

func assertQueryResult(t *testing.T, query string, tx *gorm.DB) {
	t.Helper()
	assert.Equal(t, query, sharding.LastQuery())
}
