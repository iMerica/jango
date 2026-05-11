package orm

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type SQLCompiler struct{}

func NewSQLCompiler() *SQLCompiler {
	return &SQLCompiler{}
}

func replacePlaceholders(query string, startIdx int) string {
	for strings.Contains(query, "$NEXT") {
		query = strings.Replace(query, "$NEXT", fmt.Sprintf("$%d", startIdx), 1)
		startIdx++
	}
	return query
}

func (c *SQLCompiler) CompileSelect(qs *BaseQuerySet) (string, []interface{}) {
	var b sqlBuilder
	var args []interface{}

	b.WriteString("SELECT ")

	if qs.distinct {
		b.WriteString("DISTINCT ")
	}

	if len(qs.annotations) > 0 {
		fields := c.selectFields(qs)
		b.WriteString(fields)
	} else if len(qs.selectFields) > 0 {
		b.WriteString(c.quoteColumns(qs.model, qs.selectFields))
	} else if len(qs.onlyFields) > 0 {
		b.WriteString(c.quoteColumns(qs.model, qs.onlyFields))
	} else if len(qs.deferFields) > 0 {
		allFields := c.allColumnsExcept(qs.model, qs.deferFields)
		b.WriteString(allFields)
	} else {
		b.WriteString(quote(qs.model.TableName) + ".*")
	}

	if len(qs.annotations) > 0 {
		for _, ann := range qs.annotations {
			exprSQL, exprArgs := c.compileExpr(ann.Expr, qs.model)
			b.WriteString(", ")
			b.WriteString(exprSQL)
			b.WriteString(" AS ")
			b.WriteString(quote(ann.Name))
			args = append(args, exprArgs...)
		}
	}

	b.WriteString(" FROM ")
	b.WriteString(quote(qs.model.TableName))

	if len(qs.selectRelated) > 0 {
		joinSQL, joinArgs := c.compileSelectRelated(qs.model, qs.selectRelated)
		b.WriteString(" ")
		b.WriteString(joinSQL)
		args = append(args, joinArgs...)
	}

	where, whereArgs := c.compileWhere(qs.model, qs.filters, qs.exclude)
	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
		args = append(args, whereArgs...)
	}

	if len(qs.orderBy) > 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(c.compileOrderBy(qs.model, qs.orderBy))
	} else if len(qs.model.DefaultOrdering) > 0 {
		b.WriteString(" ORDER BY ")
		b.WriteString(c.compileOrderBy(qs.model, qs.model.DefaultOrdering))
	}

	if qs.limit >= 0 {
		b.WriteString(fmt.Sprintf(" LIMIT %d", qs.limit))
	}
	if qs.offset > 0 {
		b.WriteString(fmt.Sprintf(" OFFSET %d", qs.offset))
	}

	if qs._forUpdate {
		b.WriteString(" FOR UPDATE")
		if qs._forUpdateNoWait {
			b.WriteString(" NOWAIT")
		} else if qs._forUpdateSkipLocked {
			b.WriteString(" SKIP LOCKED")
		}
	}
	return replacePlaceholders(b.String(), 1), args
}

func (c *SQLCompiler) CompileCount(model *ModelMeta, filters []QNode, exclude []QNode, distinct bool) (string, []interface{}) {
	var b sqlBuilder
	var args []interface{}

	b.WriteString("SELECT COUNT(")
	if distinct {
		b.WriteString("DISTINCT ")
	}
	b.WriteString(quote(model.TableName) + "." + quote(model.PKColumn()))
	b.WriteString(") FROM ")
	b.WriteString(quote(model.TableName))

	where, whereArgs := c.compileWhere(model, filters, exclude)
	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
		args = append(args, whereArgs...)
	}
	return replacePlaceholders(b.String(), 1), args
}

func (c *SQLCompiler) CompileAggregate(model *ModelMeta, filters []QNode, exclude []QNode, annotations []Annotation) (string, []interface{}) {
	var b sqlBuilder
	var args []interface{}

	b.WriteString("SELECT ")
	for i, ann := range annotations {
		if i > 0 {
			b.WriteString(", ")
		}
		exprSQL, exprArgs := c.compileExpr(ann.Expr, model)
		b.WriteString(exprSQL)
		b.WriteString(" AS ")
		b.WriteString(quote(ann.Name))
		args = append(args, exprArgs...)
	}

	b.WriteString(" FROM ")
	b.WriteString(quote(model.TableName))

	where, whereArgs := c.compileWhere(model, filters, exclude)
	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
		args = append(args, whereArgs...)
	}
	return replacePlaceholders(b.String(), 1), args
}

