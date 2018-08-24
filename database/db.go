package database

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/osmso/clock/models"
	. "github.com/osmso/clock/common"
	"log"
	"fmt"
)

var db *gorm.DB
var err error

func Init() *([]models.ClockExt) {
	add := fmt.Sprintf(
		"%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		AppConfig.MysqlDBUser,
		AppConfig.MysqlDBPwd,
		AppConfig.MysqlDBHost,
		AppConfig.Database)

	db, err = gorm.Open("mysql", add)
	if err != nil {
		log.Fatalf("[OpenDB]: %s\n", err)
	}
	db.AutoMigrate(&models.ClockExt{})

	return GetDbClocks(&([]models.ClockExt{}))
}

func GetDb() *gorm.DB {
	return db
}

func CloseDb() {
	db.Close()
}
