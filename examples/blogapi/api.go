package blogapi

import (
	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/rest"
	"github.com/iMerica/jango/urls"
)

func PostSerializer() rest.Serializer[Post] {
	return rest.NewModelSerializer[Post](
		PostMeta,
		rest.Fields("id", "title", "slug", "body", "author_id", "category_id", "published_at", "created_at", "updated_at", "is_published"),
		rest.Field("id", rest.ReadOnly()),
		rest.Field("created_at", rest.ReadOnly()),
		rest.Field("updated_at", rest.ReadOnly()),
	)
}

func PostViewSet() rest.ModelViewSet[Post] {
	return rest.ModelViewSet[Post]{
		GenericAPIView: rest.GenericAPIView[Post]{
			QuerySet:        orm.Objects[Post]("blogapi", "Post"),
			Serializer:      PostSerializer(),
			FilterFields:    []string{"slug", "author_id", "category_id", "is_published"},
			SearchFields:    []string{"title", "body"},
			OrderingFields:  []string{"published_at", "created_at", "title"},
			DefaultPageSize: 20,
			APIView: rest.APIView{
				Permissions: []rest.Permission{rest.AllowAny{}},
			},
		},
	}
}

func URLPatterns() []urls.Pattern {
	router := rest.NewDefaultRouter()
	router.Register("posts", "post", PostViewSet())
	return router.URLPatterns()
}
