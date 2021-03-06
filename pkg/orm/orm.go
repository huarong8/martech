package orm

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	DBTypeSQLite  dbType = "sqlite"
	DBTypeMySQL   dbType = "mysql"
	DBTypePostgre dbType = "postgres"
)

var (
	DefaultOrmOption = &Option{
		Type: DBTypeSQLite,
		DSN:  "sqlite.db",
	}
)

type dbType string

// Option 数据库配置
type Option struct {
	Debug bool   `json:"debug"`
	DSN   string `json:"dsn"`
	Type  dbType `json:"type"`
}

// New 创建数据库实例
func New(option *Option) (*gorm.DB, error) {
	if option == nil {
		option = DefaultOrmOption
	}

	dialect, err := getDialect(option.Type, option.DSN)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if option.Debug {
		db = db.Debug()
	}

	return db, nil
}

// LockDB 锁数据库db中tableName表的key行
func LockDB(db *gorm.DB, tableName string, key int) error {
	ty := dbType(db.Dialector.Name())
	name := getLockKey(ty, tableName, key)
	switch ty {
	case DBTypeMySQL:
		// docs: https://dev.mysql.com/doc/refman/8.0/en/locking-functions.html
		var res int
		db.Raw("SELECT GET_LOCK(?, 0) WHERE (SELECT IS_FREE_LOCK(?))=1;", name, name).Scan(&res)
		if res == 0 {
			return fmt.Errorf("%v has been locked", name)
		}

		return nil
	case DBTypePostgre:
		// docs: http://www.postgres.cn/docs/9.3/functions-admin.html
		var res bool
		db.Raw("SELECT pg_try_advisory_lock(?);", name).Scan(&res)
		if !res {
			return fmt.Errorf("%v has been locked", name)
		}

		return nil
	default:
		return fmt.Errorf("unsupported db type: %v", ty)
	}
}

// UnlockDB 解锁数据库db中tableName表的key行
func UnlockDB(db *gorm.DB, tableName string, key int) error {
	ty := dbType(db.Dialector.Name())
	name := getLockKey(ty, tableName, key)
	switch ty {
	case DBTypeMySQL:
		return db.Raw("SELECT RELEASE_LOCK(?);", name).Error
	case DBTypePostgre:
		return db.Raw("SELECT pg_advisory_unlock(?);", name).Error
	default:
		return fmt.Errorf("unsupported db type: %v", ty)
	}
}

func getDialect(ty dbType, dsn string) (gorm.Dialector, error) {
	switch ty {
	case DBTypeSQLite:
		return sqlite.Open(dsn), nil
	case DBTypeMySQL:
		return mysql.Open(dsn), nil
	case DBTypePostgre:
		return postgres.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported db type: %v", ty)
	}
}

func getLockKey(ty dbType, tableName string, key int) interface{} {
	name := fmt.Sprintf("%v-%v", tableName, key)
	switch ty {
	case DBTypeMySQL:
		return name
	case DBTypePostgre:
		md := md5.Sum([]byte(name))
		num, _ := strconv.ParseUint(hex.EncodeToString(md[:])[:6], 16, 63)
		return num
	default:
		return nil
	}
}

func insertOrRecover(db *gorm.DB, item interface{}) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where(item).Take(item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return tx.Create(item).Error
			}
			return err
		}
		return tx.Model(item).Update("delete_at", nil).Error
	})
}
