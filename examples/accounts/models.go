package accounts

import (
	"time"

	"github.com/iMerica/jango/orm"
)

var UserMeta *orm.ModelMeta
var ProfileMeta *orm.ModelMeta
var GroupMeta *orm.ModelMeta
var PermissionMeta *orm.ModelMeta

func init() {
	UserMeta = orm.RegisterModel("accounts", "User", &orm.ModelMeta{
		AppLabel:  "accounts",
		ModelName: "User",
		TableName: "accounts_user",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Username", 150, orm.WithUnique),
			orm.CharField("Email", 254, orm.WithUnique),
			orm.CharField("Password", 128, orm.WithVerboseName("password hash")),
			orm.CharField("FirstName", 150, orm.WithNullable, orm.WithDBColumn("first_name")),
			orm.CharField("LastName", 150, orm.WithNullable, orm.WithDBColumn("last_name")),
			orm.BooleanField("IsActive", orm.WithDefault(true), orm.WithVerboseName("active")),
			orm.BooleanField("IsStaff", orm.WithDefault(false)),
			orm.BooleanField("IsAdmin", orm.WithDefault(false)),
			orm.DateTimeField("LastLogin", orm.WithNullable, orm.WithDBColumn("last_login")),
			orm.DateTimeField("DateJoined", orm.WithAutoNowAdd, orm.WithDBColumn("date_joined")),
		},
		DefaultOrdering: []string{"username"},
		Indexes: []orm.IndexDef{
			{Name: "accounts_user_username_idx", Fields: []string{"Username"}, Unique: true},
			{Name: "accounts_user_email_idx", Fields: []string{"Email"}, Unique: true},
		},
		Options: orm.ModelOptions{
			VerboseName:        "user",
			VerboseNamePlural:  "users",
			DefaultManagerName: "objects",
		},
	})

	ProfileMeta = orm.RegisterModel("accounts", "Profile", &orm.ModelMeta{
		AppLabel:  "accounts",
		ModelName: "Profile",
		TableName: "accounts_profile",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.OneToOneField("User", "accounts.User", orm.WithRelatedName("profile"), orm.WithDBColumn("user_id")),
			orm.TextField("Bio", orm.WithNullable),
			orm.CharField("Avatar", 255, orm.WithNullable),
			orm.DateTimeField("CreatedAt", orm.WithAutoNowAdd, orm.WithDBColumn("created_at")),
			orm.DateTimeField("UpdatedAt", orm.WithAutoNow, orm.WithDBColumn("updated_at")),
		},
		Options: orm.ModelOptions{
			VerboseName:       "profile",
			VerboseNamePlural: "profiles",
		},
	})

	GroupMeta = orm.RegisterModel("accounts", "Group", &orm.ModelMeta{
		AppLabel:  "accounts",
		ModelName: "Group",
		TableName: "accounts_group",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Name", 150, orm.WithUnique),
			orm.ManyToManyField("Users", "accounts.User", orm.WithRelatedName("groups"), orm.WithThrough("accounts.UserGroup")),
		},
		Options: orm.ModelOptions{
			VerboseName:       "group",
			VerboseNamePlural: "groups",
		},
	})

	PermissionMeta = orm.RegisterModel("accounts", "Permission", &orm.ModelMeta{
		AppLabel:  "accounts",
		ModelName: "Permission",
		TableName: "accounts_permission",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Name", 255),
			orm.CharField("Codename", 100),
			orm.CharField("AppLabel", 100, orm.WithDBColumn("app_label")),
			orm.ForeignKey("Group", "accounts.Group", orm.WithNullable, orm.WithRelatedName("permissions"), orm.WithDBColumn("group_id")),
		},
		Constraints: []orm.ConstraintDef{
			{Name: "unique_permission_codename", Unique: []string{"Codename", "AppLabel"}},
		},
		Options: orm.ModelOptions{
			VerboseName:       "permission",
			VerboseNamePlural: "permissions",
		},
	})
}

type User struct {
	ID         int64
	Username   string
	Email      string
	Password   string
	FirstName  string
	LastName   string
	IsActive   bool
	IsStaff    bool
	IsAdmin    bool
	LastLogin  time.Time
	DateJoined time.Time
}

func (u *User) TableName() string        { return "accounts_user" }
func (u *User) PKValue() interface{}     { return u.ID }
func (u *User) SetPKValue(v interface{}) { u.ID = v.(int64) }

type Profile struct {
	ID        int64
	UserID    int64
	Bio       string
	Avatar    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Profile) TableName() string        { return "accounts_profile" }
func (p *Profile) PKValue() interface{}     { return p.ID }
func (p *Profile) SetPKValue(v interface{}) { p.ID = v.(int64) }

type Group struct {
	ID   int64
	Name string
}

func (g *Group) TableName() string        { return "accounts_group" }
func (g *Group) PKValue() interface{}     { return g.ID }
func (g *Group) SetPKValue(v interface{}) { g.ID = v.(int64) }

type Permission struct {
	ID       int64
	Name     string
	Codename string
	AppLabel string
	GroupID  int64
}

func (p *Permission) TableName() string        { return "accounts_permission" }
func (p *Permission) PKValue() interface{}     { return p.ID }
func (p *Permission) SetPKValue(v interface{}) { p.ID = v.(int64) }
