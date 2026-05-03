package postgres

import "github.com/iMerica/jango/orm"

func StringAgg(field string, delimiter string) orm.AggregateExpr {
	return orm.AggregateExpr{
		Name:        "string_agg",
		Function:    "STRING_AGG",
		Inner:       orm.F(field),
		OutputField: "string",
	}
}

func StringAggExpr(inner orm.Expr, delimiter string) orm.AggregateExpr {
	return orm.AggregateExpr{
		Name:        "string_agg",
		Function:    "STRING_AGG",
		Inner:       inner,
		OutputField: "string",
	}
}

func ArrayAgg(field string) orm.AggregateExpr {
	return orm.AggregateExpr{
		Name:        "array_agg",
		Function:    "ARRAY_AGG",
		Inner:       orm.F(field),
		OutputField: "array",
	}
}

func ArrayAggExpr(inner orm.Expr) orm.AggregateExpr {
	return orm.AggregateExpr{
		Name:        "array_agg",
		Function:    "ARRAY_AGG",
		Inner:       inner,
		OutputField: "array",
	}
}

func BitAnd(field string) orm.AggregateExpr {
	return orm.AggregateExpr{
		Name:     "bit_and",
		Function: "BIT_AND",
		Inner:    orm.F(field),
	}
}

func BitOr(field string) orm.AggregateExpr {
	return orm.AggregateExpr{
		Name:     "bit_or",
		Function: "BIT_OR",
		Inner:    orm.F(field),
	}
}

func BoolAnd(field string) orm.AggregateExpr {
	return orm.AggregateExpr{
		Name:     "bool_and",
		Function: "BOOL_AND",
		Inner:    orm.F(field),
	}
}

func BoolOr(field string) orm.AggregateExpr {
	return orm.AggregateExpr{
		Name:     "bool_or",
		Function: "BOOL_OR",
		Inner:    orm.F(field),
	}
}

func PercentileCont(percentile float64, field string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "PERCENTILE_CONT",
		Args: []orm.Expr{
			orm.ValueExpr{Value: percentile},
			orm.FuncExpr{Name: "WITHIN GROUP", Args: []orm.Expr{orm.FuncExpr{Name: "ORDER BY", Args: []orm.Expr{orm.F(field)}}}},
		},
	}
}

func PercentileDisc(percentile float64, field string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "PERCENTILE_DISC",
		Args: []orm.Expr{
			orm.ValueExpr{Value: percentile},
			orm.FuncExpr{Name: "WITHIN GROUP", Args: []orm.Expr{orm.FuncExpr{Name: "ORDER BY", Args: []orm.Expr{orm.F(field)}}}},
		},
	}
}

func Mode(field string) orm.FuncExpr {
	return orm.FuncExpr{
		Name: "MODE",
		Args: []orm.Expr{
			orm.FuncExpr{Name: "WITHIN GROUP", Args: []orm.Expr{orm.FuncExpr{Name: "ORDER BY", Args: []orm.Expr{orm.F(field)}}}},
		},
	}
}

func RowNumber() orm.FuncExpr {
	return orm.FuncExpr{
		Name: "ROW_NUMBER",
		Args: nil,
	}
}

func Rank() orm.FuncExpr {
	return orm.FuncExpr{
		Name: "RANK",
		Args: nil,
	}
}

func DenseRank() orm.FuncExpr {
	return orm.FuncExpr{
		Name: "DENSE_RANK",
		Args: nil,
	}
}

func Lag(field string, offset int, defaultValue interface{}) orm.FuncExpr {
	args := []orm.Expr{orm.F(field), orm.ValueExpr{Value: offset}}
	if defaultValue != nil {
		args = append(args, orm.ValueExpr{Value: defaultValue})
	}
	return orm.FuncExpr{Name: "LAG", Args: args}
}

func Lead(field string, offset int, defaultValue interface{}) orm.FuncExpr {
	args := []orm.Expr{orm.F(field), orm.ValueExpr{Value: offset}}
	if defaultValue != nil {
		args = append(args, orm.ValueExpr{Value: defaultValue})
	}
	return orm.FuncExpr{Name: "LEAD", Args: args}
}

func FirstValue(field string) orm.FuncExpr {
	return orm.FuncExpr{Name: "FIRST_VALUE", Args: []orm.Expr{orm.F(field)}}
}

func LastValue(field string) orm.FuncExpr {
	return orm.FuncExpr{Name: "LAST_VALUE", Args: []orm.Expr{orm.F(field)}}
}

func NthValue(field string, nth int) orm.FuncExpr {
	return orm.FuncExpr{Name: "NTH_VALUE", Args: []orm.Expr{orm.F(field), orm.ValueExpr{Value: nth}}}
}

func RandomExpr() orm.FuncExpr {
	return orm.FuncExpr{Name: "RANDOM"}
}

func MD5Expr(field string) orm.FuncExpr {
	return orm.FuncExpr{Name: "MD5", Args: []orm.Expr{orm.F(field)}}
}

func SHA256Expr(field string) orm.FuncExpr {
	return orm.FuncExpr{Name: "ENCODE", Args: []orm.Expr{orm.FuncExpr{Name: "DIGEST", Args: []orm.Expr{orm.F(field), orm.ValueExpr{Value: "sha256"}}}, orm.ValueExpr{Value: "hex"}}}
}

func NowExpr() orm.FuncExpr {
	return orm.FuncExpr{Name: "NOW"}
}

func CurrentDateExpr() orm.FuncExpr {
	return orm.FuncExpr{Name: "CURRENT_DATE"}
}

func CurrentTimeExpr() orm.FuncExpr {
	return orm.FuncExpr{Name: "CURRENT_TIME"}
}

func CurrentTimestampExpr() orm.FuncExpr {
	return orm.FuncExpr{Name: "CURRENT_TIMESTAMP"}
}

func DateTruncExpr(unit string, field string) orm.FuncExpr {
	return orm.FuncExpr{Name: "DATE_TRUNC", Args: []orm.Expr{orm.ValueExpr{Value: unit}, orm.F(field)}}
}

func DatePartExpr(part string, field string) orm.FuncExpr {
	return orm.FuncExpr{Name: "DATE_PART", Args: []orm.Expr{orm.ValueExpr{Value: part}, orm.F(field)}}
}

func AgeExpr(field string) orm.FuncExpr {
	return orm.FuncExpr{Name: "AGE", Args: []orm.Expr{orm.F(field)}}
}

func MakeDateExpr(year, month, day int) orm.FuncExpr {
	return orm.FuncExpr{Name: "MAKE_DATE", Args: []orm.Expr{orm.ValueExpr{Value: year}, orm.ValueExpr{Value: month}, orm.ValueExpr{Value: day}}}
}

func JSONBGetExpr(field string, path string) orm.FuncExpr {
	return orm.FuncExpr{Name: "", Args: []orm.Expr{orm.F(field), orm.ValueExpr{Value: path}}}
}

func JSONBContainsExpr(field string, value interface{}) orm.FuncExpr {
	return orm.FuncExpr{Name: "", Args: []orm.Expr{orm.F(field), orm.ValueExpr{Value: value}}}
}

func JSONBBrowserExpr(field string, path string) orm.FuncExpr {
	return orm.FuncExpr{Name: "JSONB_PATH_QUERY", Args: []orm.Expr{orm.F(field), orm.ValueExpr{Value: path}}}
}
