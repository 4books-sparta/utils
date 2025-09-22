package utils

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	encodedPassword := url.QueryEscape(c.Password)
	return fmt.Sprintf(template, c.Username, encodedPassword, c.Host, c.Port, c.Name, c.SSL)
}

func NewDatabase(c DbConfig) (*SqlDatabase, error) {
	if c.Vendor != "postgres" {
		return nil, errors.New("unsupported-dialect-" + c.Vendor)
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)
	config := gorm.Config{
		Logger: newLogger,
	}

	db, err := gorm.Open(postgres.Open(c.connectUrl()), &config)
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
