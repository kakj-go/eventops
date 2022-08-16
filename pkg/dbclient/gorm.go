package dbclient

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"tiggerops/conf"
)

func DBClient(username, password, address, port, dbname string) (db *gorm.DB, err error) {
	// https://github.com/go-sql-driver/mysql#dsn-data-source-name
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, address, port, dbname)
	client, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if conf.IsDebug() {
		client = client.Debug()
	}
	return client, nil
}
