package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/iMerica/jango/orm"
)

var UserModelLabel = "auth.User"

func GetUserModel() *orm.ModelMeta {
	parts := strings.SplitN(UserModelLabel, ".", 2)
	if len(parts) != 2 {
		panic("auth: invalid AUTH_USER_MODEL label: " + UserModelLabel)
	}
	meta, ok := orm.GlobalRegistry().Get(parts[0], parts[1])
	if !ok {
		panic("auth: user model " + UserModelLabel + " not registered")
	}
	return meta
}

type User struct {
	ID          int64
	Username    string
	Email       string
	Password    string
	FirstName   string
	LastName    string
	IsActive    bool
	IsStaff     bool
	IsAdmin     bool
	IsSuperuser bool
	LastLogin   *time.Time
	DateJoined  time.Time
	Groups      []string
	Permissions []string
}

func (u *User) TableName() string        { return "auth_user" }
func (u *User) PKValue() interface{}     { return u.ID }
func (u *User) SetPKValue(v interface{}) { u.ID = v.(int64) }

func (u *User) IsAuthenticated() bool { return u.ID != 0 && u.IsActive }
func (u *User) IsAnonymous() bool     { return u.ID == 0 }

func (u *User) FullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	return u.Username
}

func (u *User) ShortName() string {
	if u.FirstName != "" {
		return u.FirstName
	}
	return u.Username
}

func (u *User) HasPerm(perm string) bool {
	if u.IsSuperuser || u.IsAdmin {
		return true
	}
	for _, p := range u.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

func (u *User) HasModulePerms(appLabel string) bool {
	if u.IsSuperuser || u.IsAdmin {
		return true
	}
	for _, p := range u.Permissions {
		if strings.HasPrefix(p, appLabel+".") {
			return true
		}
	}
	return false
}

func (u *User) String() string { return u.Username }

type AnonymousUser struct{}

func (a *AnonymousUser) TableName() string                   { return "" }
func (a *AnonymousUser) PKValue() interface{}                { return int64(0) }
func (a *AnonymousUser) SetPKValue(v interface{})            {}
func (a *AnonymousUser) IsAuthenticated() bool               { return false }
func (a *AnonymousUser) IsAnonymous() bool                   { return true }
func (a *AnonymousUser) HasPerm(perm string) bool            { return false }
func (a *AnonymousUser) HasModulePerms(appLabel string) bool { return false }
func (a *AnonymousUser) String() string                      { return "AnonymousUser" }

var Anonymous = &AnonymousUser{}

var (
	UserMeta       *orm.ModelMeta
	GroupMeta      *orm.ModelMeta
	PermissionMeta *orm.ModelMeta
)

func init() {
	UserMeta = orm.GlobalRegistry().Register("auth", "User", &orm.ModelMeta{
		AppLabel:  "auth",
		ModelName: "User",
		TableName: "auth_user",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Username", 150, orm.WithUnique, orm.WithVerboseName("username")),
			orm.CharField("Email", 254, orm.WithVerboseName("email address")),
			orm.CharField("Password", 128, orm.WithVerboseName("password")),
			orm.CharField("FirstName", 150, orm.WithNullable, orm.WithDBColumn("first_name"), orm.WithVerboseName("first name")),
			orm.CharField("LastName", 150, orm.WithNullable, orm.WithDBColumn("last_name"), orm.WithVerboseName("last name")),
			orm.BooleanField("IsActive", orm.WithDefault(true), orm.WithVerboseName("active")),
			orm.BooleanField("IsStaff", orm.WithDefault(false), orm.WithVerboseName("staff status")),
			orm.BooleanField("IsAdmin", orm.WithDefault(false), orm.WithVerboseName("superuser status")),
			orm.DateTimeField("LastLogin", orm.WithNullable, orm.WithDBColumn("last_login")),
			orm.DateTimeField("DateJoined", orm.WithAutoNowAdd, orm.WithDBColumn("date_joined")),
		},
		DefaultOrdering: []string{"username"},
		Indexes: []orm.IndexDef{
			{Name: "auth_user_username_idx", Fields: []string{"Username"}, Unique: true},
		},
		Options: orm.ModelOptions{
			VerboseName:        "user",
			VerboseNamePlural:  "users",
			DefaultManagerName: "objects",
			DefaultPermissions: []string{"add", "change", "delete", "view"},
		},
	})

	GroupMeta = orm.GlobalRegistry().Register("auth", "Group", &orm.ModelMeta{
		AppLabel:  "auth",
		ModelName: "Group",
		TableName: "auth_group",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Name", 150, orm.WithUnique, orm.WithVerboseName("name")),
		},
		DefaultOrdering: []string{"name"},
		Options: orm.ModelOptions{
			VerboseName:        "group",
			VerboseNamePlural:  "groups",
			DefaultManagerName: "objects",
			DefaultPermissions: []string{"add", "change", "delete", "view"},
		},
	})

	PermissionMeta = orm.GlobalRegistry().Register("auth", "Permission", &orm.ModelMeta{
		AppLabel:  "auth",
		ModelName: "Permission",
		TableName: "auth_permission",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Name", 255, orm.WithVerboseName("name")),
			orm.CharField("Codename", 100, orm.WithVerboseName("codename")),
			orm.CharField("AppLabel", 100, orm.WithDBColumn("app_label"), orm.WithVerboseName("app label")),
			orm.ForeignKey("Group", "auth.Group", orm.WithNullable, orm.WithRelatedName("permissions"), orm.WithDBColumn("group_id")),
		},
		Constraints: []orm.ConstraintDef{
			{Name: "auth_permission_codename_app_label_idx", Unique: []string{"Codename", "AppLabel"}},
		},
		Options: orm.ModelOptions{
			VerboseName:        "permission",
			VerboseNamePlural:  "permissions",
			DefaultManagerName: "objects",
			DefaultPermissions: []string{"add", "change", "delete", "view"},
		},
	})
}

type Backend interface {
	Authenticate(ctx context.Context, username, password string) (*User, error)
	GetUser(ctx context.Context, userID int64) (*User, error)
}

type BackendChain struct {
	Backends []Backend
}

func NewBackendChain(backends ...Backend) *BackendChain {
	return &BackendChain{Backends: backends}
}

func (bc *BackendChain) Authenticate(ctx context.Context, username, password string) (*User, error) {
	for _, b := range bc.Backends {
		user, err := b.Authenticate(ctx, username, password)
		if err != nil {
			continue
		}
		if user != nil {
			return user, nil
		}
	}
	return nil, ErrAuthenticationFailed
}

func (bc *BackendChain) GetUser(ctx context.Context, userID int64) (*User, error) {
	for _, b := range bc.Backends {
		user, err := b.GetUser(ctx, userID)
		if err != nil {
			continue
		}
		if user != nil {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

var (
	ErrAuthenticationFailed = fmt.Errorf("auth: authentication failed")
	ErrUserNotFound         = fmt.Errorf("auth: user not found")
)

var backends []Backend

func RegisterBackend(b Backend) {
	backends = append(backends, b)
}

func GetBackends() []Backend {
	if len(backends) == 0 {
		return []Backend{&ModelBackend{}}
	}
	return backends
}

func NewDefaultBackendChain() *BackendChain {
	return NewBackendChain(GetBackends()...)
}
