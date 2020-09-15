package orm

import (
	"github.com/go-sql-driver/mysql"
	"time"
	"xorm.io/xorm"
)

const (
	MySQL_Driver = "mysql"
)

func amendMySQLDSN(dsn string, extra ...map[string]string) (newDSN string, err error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return
	}
	cfg.ParseTime = true
	cfg.Loc = time.Local
	for _, vs := range extra {
		for k, v := range vs {
			cfg.Params[k] = v
		}
	}
	return cfg.FormatDSN(), nil
}

func NewMySQL(cfg *ORMOption) (*xorm.Engine, error) {
	dsn, err := amendMySQLDSN(cfg.DSN, cfg.Extra)
	if err != nil {
		return nil, err
	}
	dbe, err := xorm.NewEngine(MySQL_Driver, dsn)
	if err != nil {
		return nil, err
	}
	dbe.ShowSQL(cfg.ShowSQL)
	return dbe, nil
}
