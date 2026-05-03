package db

import (
	"context"
	"fmt"
	"time"

	"github.com/iMerica/jango/orm"
)

type Config struct {
	Engine            string
	Host              string
	Port              int
	Name              string
	User              string
	Password          string
	Options           map[string]string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type Connection interface {
	Close() error
	Ping() error
}

func DefaultConfig() *Config {
	return &Config{
		Engine:            "postgresql",
		Host:              "localhost",
		Port:              5432,
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}
}

func (c *Config) ToORMConfig() *orm.DBConfig {
	sslMode := "prefer"
	if c.Options != nil {
		if v, ok := c.Options["sslmode"]; ok {
			sslMode = v
		}
	}
	return &orm.DBConfig{
		Host:              c.Host,
		Port:              c.Port,
		Name:              c.Name,
		User:              c.User,
		Password:          c.Password,
		SSLMode:           sslMode,
		MaxConns:          c.MaxConns,
		MinConns:          c.MinConns,
		MaxConnLifetime:   c.MaxConnLifetime,
		MaxConnIdleTime:   c.MaxConnIdleTime,
		HealthCheckPeriod: c.HealthCheckPeriod,
	}
}

func (c *Config) DSN() string {
	sslMode := "prefer"
	if c.Options != nil {
		if v, ok := c.Options["sslmode"]; ok {
			sslMode = v
		}
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Name, c.User, c.Password, sslMode)
}

func Connect(ctx context.Context, config *Config) (*orm.DB, error) {
	ormConfig := config.ToORMConfig()
	return orm.OpenDB(ctx, ormConfig)
}
