package mysql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	mySqlDsn = "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local"
)

type MysqlConfig struct {
	Enable   bool   `tome:"enable" yaml:"enable"`
	Host     string `toml:"host" yaml:"host"`
	Port     int    `toml:"port" yaml:"port"`
	Username string `toml:"username" yaml:"username"`
	Password string `toml:"password" yaml:"password"`
	DB       string `toml:"db" yaml:"db"`
}

func NewMysqlConfig(my MysqlConfig) *MysqlConfig {
	return &MysqlConfig{
		Enable:   my.Enable,
		Host:     my.Host,
		Port:     my.Port,
		Username: my.Username,
		Password: my.Password,
		DB:       my.DB,
	}

}

func BuildDSN(host string, port int, username string, password string, db string) string {
	return fmt.Sprintf(mySqlDsn, username, password, host, port, db)
}

func GormClient(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err

	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 设置数据库连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	//sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}
