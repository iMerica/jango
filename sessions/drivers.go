package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/iMerica/jango/conf"
	"github.com/iMerica/jango/orm"
)

type ormSessionDriver struct {
	settings *conf.Settings
}

func NewORMSessionDriver(settings *conf.Settings) *ormSessionDriver {
	return &ormSessionDriver{settings: settings}
}

func (d *ormSessionDriver) Load(sessionKey string) (map[string]interface{}, error) {
	db := orm.DefaultDB()
	if db == nil {
		return nil, fmt.Errorf("sessions: database not available")
	}

	ctx := context.Background()
	row := db.QueryRow(ctx,
		"SELECT session_data, expire_date FROM django_session WHERE session_key = $1",
		sessionKey)

	var data []byte
	var expireDate time.Time
	if err := row.Scan(&data, &expireDate); err != nil {
		return make(map[string]interface{}), nil
	}

	if time.Now().After(expireDate) {
		_, _ = db.Exec(ctx, "DELETE FROM django_session WHERE session_key = $1", sessionKey)
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return make(map[string]interface{}), nil
	}
	return result, nil
}

func (d *ormSessionDriver) Save(sessionKey string, data map[string]interface{}, expiry time.Time) error {
	db := orm.DefaultDB()
	if db == nil {
		return fmt.Errorf("sessions: database not available")
	}

	ctx := context.Background()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("sessions: failed to marshal session data: %w", err)
	}

	_, err = db.Exec(ctx,
		`INSERT INTO django_session (session_key, session_data, expire_date)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (session_key) DO UPDATE SET session_data = $2, expire_date = $3`,
		sessionKey, jsonData, expiry)

	return err
}

func (d *ormSessionDriver) Delete(sessionKey string) error {
	db := orm.DefaultDB()
	if db == nil {
		return fmt.Errorf("sessions: database not available")
	}

	ctx := context.Background()
	_, err := db.Exec(ctx, "DELETE FROM django_session WHERE session_key = $1", sessionKey)
	return err
}

func (d *ormSessionDriver) Exists(sessionKey string) bool {
	db := orm.DefaultDB()
	if db == nil {
		return false
	}

	ctx := context.Background()
	row := db.QueryRow(ctx,
		"SELECT 1 FROM django_session WHERE session_key = $1 AND expire_date > $2",
		sessionKey, time.Now())

	var exists int
	return row.Scan(&exists) == nil
}

type cacheDriverAdapter struct {
	cache interface {
		Get(key string) (interface{}, bool)
		Set(key string, value interface{}, ttl time.Duration)
		Delete(key string)
	}
}

func NewCacheDriverFromSettings(settings *conf.Settings) CacheSessionDriver {
	return &fallbackCacheDriver{}
}

type fallbackCacheDriver struct{}

func (d *fallbackCacheDriver) Get(key string) (interface{}, bool)                   { return nil, false }
func (d *fallbackCacheDriver) Set(key string, value interface{}, ttl time.Duration) {}
func (d *fallbackCacheDriver) Delete(key string)                                    {}
