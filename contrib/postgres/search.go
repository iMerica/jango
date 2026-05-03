package postgres

import "github.com/iMerica/jango/orm"

func SearchVectorField(name string, opts ...orm.FieldOption) orm.FieldDef {
	f := orm.FieldDef{
		Name:      name,
		FieldType: orm.FieldType("searchvector"),
		DBColumn:  name,
		DBComment: "TSVECTOR",
		Nullable:  true,
		Editable:  false,
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func SearchVectorFieldSimple(name string) orm.FieldDef {
	return SearchVectorField(name)
}

func SearchQueryField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.FieldType("searchquery"),
		DBColumn:  name,
		DBComment: "TSQUERY",
		Nullable:  true,
		Editable:  false,
	}
}

func SearchRankField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.FieldType("searchrank"),
		DBColumn:  name,
		DBComment: "FLOAT",
		Nullable:  true,
		Editable:  false,
	}
}

func SearchVector(fields ...string) orm.FuncExpr {
	args := make([]orm.Expr, len(fields))
	for i, f := range fields {
		args[i] = orm.FuncExpr{Name: "COALESCE", Args: []orm.Expr{orm.F(f), orm.ValueExpr{Value: ""}}}
	}
	return orm.FuncExpr{Name: "TO_TSVECTOR", Args: args}
}

func SearchQuery(config string, query string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "TO_TSQUERY",
		Args: []orm.Expr{
			orm.ValueExpr{Value: config},
			orm.ValueExpr{Value: query},
		},
	}
}

func SearchPlainQuery(config string, query string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "PLAINTO_TSQUERY",
		Args: []orm.Expr{
			orm.ValueExpr{Value: config},
			orm.ValueExpr{Value: query},
		},
	}
}

func SearchPhraseQuery(config string, phrase string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "PHRASETO_TSQUERY",
		Args: []orm.Expr{
			orm.ValueExpr{Value: config},
			orm.ValueExpr{Value: phrase},
		},
	}
}

func SearchWebQuery(config string, query string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "WEBSEARCH_TO_TSQUERY",
		Args: []orm.Expr{
			orm.ValueExpr{Value: config},
			orm.ValueExpr{Value: query},
		},
	}
}

func SearchRank(vector string, query string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "TS_RANK",
		Args: []orm.Expr{
			orm.F(vector),
			orm.F(query),
		},
	}
}

func SearchRankCD(vector string, query string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "TS_RANK_CD",
		Args: []orm.Expr{
			orm.F(vector),
			orm.F(query),
		},
	}
}

func SearchHeadline(document string, query string, opts ...string) orm.FuncExpr {
	args := []orm.Expr{orm.F(document), orm.F(query)}
	if len(opts) > 0 {
		args = append(args, orm.ValueExpr{Value: opts[0]})
	}
	return orm.FuncExpr{Name: "TS_HEADLINE", Args: args}
}
