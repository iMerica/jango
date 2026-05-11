package rest

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iMerica/jango/auth"
	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/sessions"
)

type Authenticator interface {
	Authenticate(req *APIRequest) (interface{}, error)
	WWWAuthenticate() string
}

var ErrNoCredentials = errors.New("rest: no credentials")

type SessionAuthentication struct{}

func (a SessionAuthentication) Authenticate(req *APIRequest) (interface{}, error) {
	if req.User != nil {
		return req.User, nil
	}
	session, ok := req.Session.(sessions.Session)
	if !ok {
		return nil, ErrNoCredentials
	}
	rawID, ok := session.Get("_auth_user_id")
	if !ok {
		return nil, ErrNoCredentials
	}
	id, ok := numericID(rawID)
	if !ok {
		return nil, ErrNoCredentials
	}
	user, err := auth.NewDefaultBackendChain().GetUser(req.Context(), id)
	if err != nil {
		return nil, err
	}
	req.User = user
	return user, nil
}

func (a SessionAuthentication) WWWAuthenticate() string { return "" }

type BasicAuthentication struct{}

func (a BasicAuthentication) Authenticate(req *APIRequest) (interface{}, error) {
	header := req.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Basic ") {
		return nil, ErrNoCredentials
	}
	payload, err := base64.StdEncoding.DecodeString(strings.TrimSpace(strings.TrimPrefix(header, "Basic ")))
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid basic credentials")
	}
	user, err := auth.NewDefaultBackendChain().Authenticate(req.Context(), parts[0], parts[1])
	if err != nil {
		return nil, err
	}
	req.User = user
	return user, nil
}

func (a BasicAuthentication) WWWAuthenticate() string { return `Basic realm="api"` }

type Token struct {
	Key       string
	UserID    int64
	CreatedAt time.Time
}

var TokenMeta *orm.ModelMeta

func init() {
	TokenMeta = orm.GlobalRegistry().Register("rest", "Token", &orm.ModelMeta{
		AppLabel:  "rest",
		ModelName: "Token",
		TableName: "rest_token",
		PKField:   "Key",
		Fields: []orm.FieldDef{
			{Name: "Key", DBColumn: "key", FieldType: orm.CharFieldType, MaxLength: 40, PrimaryKey: true, Editable: true},
			orm.ForeignKey("User", "auth.User", orm.WithDBColumn("user_id")),
			orm.DateTimeField("CreatedAt", orm.WithAutoNowAdd, orm.WithDBColumn("created_at")),
		},
	})
}

type TokenAuthentication struct{}

func (a TokenAuthentication) Authenticate(req *APIRequest) (interface{}, error) {
	header := req.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Token ") && !strings.HasPrefix(header, "Bearer ") {
		return nil, ErrNoCredentials
	}
	key := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(header, "Token "), "Bearer "))
	if key == "" {
		return nil, ErrNoCredentials
	}
	db := orm.DefaultDB()
	if db == nil {
		return nil, fmt.Errorf("rest: database not available")
	}
	var userID int64
	if err := db.QueryRow(req.Context(), "SELECT user_id FROM rest_token WHERE key = $1", key).Scan(&userID); err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	user, err := auth.NewDefaultBackendChain().GetUser(req.Context(), userID)
	if err != nil {
		return nil, err
	}
	req.User = user
	return user, nil
}

func (a TokenAuthentication) WWWAuthenticate() string { return "Token" }

func authenticateRequest(req *APIRequest, authenticators []Authenticator) *APIResponse {
	if len(authenticators) == 0 {
		return nil
	}
	var challenge string
	for _, authenticator := range authenticators {
		user, err := authenticator.Authenticate(req)
		if err == nil {
			req.User = user
			return nil
		}
		if challenge == "" {
			challenge = authenticator.WWWAuthenticate()
		}
		if !errors.Is(err, ErrNoCredentials) {
			resp := ErrorResponse(err.Error(), http.StatusUnauthorized)
			if challenge != "" {
				resp.SetHeader("WWW-Authenticate", challenge)
			}
			return resp
		}
	}
	return nil
}

func numericID(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case string:
		var id int64
		if _, err := fmt.Sscanf(v, "%d", &id); err == nil {
			return id, true
		}
	}
	return 0, false
}

func CreateToken(ctx context.Context, key string, userID int64) (*Token, error) {
	token := &Token{Key: key, UserID: userID, CreatedAt: time.Now()}
	db := orm.DefaultDB()
	if db == nil {
		return nil, fmt.Errorf("rest: database not available")
	}
	_, err := db.Exec(ctx, "INSERT INTO rest_token (key, user_id, created_at) VALUES ($1, $2, $3)", token.Key, token.UserID, token.CreatedAt)
	return token, err
}
