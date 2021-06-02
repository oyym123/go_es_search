package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"push_service/src/core"
)

var db map[string]*gorm.DB

func init() {
	_db = make(map[string]*gorm.DB)
	dBConf := core.Config.Database
	for key, conf := range dBConf {
		db, err := gorm.Open(conf.DriverName, conf.DataSourceName)
		if err != nil {
			panic("连接数据库失败, error=" + err.Error())
		}

		//设置最大空闲连接数
		db.DB().SetMaxIdleConns(conf.MaxIdleNum)
		//设置最大连接数
		db.DB().SetMaxOpenConns(conf.MaxOpenNum)
		// 开启 Logger, 以展示详细的db日志
		//db.LogMode(core.Config.Logger.Debug)
		dbLog := core.Config.Logger.New("db_" + key)
		db.SetLogger(dbLog)
		_db[key] = db
	}

}

func getDB(DBName string) *gorm.DB {
	return _db[DBName]
}

func closeDB(DBName string) {
	_db[DBName].Close()
}
