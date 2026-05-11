package blogapi

import (
	"time"

	"github.com/iMerica/jango/contrib/postgres"
	"github.com/iMerica/jango/orm"
)

var PostMeta *orm.ModelMeta
var TagMeta *orm.ModelMeta
var CategoryMeta *orm.ModelMeta
var CommentMeta *orm.ModelMeta

func init() {
	PostMeta = orm.GlobalRegistry().Register("blogapi", "Post", &orm.ModelMeta{
		AppLabel:  "blogapi",
		ModelName: "Post",
		TableName: "blogapi_post",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Title", 200),
			orm.SlugField("Slug", 200, orm.WithUnique),
			orm.TextField("Body"),
			orm.JSONField("Metadata"),
			orm.ForeignKey("Author", "accounts.User", orm.WithRelatedName("posts"), orm.WithDBColumn("author_id")),
			orm.ForeignKey("Category", "blogapi.Category", orm.WithNullable, orm.WithOnDelete(orm.SetNull), orm.WithRelatedName("posts"), orm.WithDBColumn("category_id")),
			orm.ManyToManyField("Tags", "blogapi.Tag", orm.WithRelatedName("posts"), orm.WithThrough("blogapi.PostTag")),
			orm.DateTimeField("PublishedAt", orm.WithNullable, orm.WithDBColumn("published_at")),
			orm.DateTimeField("CreatedAt", orm.WithAutoNowAdd, orm.WithDBColumn("created_at")),
			orm.DateTimeField("UpdatedAt", orm.WithAutoNow, orm.WithDBColumn("updated_at")),
			orm.BooleanField("IsPublished", orm.WithDefault(false), orm.WithDBColumn("is_published")),
			postgres.SearchVectorFieldSimple("SearchVector"),
		},
		DefaultOrdering: []string{"-published_at"},
		Indexes: []orm.IndexDef{
			{Name: "blogapi_post_slug_idx", Fields: []string{"Slug"}, Unique: true},
			{Name: "blogapi_post_search_idx", Fields: []string{"SearchVector"}, Opclasses: []string{"gin_trgm_ops"}},
			postgres.PartialIndex("blogapi_post_published_idx", []string{"is_published"}, "is_published = true"),
		},
		Options: orm.ModelOptions{
			VerboseName:        "post",
			VerboseNamePlural:  "posts",
			DefaultManagerName: "objects",
		},
	})

	TagMeta = orm.GlobalRegistry().Register("blogapi", "Tag", &orm.ModelMeta{
		AppLabel:  "blogapi",
		ModelName: "Tag",
		TableName: "blogapi_tag",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Name", 100, orm.WithUnique),
			orm.SlugField("Slug", 100, orm.WithUnique),
		},
		Options: orm.ModelOptions{
			VerboseName:       "tag",
			VerboseNamePlural: "tags",
		},
	})

	CategoryMeta = orm.GlobalRegistry().Register("blogapi", "Category", &orm.ModelMeta{
		AppLabel:  "blogapi",
		ModelName: "Category",
		TableName: "blogapi_category",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.CharField("Name", 100, orm.WithUnique),
			orm.SlugField("Slug", 100, orm.WithUnique),
			orm.CharField("Description", 500, orm.WithNullable),
		},
		Options: orm.ModelOptions{
			VerboseName:       "category",
			VerboseNamePlural: "categories",
		},
	})

	CommentMeta = orm.GlobalRegistry().Register("blogapi", "Comment", &orm.ModelMeta{
		AppLabel:  "blogapi",
		ModelName: "Comment",
		TableName: "blogapi_comment",
		PKField:   "ID",
		Fields: []orm.FieldDef{
			orm.BigAutoField("ID"),
			orm.TextField("Body"),
			orm.CharField("AuthorName", 150, orm.WithDBColumn("author_name")),
			orm.CharField("AuthorEmail", 254, orm.WithDBColumn("author_email")),
			orm.ForeignKey("Post", "blogapi.Post", orm.WithRelatedName("comments"), orm.WithOnDelete(orm.Cascade), orm.WithDBColumn("post_id")),
			orm.DateTimeField("CreatedAt", orm.WithAutoNowAdd, orm.WithDBColumn("created_at")),
			orm.BooleanField("IsApproved", orm.WithDefault(false), orm.WithDBColumn("is_approved")),
		},
		DefaultOrdering: []string{"-created_at"},
		Indexes: []orm.IndexDef{
			postgres.PartialIndex("blogapi_comment_approved_idx", []string{"is_approved"}, "is_approved = true"),
		},
		Options: orm.ModelOptions{
			VerboseName:       "comment",
			VerboseNamePlural: "comments",
		},
	})
}

type Post struct {
	ID          int64
	Title       string
	Slug        string
	Body        string
	Metadata    interface{}
	AuthorID    int64
	CategoryID  *int64
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsPublished bool
}

func (p *Post) TableName() string        { return "blogapi_post" }
func (p *Post) PKValue() interface{}     { return p.ID }
func (p *Post) SetPKValue(v interface{}) { p.ID = v.(int64) }

type Tag struct {
	ID   int64
	Name string
	Slug string
}

func (t *Tag) TableName() string        { return "blogapi_tag" }
func (t *Tag) PKValue() interface{}     { return t.ID }
func (t *Tag) SetPKValue(v interface{}) { t.ID = v.(int64) }

type Category struct {
	ID          int64
	Name        string
	Slug        string
	Description string
}

func (c *Category) TableName() string        { return "blogapi_category" }
func (c *Category) PKValue() interface{}     { return c.ID }
func (c *Category) SetPKValue(v interface{}) { c.ID = v.(int64) }

type Comment struct {
	ID          int64
	Body        string
	AuthorName  string
	AuthorEmail string
	PostID      int64
	CreatedAt   time.Time
	IsApproved  bool
}

func (c *Comment) TableName() string        { return "blogapi_comment" }
func (c *Comment) PKValue() interface{}     { return c.ID }
func (c *Comment) SetPKValue(v interface{}) { c.ID = v.(int64) }
