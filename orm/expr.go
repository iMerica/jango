package orm

type Lookups []Lookup

type Lookup struct {
	Field string
	Op    string
	Value interface{}
}

func L(field string, value interface{}) Lookup {
	op, cleanField := parseLookup(field)
	return Lookup{Field: cleanField, Op: op, Value: value}
}

func parseLookup(field string) (string, string) {
	lookups := []string{
		"exact", "iexact", "contains", "icontains", "startswith", "istartswith",
		"endswith", "iendswith", "regex", "iregex", "gt", "gte", "lt", "lte",
		"in", "isnull", "range", "year", "month", "day", "week_day", "hour",
		"minute", "second", "date", "search",
	}
	for _, l := range lookups {
		suffix := "__" + l
		if len(field) > len(suffix) && field[len(field)-len(suffix):] == suffix {
			return l, field[:len(field)-len(suffix)]
		}
	}
	return "exact", field
}

type QNode struct {
	Connector Connector
	Negated   bool
	Children  []QNode
	Lookups   Lookups
}

type Connector string

const (
	AND Connector = "AND"
	OR  Connector = "OR"
)

func Q(lookups ...Lookup) QNode {
	return QNode{
		Connector: AND,
		Lookups:   lookups,
	}
}

func QAnd(children ...QNode) QNode {
	return QNode{
		Connector: AND,
		Children:  children,
	}
}

func QOr(children ...QNode) QNode {
	return QNode{
		Connector: OR,
		Children:  children,
	}
}

func QNot(child QNode) QNode {
	return QNode{
		Connector: AND,
		Negated:   true,
		Children:  []QNode{child},
	}
}

func QWith(connector Connector, negated bool, lookups Lookups, children ...QNode) QNode {
	return QNode{
		Connector: connector,
		Negated:   negated,
		Lookups:   lookups,
		Children:  children,
	}
}

type FExpr struct {
	Field string
}

func F(field string) FExpr {
	return FExpr{Field: field}
}

type Expr interface {
	ExprType() string
}

func (f FExpr) ExprType() string { return "f" }

type ValueExpr struct {
	Value interface{}
}

func (v ValueExpr) ExprType() string { return "value" }

func Value(v interface{}) ValueExpr {
	return ValueExpr{Value: v}
}

type RawExpr struct {
	SQL  string
	Args []interface{}
}

func (r RawExpr) ExprType() string { return "raw" }

func Raw(sql string, args ...interface{}) RawExpr {
	return RawExpr{SQL: sql, Args: args}
}

type CaseExpr struct {
	Conditions []WhenClause
	ElseExpr   Expr
	OutputField string
}

func (c CaseExpr) ExprType() string { return "case" }

type WhenClause struct {
	Condition QNode
	Result    Expr
}

func Case(whens ...WhenClause) CaseExprBuilder {
	return CaseExprBuilder{whens: whens}
}

type CaseExprBuilder struct {
	whens    []WhenClause
	elseExpr Expr
}

func (b CaseExprBuilder) When(condition QNode, result Expr) CaseExprBuilder {
	b.whens = append(b.whens, WhenClause{Condition: condition, Result: result})
	return b
}

func (b CaseExprBuilder) Default(expr Expr) CaseExpr {
	return CaseExpr{Conditions: b.whens, ElseExpr: expr}
}

type CoalesceExpr struct {
	Exprs []Expr
}

func (c CoalesceExpr) ExprType() string { return "coalesce" }

func Coalesce(exprs ...Expr) CoalesceExpr {
	return CoalesceExpr{Exprs: exprs}
}

type AggregateExpr struct {
	Name       string
	Function   string
	Inner      Expr
	Filter     *QNode
	Distinct   bool
	OutputField string
}

func (a AggregateExpr) ExprType() string { return "aggregate" }

func Count(inner Expr, opts ...AggOption) AggregateExpr {
	a := AggregateExpr{Name: "count", Function: "COUNT", Inner: inner}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

func CountStar(opts ...AggOption) AggregateExpr {
	a := AggregateExpr{Name: "count", Function: "COUNT", Inner: ValueExpr{Value: "*"}}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

func Sum(inner Expr, opts ...AggOption) AggregateExpr {
	a := AggregateExpr{Name: "sum", Function: "SUM", Inner: inner}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

func Avg(inner Expr, opts ...AggOption) AggregateExpr {
	a := AggregateExpr{Name: "avg", Function: "AVG", Inner: inner}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

func Max(inner Expr, opts ...AggOption) AggregateExpr {
	a := AggregateExpr{Name: "max", Function: "MAX", Inner: inner}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

func Min(inner Expr, opts ...AggOption) AggregateExpr {
	a := AggregateExpr{Name: "min", Function: "MIN", Inner: inner}
	for _, opt := range opts {
		opt(&a)
	}
	return a
}

type AggOption func(*AggregateExpr)

func WithFilter(q QNode) AggOption {
	return func(a *AggregateExpr) { a.Filter = &q }
}

func WithDistinct() AggOption {
	return func(a *AggregateExpr) { a.Distinct = true }
}

func WithOutputField(field string) AggOption {
	return func(a *AggregateExpr) { a.OutputField = field }
}

type SubqueryExpr struct {
	BaseQuerySet *BaseQuerySet
	OutputField string
}

func (s SubqueryExpr) ExprType() string { return "subquery" }

func Subquery(qs *BaseQuerySet) SubqueryExpr {
	return SubqueryExpr{BaseQuerySet: qs}
}

type OuterRef struct {
	Field string
}

func (o OuterRef) ExprType() string { return "outerref" }

type ExistsExpr struct {
	Subquery SubqueryExpr
	Negated bool
}

func (e ExistsExpr) ExprType() string { return "exists" }

func Exists(qs *BaseQuerySet) ExistsExpr {
	return ExistsExpr{Subquery: Subquery(qs)}
}

type FuncExpr struct {
	Name    string
	Args    []Expr
	OutputField string
}

func (f FuncExpr) ExprType() string { return "func" }

func Func(name string, args ...Expr) FuncExpr {
	return FuncExpr{Name: name, Args: args}
}

func CoalesceFunc(args ...Expr) FuncExpr {
	return FuncExpr{Name: "COALESCE", Args: args}
}

func ConcatFunc(args ...Expr) FuncExpr {
	return FuncExpr{Name: "CONCAT", Args: args}
}

func LengthFunc(arg Expr) FuncExpr {
	return FuncExpr{Name: "LENGTH", Args: []Expr{arg}}
}

func LowerFunc(arg Expr) FuncExpr {
	return FuncExpr{Name: "LOWER", Args: []Expr{arg}}
}

func UpperFunc(arg Expr) FuncExpr {
	return FuncExpr{Name: "UPPER", Args: []Expr{arg}}
}

func NowFunc() FuncExpr {
	return FuncExpr{Name: "NOW", Args: nil}
}

func ExtractFunc(field string, lookup string) FuncExpr {
	return FuncExpr{Name: "EXTRACT", Args: []Expr{RawExpr{SQL: lookup + " FROM " + field}}}
}

func DateTruncFunc(field string, lookup string) FuncExpr {
	return FuncExpr{Name: "DATE_TRUNC", Args: []Expr{ValueExpr{Value: lookup}, FExpr{Field: field}}}
}

type Annotation struct {
	Name string
	Expr Expr
}

func Annotate(name string, expr Expr) Annotation {
	return Annotation{Name: name, Expr: expr}
}