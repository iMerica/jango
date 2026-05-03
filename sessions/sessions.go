package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const (
	defaultSessionCookieName    = "sessionid"
	defaultSessionCookieAge      = 1209600
	defaultSessionCookieHTTPOnly = true
	defaultSessionCookieSameSite = "Lax"
)

func generateSessionKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type Session interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Has(key string) bool
	Keys() []string
	Save() error
	IsModified() bool
	Accessed() bool
	SessionKey() string
	ExpireDate() time.Time
	SetExpiry(duration time.Duration)
}

type SessionEngine interface {
	Create(sessionKey string) (Session, error)
	Load(sessionKey string) (Session, error)
	Save(session Session) error
	Delete(sessionKey string) error
	Exists(sessionKey string) bool
}

type Config struct {
	Engine        string
	CookieName    string
	CookieAge     int
	CookieSecure  bool
	CookieHTTPOnly bool
	CookieSameSite string
	CookieDomain  string
	CookiePath    string
}

func DefaultConfig() *Config {
	return &Config{
		Engine:        "db",
		CookieName:    defaultSessionCookieName,
		CookieAge:     defaultSessionCookieAge,
		CookieSecure:  false,
		CookieHTTPOnly: defaultSessionCookieHTTPOnly,
		CookieSameSite: defaultSessionCookieSameSite,
		CookiePath:    "/",
	}
}

type baseSession struct {
	key       string
	data      map[string]interface{}
	modified  bool
	accessed  bool
	expiry    time.Time
	engine    SessionEngine
}

func newBaseSession(key string, engine SessionEngine) *baseSession {
	return &baseSession{
		key:   key,
		data:  make(map[string]interface{}),
		engine: engine,
	}
}

func (s *baseSession) Get(key string) (interface{}, bool) {
	s.accessed = true
	val, ok := s.data[key]
	return val, ok
}

func (s *baseSession) Set(key string, value interface{}) {
	s.modified = true
	s.accessed = true
	s.data[key] = value
}

func (s *baseSession) Delete(key string) {
	s.modified = true
	s.accessed = true
	delete(s.data, key)
}

func (s *baseSession) Has(key string) bool {
	s.accessed = true
	_, ok := s.data[key]
	return ok
}

func (s *baseSession) Keys() []string {
	s.accessed = true
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s *baseSession) Save() error {
	s.modified = true
	return s.engine.Save(s)
}

func (s *baseSession) IsModified() bool { return s.modified }
func (s *baseSession) Accessed() bool    { return s.accessed }
func (s *baseSession) SessionKey() string { return s.key }

func (s *baseSession) ExpireDate() time.Time {
	if s.expiry.IsZero() {
		return time.Now().Add(time.Duration(defaultSessionCookieAge) * time.Second)
	}
	return s.expiry
}

func (s *baseSession) SetExpiry(duration time.Duration) {
	s.expiry = time.Now().Add(duration)
	s.modified = true
}

type DBSessionEngine struct {
	driver DBSessionDriver
	mu     sync.Mutex
}

type DBSessionDriver interface {
	Load(sessionKey string) (map[string]interface{}, error)
	Save(sessionKey string, data map[string]interface{}, expiry time.Time) error
	Delete(sessionKey string) error
	Exists(sessionKey string) bool
}

func NewDBSessionEngine(driver DBSessionDriver) *DBSessionEngine {
	return &DBSessionEngine{driver: driver}
}

func (e *DBSessionEngine) Create(sessionKey string) (Session, error) {
	if sessionKey == "" {
		key, err := generateSessionKey()
		if err != nil {
			return nil, err
		}
		sessionKey = key
	}
	s := newBaseSession(sessionKey, e)
	return s, nil
}

func (e *DBSessionEngine) Load(sessionKey string) (Session, error) {
	if sessionKey == "" {
		return e.Create("")
	}
	data, err := e.driver.Load(sessionKey)
	if err != nil {
		return e.Create("")
	}
	s := newBaseSession(sessionKey, e)
	s.data = data
	s.accessed = true
	return s, nil
}

