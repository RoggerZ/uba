//MySql引擎层
package db

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// sqlx 全局变量
var Sqlx *sqlx.DB
var ClickHouseSqlx *sqlx.DB

// 用squirrel生成sql语句
var SqlBuilder = squirrel.StatementBuilder

type Eq = squirrel.Eq
type Or = squirrel.Or
type And = squirrel.And

type NotEq = squirrel.NotEq
type Gt = squirrel.Gt
type Lt = squirrel.Lt
type GtOrEq = squirrel.GtOrEq
type LtOrEq = squirrel.LtOrEq

type Like = squirrel.Like
type Gte = squirrel.GtOrEq
type Lte = squirrel.LtOrEq
type SelectBuilder = squirrel.SelectBuilder
type InsertBuilder = squirrel.InsertBuilder
type UpdateBuilder = squirrel.UpdateBuilder

// NewMySQL 创建一个连接MySQL的实体池
func NewSQLX(driverName, dbSource string, maxOpenConns, maxIdleConns int) (db *sqlx.DB, err error) {
	return NewSQLXWithOptions(driverName, dbSource, maxOpenConns, maxIdleConns, DBHealthOptions{})
}

func NewSQLXWithOptions(driverName, dbSource string, maxOpenConns, maxIdleConns int, healthOptions DBHealthOptions) (db *sqlx.DB, err error) {
	db, err = sqlx.Open(driverName, dbSource)
	if err != nil {
		return
	}
	if maxOpenConns > 0 {
		db.SetMaxOpenConns(maxOpenConns)
	}

	if maxIdleConns > 0 {
		db.SetMaxIdleConns(maxIdleConns)
	}
	err = db.Ping()
	if err != nil {
		return
	}
	monitor := newDBHealthMonitor(db, driverName, maxIdleConns, healthOptions)
	monitor.start()
	return
}

// 创建分页查询
func CreatePage(page, limit uint64) uint64 {
	tmp := (page - 1) * limit
	return tmp
}

// 创建模糊查询
func CreateLike(column string) string {
	return fmt.Sprint("%", column, "%")
}