func (c *SQLCompiler) CompileInsert(model *ModelMeta, values map[string]interface{}) (string, []interface{}) {
	var b sqlBuilder

	keys := make([]string, 0, len(values))
	for k, v := range values {
		if isZeroAutoPKValue(model, k, v) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.WriteString("INSERT INTO ")
	b.WriteString(quote(model.TableName))

	if len(keys) == 0 {
		b.WriteString(" DEFAULT VALUES")
	} else {
		b.WriteString(" (")

		for i, k := range keys {
			colName := model.DBColumnForField(k)
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(quote(colName))
		}

		b.WriteString(") VALUES (")
	}

	args := make([]interface{}, 0, len(keys))
	if len(keys) > 0 {
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("$%d", i+1))
			args = append(args, values[k])
		}

		b.WriteString(")")
	}

	if _, hasAutoPK := model.AutoPKField(); hasAutoPK {
		b.WriteString(" RETURNING ")
		b.WriteString(quote(model.PKColumn()))
	}

	return b.String(), args
}

func isZeroAutoPKValue(model *ModelMeta, name string, value interface{}) bool {
	f, ok := model.FieldForNameOrColumn(name)
	if !ok || !f.PrimaryKey || !f.Auto {
		return false
	}
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return v.IsZero()
}

func (c *SQLCompiler) CompileUpdate(model *ModelMeta, values map[string]interface{}, filters []QNode, exclude []QNode) (string, []interface{}) {
	var b sqlBuilder
	var args []interface{}
	argIdx := 1

	b.WriteString("UPDATE ")
	b.WriteString(quote(model.TableName))
	b.WriteString(" SET ")

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		colName := model.DBColumnForField(k)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(quote(colName))
		b.WriteString(" = ")
		b.WriteString(fmt.Sprintf("$%d", argIdx))
		argIdx++
		args = append(args, values[k])
	}

	where, whereArgs := c.compileWhere(model, filters, exclude)
	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
		args = append(args, whereArgs...)
	}
	return replacePlaceholders(b.String(), argIdx), args
}

func (c *SQLCompiler) CompileUpdateExpr(model *ModelMeta, updates map[string]Expr, filters []QNode, exclude []QNode) (string, []interface{}) {
	var b sqlBuilder
	var args []interface{}
	argIdx := 1

	b.WriteString("UPDATE ")
	b.WriteString(quote(model.TableName))
	b.WriteString(" SET ")

	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		colName := model.DBColumnForField(k)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(quote(colName))
		b.WriteString(" = ")
		exprSQL, exprArgs := c.compileExpr(updates[k], model)
		b.WriteString(exprSQL)
		args = append(args, exprArgs...)
	}

	where, whereArgs := c.compileWhere(model, filters, exclude)
	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
		args = append(args, whereArgs...)
	}

	_ = argIdx
	return replacePlaceholders(b.String(), argIdx), args
}

func (c *SQLCompiler) CompileDelete(model *ModelMeta, filters []QNode, exclude []QNode) (string, []interface{}) {
	var b sqlBuilder
	var args []interface{}

	b.WriteString("DELETE FROM ")
	b.WriteString(quote(model.TableName))

	where, whereArgs := c.compileWhere(model, filters, exclude)
	if where != "" {
		b.WriteString(" WHERE ")
		b.WriteString(where)
		args = append(args, whereArgs...)
	}
	return replacePlaceholders(b.String(), 1), args
}

