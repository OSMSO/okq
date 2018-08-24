package common

import (
	"os"
	"log"
	"encoding/json"
)

type (
	configuration struct {
		Server      string
		MysqlDBHost string
		MysqlDBUser string
		MysqlDBPwd  string
		Database    string
		UseDB       bool
	}
)

// AppConfig holds the configuration values from config.json file
var AppConfig configuration

// init Config file
func InitConfig() {
	file, err := os.Open("common/config.json")
	defer file.Close()
	if err != nil {
		log.Fatalf("[loadConfig]: %s\n", err)
	}
	decoder := json.NewDecoder(file)
	AppConfig = configuration{}
	err = decoder.Decode(&AppConfig)
	if err != nil {
		log.Fatalf("[logAppConfig]: %s\n", err)
	}
}
