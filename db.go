package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbConfig struct {
	Vendor   string
	Host     string
	Port     uint16
	Username string
	Password string
	Name     string
	SSL      string
	Timeout  time.Duration
	MaxIdle  int
	MaxOpen  int
}

type SqlDatabase struct {
	*gorm.DB
}

func (c DbConfig) connectUrl() string {
	const template = "postgres://%s:%s@%s:%d/%s?sslmode=%s"
	return fmt.Sprintf(template, c.Username, c.Password, c.Host, c.Port, c.Name, c.SSL)
}

func NewDatabase(c DbConfig) (*SqlDatabase, error) {
	if c.Vendor != "postgres" {
		return nil, errors.New("unsupported-dialect-" + c.Vendor)
	}

	db, err := gorm.Open(postgres.Open(c.connectUrl()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(c.MaxIdle)
	sqlDB.SetMaxOpenConns(c.MaxOpen)
	sqlDB.SetConnMaxLifetime(c.Timeout)

	return &SqlDatabase{db}, nil
}

func GetDbConfig() DbConfig {
	return DbConfig{
		Vendor:   viper.GetString("db_vendor"),
		Host:     viper.GetString("db_host"),
		Port:     uint16(viper.GetInt("db_port")),
		Username: viper.GetString("db_user"),
		Password: viper.GetString("db_password"),
		Name:     viper.GetString("db_name"),
		SSL:      viper.GetString("db_ssl"),
		Timeout:  time.Duration(viper.GetInt("db_timeout")) * time.Minute,
		MaxIdle:  viper.GetInt("db_idle"),
		MaxOpen:  viper.GetInt("db_open"),
	}
}

func DBErrorNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