func (c *SQLCompiler) compileWhere(model *ModelMeta, filters []QNode, exclude []QNode) (string, []interface{}) {
	var parts []string
	var args []interface{}

	for _, q := range filters {
		sql, sqlArgs := c.compileQNode(model, q)
		parts = append(parts, sql)
		args = append(args, sqlArgs...)
	}

	for _, q := range exclude {
		sql, sqlArgs := c.compileQNode(model, q)
		parts = append(parts, "NOT ("+sql+")")
		args = append(args, sqlArgs...)
	}

	if len(parts) == 0 {
		return "", nil
	}
	if len(parts) == 1 {
		return parts[0], args
	}
	return "(" + strings.Join(parts, " AND ") + ")", args
}

func (c *SQLCompiler) compileQNode(model *ModelMeta, node QNode) (string, []interface{}) {
	var parts []string
	var args []interface{}

	for _, child := range node.Children {
		sql, sqlArgs := c.compileQNode(model, child)
		parts = append(parts, sql)
		args = append(args, sqlArgs...)
	}

	for _, lookup := range node.Lookups {
		sql, sqlArgs := c.compileLookup(model, lookup)
		parts = append(parts, sql)
		args = append(args, sqlArgs...)
	}

	if len(parts) == 0 {
		return "1=1", nil
	}

	var connector string
	if node.Connector == OR {
		connector = " OR "
	} else {
		connector = " AND "
	}

	result := strings.Join(parts, connector)
	if len(parts) > 1 {
		result = "(" + result + ")"
	}

	if node.Negated {
		result = "NOT " + result
	}

	return result, args
}

func (c *SQLCompiler) compileLookup(model *ModelMeta, lookup Lookup) (string, []interface{}) {
	colName := model.DBColumnForField(lookup.Field)
	qualifiedCol := quote(model.TableName) + "." + quote(colName)

	switch lookup.Op {
	case "exact":
		if lookup.Value == nil {
			return qualifiedCol + " IS NULL", nil
		}
		return qualifiedCol + " = $NEXT", []interface{}{lookup.Value}
	case "iexact":
		return "LOWER(" + qualifiedCol + ") = LOWER($NEXT)", []interface{}{lookup.Value}
	case "contains":
		return qualifiedCol + " LIKE $NEXT", []interface{}{"%" + fmt.Sprint(lookup.Value) + "%"}
	case "icontains":
		return "LOWER(" + qualifiedCol + ") LIKE LOWER($NEXT)", []interface{}{"%" + fmt.Sprint(lookup.Value) + "%"}
	case "startswith":
		return qualifiedCol + " LIKE $NEXT", []interface{}{fmt.Sprint(lookup.Value) + "%"}
	case "istartswith":
		return "LOWER(" + qualifiedCol + ") LIKE LOWER($NEXT)", []interface{}{fmt.Sprint(lookup.Value) + "%"}
	case "endswith":
		return qualifiedCol + " LIKE $NEXT", []interface{}{"%" + fmt.Sprint(lookup.Value)}
	case "iendswith":
		return "LOWER(" + qualifiedCol + ") LIKE LOWER($NEXT)", []interface{}{"%" + fmt.Sprint(lookup.Value)}
	case "regex":
		return qualifiedCol + " ~ $NEXT", []interface{}{lookup.Value}
	case "iregex":
		return qualifiedCol + " ~* $NEXT", []interface{}{lookup.Value}
	case "gt":
		return qualifiedCol + " > $NEXT", []interface{}{lookup.Value}
	case "gte":
		return qualifiedCol + " >= $NEXT", []interface{}{lookup.Value}
	case "lt":
		return qualifiedCol + " < $NEXT", []interface{}{lookup.Value}
	case "lte":
		return qualifiedCol + " <= $NEXT", []interface{}{lookup.Value}
	case "in":
		return c.compileIn(model, lookup)
	case "isnull":
		if lookup.Value.(bool) {
			return qualifiedCol + " IS NULL", nil
		}
		return qualifiedCol + " IS NOT NULL", nil
	case "range":
		switch v := lookup.Value.(type) {
		case [2]interface{}:
			return "(" + qualifiedCol + " BETWEEN $NEXT AND $NEXT)", []interface{}{v[0], v[1]}
		default:
			return qualifiedCol + " BETWEEN $NEXT AND $NEXT", []interface{}{lookup.Value}
		}
	case "search":
		return c.compileSearch(model, lookup)
	case "year", "month", "day", "hour", "minute", "second":
		return c.compileDateExtract(model, lookup, lookup.Op)
	case "date":
		return qualifiedCol + "::date = $NEXT::date", []interface{}{lookup.Value}
	default:
		return qualifiedCol + " = $NEXT", []interface{}{lookup.Value}
	}
}

