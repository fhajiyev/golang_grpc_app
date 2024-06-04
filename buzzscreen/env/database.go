package env

import (
	"fmt"

	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	gormtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/jinzhu/gorm"
)

type dbHolder struct {
	db *gorm.DB
}

var dbh *dbHolder

// GetDatabase returns a gorm db instance.
func GetDatabase() (*gorm.DB, error) {
	if dbh != nil {
		return dbh.db, nil
	}
	sd, err := getDBInstance()
	if err != nil {
		return nil, err
	}
	dbh = sd
	return dbh.db, nil
}

func getDBInstance() (*dbHolder, error) {
	core.Logger.Infof("env.getDBInstance() - %s:%s@tcp(%s:%s)/%s?parseTime=True", Config.Database.User, Config.Database.Password, Config.Database.Host, Config.Database.Port, Config.Database.Name)
	sqltrace.Register("mysql", &mysql.MySQLDriver{}, sqltrace.WithServiceName("buzzscreen-api-mysql"))
	gDB, err := gormtrace.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True", Config.Database.User, Config.Database.Password, Config.Database.Host, Config.Database.Port, Config.Database.Name))
	if err != nil {
		return nil, err
	}
	if Config.Database.LogMode {
		gDB.LogMode(true)
	}
	gDB.DB().SetMaxIdleConns(0)

	holder := &dbHolder{
		db: gDB,
	}
	gorm.NowFunc = func() time.Time {
		return time.Now().UTC()
	}
	return holder, nil
}
