package db_repo

import (
	"fmt"
	"gin-api/internal/pkg/config"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var _ Repo = (*dbRepo)(nil)

type Repo interface {
	i()
	GetDbR() *gorm.DB
	GetDbW() *gorm.DB
	DbRClose() error
	DbWClose() error
}

type dbRepo struct {
	DbR *gorm.DB
	DbW *gorm.DB
}

func New() (Repo, error) {
	cfg := config.GetConfig()

	dbr, err := dbConnect(cfg.GetString("mysql.read.user"), cfg.GetString("mysql.read.pass"), cfg.GetString("mysql.read.host"), cfg.GetString("mysql.read.db"))
	if err != nil {
		return nil, err
	}

	dbw, err := dbConnect(cfg.GetString("mysql.write.user"), cfg.GetString("mysql.write.pass"), cfg.GetString("mysql.write.host"), cfg.GetString("mysql.write.db"))
	if err != nil {
		return nil, err
	}

	return &dbRepo{
		DbR: dbr,
		DbW: dbw,
	}, nil
}

func (d *dbRepo) i() {}

func (d *dbRepo) GetDbR() *gorm.DB {
	return d.DbR
}

func (d *dbRepo) GetDbW() *gorm.DB {
	return d.DbW
}

func (d *dbRepo) DbRClose() error {
	sqlDB, err := d.DbR.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *dbRepo) DbWClose() error {
	sqlDB, err := d.DbW.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func dbConnect(user, pass, addr, dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		user,
		pass,
		addr,
		dbName,
		true,
		"Local")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[db connection failed] Database name: %s", dbName))
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	cfg := config.GetConfig()

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(cfg.GetInt("mysql.base.maxOpenConn"))

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(cfg.GetInt("mysql.base.maxIdleConn"))

	// 设置最大连接超时
	sqlDB.SetConnMaxLifetime(time.Minute * time.Duration(cfg.GetInt("mysql.base.connMaxLifeTime")))

	// 使用插件
	db.Use(&TracePlugin{})

	return db, nil
}