func (c *SQLCompiler) compileIn(model *ModelMeta, lookup Lookup) (string, []interface{}) {
	colName := model.DBColumnForField(lookup.Field)
	qualifiedCol := quote(model.TableName) + "." + quote(colName)

	switch v := lookup.Value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return "1=0", nil
		}
		placeholders := make([]string, len(v))
		for i := range v {
			placeholders[i] = "$NEXT"
		}
		return qualifiedCol + " IN (" + strings.Join(placeholders, ", ") + ")", v
	case *BaseQuerySet:
		subSQL, subArgs := c.CompileSelect(v)
		return qualifiedCol + " IN (" + subSQL + ")", subArgs
	default:
		return qualifiedCol + " IN ($NEXT)", []interface{}{lookup.Value}
	}
}

func (c *SQLCompiler) compileSearch(model *ModelMeta, lookup Lookup) (string, []interface{}) {
	colName := model.DBColumnForField(lookup.Field)
	return quote(model.TableName) + "." + quote(colName) + " @@ plainto_tsquery($NEXT)",
		[]interface{}{lookup.Value}
}

func (c *SQLCompiler) compileDateExtract(model *ModelMeta, lookup Lookup, component string) (string, []interface{}) {
	colName := model.DBColumnForField(lookup.Field)
	qualifiedCol := quote(model.TableName) + "." + quote(colName)
	return "EXTRACT(" + strings.ToUpper(component) + " FROM " + qualifiedCol + ") = $NEXT",
		[]interface{}{lookup.Value}
}

func (c *SQLCompiler) compileExpr(expr Expr, model *ModelMeta) (string, []interface{}) {
	switch e := expr.(type) {
	case ValueExpr:
		return "$NEXT", []interface{}{e.Value}
	case FExpr:
		colName := model.DBColumnForField(e.Field)
		return quote(model.TableName) + "." + quote(colName), nil
	case RawExpr:
		return e.SQL, e.Args
	case FuncExpr:
		return c.compileFuncExpr(e, model)
	case AggregateExpr:
		return c.compileAggregateExpr(e, model)
	case CaseExpr:
		return c.compileCaseExpr(e, model)
	case CoalesceExpr:
		return c.compileCoalesceExpr(e, model)
	case SubqueryExpr:
		subSQL, subArgs := c.CompileSelect(e.BaseQuerySet)
		return "(" + subSQL + ")", subArgs
	case ExistsExpr:
		subSQL, subArgs := c.CompileSelect(e.Subquery.BaseQuerySet)
		if e.Negated {
			return "NOT EXISTS (" + subSQL + ")", subArgs
		}
		return "EXISTS (" + subSQL + ")", subArgs
	default:
		return "$NEXT", nil
	}
}

func (c *SQLCompiler) compileFuncExpr(f FuncExpr, model *ModelMeta) (string, []interface{}) {
	var args []interface{}
	argSQLs := make([]string, len(f.Args))
	for i, arg := range f.Args {
		sql, sqlArgs := c.compileExpr(arg, model)
		argSQLs[i] = sql
		args = append(args, sqlArgs...)
	}
	return f.Name + "(" + strings.Join(argSQLs, ", ") + ")", args
}

func (c *SQLCompiler) compileAggregateExpr(a AggregateExpr, model *ModelMeta) (string, []interface{}) {
	var args []interface{}

	if a.Inner != nil {
		innerSQL, innerArgs := c.compileExpr(a.Inner, model)
		args = append(args, innerArgs...)
		distinct := ""
		if a.Distinct {
			distinct = "DISTINCT "
		}
		if a.Filter != nil {
			filterSQL, filterArgs := c.compileQNode(model, *a.Filter)
			args = append(args, filterArgs...)
			return a.Function + "(" + distinct + innerSQL + ") FILTER (WHERE " + filterSQL + ")", args
		}
		return a.Function + "(" + distinct + innerSQL + ")", args
	}

	distinct := ""
	if a.Distinct {
		distinct = "DISTINCT "
	}
	return a.Function + "(" + distinct + "*)", args
}