func (e *DBSessionEngine) Save(session Session) error {
	bs := session.(*baseSession)
	return e.driver.Save(bs.key, bs.data, bs.ExpireDate())
}

func (e *DBSessionEngine) Delete(sessionKey string) error {
	return e.driver.Delete(sessionKey)
}

func (e *DBSessionEngine) Exists(sessionKey string) bool {
	return e.driver.Exists(sessionKey)
}

type CacheSessionEngine struct {
	driver CacheSessionDriver
}

type CacheSessionDriver interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
}

func NewCacheSessionEngine(driver CacheSessionDriver) *CacheSessionEngine {
	return &CacheSessionEngine{driver: driver}
}

func (e *CacheSessionEngine) Create(sessionKey string) (Session, error) {
	if sessionKey == "" {
		key, err := generateSessionKey()
		if err != nil {
			return nil, err
		}
		sessionKey = key
	}
	s := newBaseSession(sessionKey, e)
	return s, nil
}

func (e *CacheSessionEngine) Load(sessionKey string) (Session, error) {
	if sessionKey == "" {
		return e.Create("")
	}
	val, ok := e.driver.Get("session:" + sessionKey)
	if !ok {
		return e.Create("")
	}
	data, ok := val.(map[string]interface{})
	if !ok {
		return e.Create("")
	}
	s := newBaseSession(sessionKey, e)
	s.data = data
	s.accessed = true
	return s, nil
}

func (e *CacheSessionEngine) Save(session Session) error {
	bs := session.(*baseSession)
	ttl := time.Until(bs.ExpireDate())
	if ttl <= 0 {
		ttl = time.Duration(defaultSessionCookieAge) * time.Second
	}
	e.driver.Set("session:"+bs.key, bs.data, ttl)
	return nil
}

func (e *CacheSessionEngine) Delete(sessionKey string) error {
	e.driver.Delete("session:" + sessionKey)
	return nil
}

func (e *CacheSessionEngine) Exists(sessionKey string) bool {
	_, ok := e.driver.Get("session:" + sessionKey)
	return ok
}

type SignedCookieSessionEngine struct {
	secretKey        string
	salt             string
	maxAge           int
	fallbackKeys     []string
}

func NewSignedCookieSessionEngine(secretKey string, opts ...SignedCookieOption) *SignedCookieSessionEngine {
	e := &SignedCookieSessionEngine{
		secretKey:   secretKey,
		salt:        "jango.sessions",
		maxAge:       defaultSessionCookieAge,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

type SignedCookieOption func(*SignedCookieSessionEngine)

func WithSignedCookieSalt(salt string) SignedCookieOption {
	return func(e *SignedCookieSessionEngine) { e.salt = salt }
}

func WithSignedCookieMaxAge(maxAge int) SignedCookieOption {
	return func(e *SignedCookieSessionEngine) { e.maxAge = maxAge }
}

func WithSignedCookieFallbackKeys(keys []string) SignedCookieOption {
	return func(e *SignedCookieSessionEngine) { e.fallbackKeys = keys }
}

func (e *SignedCookieSessionEngine) Create(sessionKey string) (Session, error) {
	if sessionKey == "" {
		key, err := generateSessionKey()
		if err != nil {
			return nil, err
		}
		sessionKey = key
	}
	s := newBaseSession(sessionKey, e)
	return s, nil
}

func (e *SignedCookieSessionEngine) Load(sessionKey string) (Session, error) {
	if sessionKey == "" {
		return e.Create("")
	}
	s := newBaseSession(sessionKey, e)
	s.accessed = true
	return s, nil
}

func (e *SignedCookieSessionEngine) Save(session Session) error {
	return nil
}

func (e *SignedCookieSessionEngine) Delete(sessionKey string) error {
	return nil
}

func (e *SignedCookieSessionEngine) Exists(sessionKey string) bool {
	return sessionKey != ""
}