func (c *SQLCompiler) compileCaseExpr(e CaseExpr, model *ModelMeta) (string, []interface{}) {
	var args []interface{}
	b := "CASE"
	for _, w := range e.Conditions {
		condSQL, condArgs := c.compileQNode(model, w.Condition)
		resultSQL, resultArgs := c.compileExpr(w.Result, model)
		b += " WHEN " + condSQL + " THEN " + resultSQL
		args = append(args, condArgs...)
		args = append(args, resultArgs...)
	}
	if e.ElseExpr != nil {
		elseSQL, elseArgs := c.compileExpr(e.ElseExpr, model)
		b += " ELSE " + elseSQL
		args = append(args, elseArgs...)
	}
	b += " END"
	return b, args
}

func (c *SQLCompiler) compileCoalesceExpr(e CoalesceExpr, model *ModelMeta) (string, []interface{}) {
	var args []interface{}
	parts := make([]string, len(e.Exprs))
	for i, expr := range e.Exprs {
		sql, sqlArgs := c.compileExpr(expr, model)
		parts[i] = sql
		args = append(args, sqlArgs...)
	}
	return "COALESCE(" + strings.Join(parts, ", ") + ")", args
}

func (c *SQLCompiler) compileSelectRelated(model *ModelMeta, fields []string) (string, []interface{}) {
	var parts []string
	var args []interface{}

	for _, field := range fields {
		fd, ok := model.FieldForNameOrColumn(field)
		if !ok || (fd.FieldType != ForeignKeyType && fd.FieldType != OneToOneType) {
			continue
		}
		relatedModel, ok := GlobalRegistry().Get(model.AppLabel, fd.RelatedModel)
		if !ok {
			relatedModel, ok = GlobalRegistry().Get("", fd.RelatedModel)
		}
		if !ok {
			continue
		}
		relatedPKColumn := relatedModel.PKColumn()
		parts = append(parts, "LEFT JOIN "+quote(relatedModel.TableName)+
			" ON "+quote(model.TableName)+"."+quote(model.DBColumnForField(fd.Name))+
			" = "+quote(relatedModel.TableName)+"."+quote(relatedPKColumn))
	}

	return strings.Join(parts, " "), args
}

func (c *SQLCompiler) compileOrderBy(model *ModelMeta, fields []string) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		desc := false
		fieldName := f
		if len(f) > 0 && f[0] == '-' {
			desc = true
			fieldName = f[1:]
		}
		colName := model.DBColumnForField(fieldName)
		parts[i] = quote(model.TableName) + "." + quote(colName)
		if desc {
			parts[i] += " DESC"
		}
	}
	return strings.Join(parts, ", ")
}

func (c *SQLCompiler) selectFields(qs *BaseQuerySet) string {
	var fields []string
	for _, f := range qs.model.Fields {
		if f.FieldType == ManyToManyType {
			continue
		}
		colName := qs.model.DBColumnForField(f.Name)
		fields = append(fields, quote(qs.model.TableName)+"."+quote(colName))
	}
	return strings.Join(fields, ", ")
}

func (c *SQLCompiler) quoteColumns(model *ModelMeta, fieldNames []string) string {
	parts := make([]string, len(fieldNames))
	for i, name := range fieldNames {
		colName := model.DBColumnForField(name)
		parts[i] = quote(model.TableName) + "." + quote(colName)
	}
	return strings.Join(parts, ", ")
}

func (c *SQLCompiler) allColumnsExcept(model *ModelMeta, except []string) string {
	exceptSet := make(map[string]bool, len(except))
	for _, f := range except {
		exceptSet[f] = true
	}
	var parts []string
	for _, f := range model.Fields {
		if exceptSet[f.Name] {
			continue
		}
		if f.FieldType == ManyToManyType {
			continue
		}
		colName := model.DBColumnForField(f.Name)
		parts = append(parts, quote(model.TableName)+"."+quote(colName))
	}
	return strings.Join(parts, ", ")
}

func quote(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

type sqlBuilder struct {
	strings.Builder
}

func (b *sqlBuilder) String() string {
	return b.Builder.String()
}
