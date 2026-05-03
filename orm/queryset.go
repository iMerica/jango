package orm

import (
	"context"
	"fmt"
	"strings"
)

type QuerySet struct {
	model   *ModelMeta
	db      *DB
	filters []QNode
	exclude []QNode
	orderBy []string
	limit   int
	offset  int
	distinct bool
	selectFields []string
	deferFields []string
	onlyFields   []string
	valuesFields []string
	valuesListFields []string
	flat      bool
	annotations []Annotation
	selectRelated []string
	prefetchRelated []PrefetchSpec
	compiler  *SQLCompiler
	_forUpdate bool
	_forUpdateNoWait bool
	_forUpdateSkipLocked bool
	noop      bool
}

type PrefetchSpec struct {
	Field     string
	QuerySet  *QuerySet
	ToAttr   string
}

func NewQuerySet(model *ModelMeta, db *DB) *QuerySet {
	return &QuerySet{
		model:    model,
		db:       db,
		filters:  nil,
		exclude:  nil,
		limit:    -1,
		offset:   0,
		distinct: false,
		compiler: NewSQLCompiler(),
	}
}

func (qs *QuerySet) clone() *QuerySet {
	newQS := &QuerySet{
		model:    qs.model,
		db:       qs.db,
		filters:  append([]QNode(nil), qs.filters...),
		exclude:  append([]QNode(nil), qs.exclude...),
		orderBy:  append([]string(nil), qs.orderBy...),
		limit:    qs.limit,
		offset:   qs.offset,
		distinct: qs.distinct,
		selectFields: append([]string(nil), qs.selectFields...),
		deferFields:  append([]string(nil), qs.deferFields...),
		onlyFields:   append([]string(nil), qs.onlyFields...),
		valuesFields: append([]string(nil), qs.valuesFields...),
		valuesListFields: append([]string(nil), qs.valuesListFields...),
		flat:        qs.flat,
		annotations: append([]Annotation(nil), qs.annotations...),
		selectRelated: append([]string(nil), qs.selectRelated...),
		prefetchRelated: append([]PrefetchSpec(nil), qs.prefetchRelated...),
		compiler:     qs.compiler,
		_forUpdate:   qs._forUpdate,
		_forUpdateNoWait: qs._forUpdateNoWait,
		_forUpdateSkipLocked: qs._forUpdateSkipLocked,
		noop:        qs.noop,
	}
	return newQS
}

func (qs *QuerySet) All() *QuerySet {
	return qs.clone()
}

func (qs *QuerySet) Filter(lookups ...Lookup) *QuerySet {
	newQS := qs.clone()
	q := Q(lookups...)
	newQS.filters = append(newQS.filters, q)
	return newQS
}

func (qs *QuerySet) FilterQ(q QNode) *QuerySet {
	newQS := qs.clone()
	newQS.filters = append(newQS.filters, q)
	return newQS
}

func (qs *QuerySet) Exclude(lookups ...Lookup) *QuerySet {
	newQS := qs.clone()
	q := Q(lookups...)
	newQS.exclude = append(newQS.exclude, q)
	return newQS
}

func (qs *QuerySet) ExcludeQ(q QNode) *QuerySet {
	newQS := qs.clone()
	newQS.exclude = append(newQS.exclude, q)
	return newQS
}

func (qs *QuerySet) OrderBy(fields ...string) *QuerySet {
	newQS := qs.clone()
	newQS.orderBy = append(newQS.orderBy, fields...)
	return newQS
}

func (qs *QuerySet) Reverse() *QuerySet {
	newQS := qs.clone()
	if len(newQS.orderBy) == 0 {
		if newQS.model != nil {
			newQS.orderBy = defaultReverseOrdering(newQS.model)
		}
	} else {
		for i := range newQS.orderBy {
			field := newQS.orderBy[i]
			if strings_HasPrefix(field, "-") {
				newQS.orderBy[i] = field[1:]
			} else {
				newQS.orderBy[i] = "-" + field
			}
		}
	}
	return newQS
}

func defaultReverseOrdering(model *ModelMeta) []string {
	ordering := model.DefaultOrdering
	if len(ordering) == 0 {
		ordering = model.Options.Ordering
	}
	if len(ordering) == 0 {
		return []string{}
	}
	result := make([]string, len(ordering))
	for i, f := range ordering {
		if strings_HasPrefix(f, "-") {
			result[i] = f[1:]
		} else {
			result[i] = "-" + f
		}
	}
	return result
}

func strings_HasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func (qs *QuerySet) Limit(n int) *QuerySet {
	newQS := qs.clone()
	newQS.limit = n
	return newQS
}

func (qs *QuerySet) Offset(n int) *QuerySet {
	newQS := qs.clone()
	newQS.offset = n
	return newQS
}

func (qs *QuerySet) Distinct() *QuerySet {
	newQS := qs.clone()
	newQS.distinct = true
	return newQS
}

func (qs *QuerySet) Only(fields ...string) *QuerySet {
	newQS := qs.clone()
	newQS.onlyFields = fields
	newQS.deferFields = nil
	return newQS
}

func (qs *QuerySet) Defer(fields ...string) *QuerySet {
	newQS := qs.clone()
	newQS.deferFields = append(newQS.deferFields, fields...)
	newQS.onlyFields = nil
	return newQS
}

func (qs *QuerySet) Select(fields ...string) *QuerySet {
	newQS := qs.clone()
	newQS.selectFields = fields
	return newQS
}

func (qs *QuerySet) Values(fields ...string) *ValuesQuerySet {
	newQS := qs.clone()
	vqs := &ValuesQuerySet{
		QuerySet: newQS,
	}
	vqs.valuesFields = fields
	return vqs
}

func (qs *QuerySet) ValuesList(fields ...string) *ValuesListQuerySet {
	newQS := qs.clone()
	vlqs := &ValuesListQuerySet{
		QuerySet: newQS,
	}
	vlqs.valuesListFields = fields
	return vlqs
}

func (qs *QuerySet) Annotate(annotations ...Annotation) *QuerySet {
	newQS := qs.clone()
	newQS.annotations = append(newQS.annotations, annotations...)
	return newQS
}

func (qs *QuerySet) Alias(name string, expr Expr) *QuerySet {
	newQS := qs.clone()
	newQS.annotations = append(newQS.annotations, Annotation{Name: name, Expr: expr})
	return newQS
}

func (qs *QuerySet) SelectRelated(fields ...string) *QuerySet {
	newQS := qs.clone()
	newQS.selectRelated = append(newQS.selectRelated, fields...)
	return newQS
}

func (qs *QuerySet) PrefetchRelated(field string, opts ...PrefetchOption) *QuerySet {
	newQS := qs.clone()
	spec := PrefetchSpec{Field: field}
	for _, opt := range opts {
		opt(&spec)
	}
	newQS.prefetchRelated = append(newQS.prefetchRelated, spec)
	return newQS
}

type PrefetchOption func(*PrefetchSpec)

func WithPrefetchQuerySet(qs *QuerySet) PrefetchOption {
	return func(s *PrefetchSpec) { s.QuerySet = qs }
}

func WithToAttr(attr string) PrefetchOption {
	return func(s *PrefetchSpec) { s.ToAttr = attr }
}

func (qs *QuerySet) ForUpdate(opts ...ForUpdateOption) *QuerySet {
	newQS := qs.clone()
	newQS._forUpdate = true
	for _, opt := range opts {
		opt(newQS)
	}
	return newQS
}

type ForUpdateOption func(*QuerySet)

func NoWait(qs *QuerySet)       { qs._forUpdateNoWait = true }
func SkipLocked(qs *QuerySet)   { qs._forUpdateSkipLocked = true }

func (qs *QuerySet) None() *QuerySet {
	newQS := qs.clone()
	newQS.noop = true
	return newQS
}

func (qs *QuerySet) Using(db *DB) *QuerySet {
	newQS := qs.clone()
	newQS.db = db
	return newQS
}

func (qs *QuerySet) First(ctx context.Context) (map[string]interface{}, error) {
	limited := qs.OrderBy(qs.model.PKField).Limit(1)
	results, err := limited.executeSelect(ctx)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (qs *QuerySet) Last(ctx context.Context) (map[string]interface{}, error) {
	limited := qs.Reverse().OrderBy(qs.model.PKField).Limit(1)
	results, err := limited.executeSelect(ctx)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (qs *QuerySet) Get(ctx context.Context, lookups ...Lookup) (map[string]interface{}, error) {
	filtered := qs.Filter(lookups...)
	results, err := filtered.executeSelect(ctx)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, &DoesNotExist{ModelName: qs.model.ModelName, Filter: formatLookups(lookups)}
	}
	if len(results) > 1 {
		return nil, &MultipleObjectsReturned{ModelName: qs.model.ModelName, Filter: formatLookups(lookups)}
	}
	return results[0], nil
}

func (qs *QuerySet) AllRecords(ctx context.Context) ([]map[string]interface{}, error) {
	return qs.executeSelect(ctx)
}

func (qs *QuerySet) Count(ctx context.Context) (int64, error) {
	sql, args := qs.compiler.CompileCount(qs.model, qs.filters, qs.exclude, qs.distinct)
	var count int64
	err := qs.db.QueryRow(ctx, sql, args...).Scan(&count)
	return count, err
}

func (qs *QuerySet) Exists(ctx context.Context) (bool, error) {
	count, err := qs.Limit(1).Count(ctx)
	return count > 0, err
}

func (qs *QuerySet) Aggregate(ctx context.Context, annotations ...Annotation) (map[string]interface{}, error) {
	sql, args := qs.compiler.CompileAggregate(qs.model, qs.filters, qs.exclude, annotations)
	row := qs.db.QueryRow(ctx, sql, args...)
	result := make(map[string]interface{})
	cols, err := row.Columns()
	if err != nil {
		return nil, err
	}
	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}
	if err := row.Scan(valPtrs...); err != nil {
		return nil, err
	}
	for i, col := range cols {
		result[col] = vals[i]
	}
	return result, nil
}

func (qs *QuerySet) Update(ctx context.Context, values map[string]interface{}) (int64, error) {
	sql, args := qs.compiler.CompileUpdate(qs.model, values, qs.filters, qs.exclude)
	tag, err := qs.db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (qs *QuerySet) UpdateExpr(ctx context.Context, updates map[string]Expr) (int64, error) {
	sql, args := qs.compiler.CompileUpdateExpr(qs.model, updates, qs.filters, qs.exclude)
	tag, err := qs.db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (qs *QuerySet) Delete(ctx context.Context) (int64, error) {
	sql, args := qs.compiler.CompileDelete(qs.model, qs.filters, qs.exclude)
	tag, err := qs.db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (qs *QuerySet) Create(ctx context.Context, values map[string]interface{}) (map[string]interface{}, error) {
	if qs.db == nil {
		return nil, fmt.Errorf("orm: no database connection")
	}
	sql, args := qs.compiler.CompileInsert(qs.model, values)
	tag, err := qs.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	if id, ok := tag.LastInsertId(); ok && id > 0 {
		return qs.Using(qs.db).Filter(L(qs.model.PKField, id)).First(ctx)
	}
	return values, nil
}

func (qs *QuerySet) GetOrCreate(ctx context.Context, defaults map[string]interface{}, lookups ...Lookup) (map[string]interface{}, bool, error) {
	filtered := qs.Filter(lookups...)
	results, err := filtered.executeSelect(ctx)
	if err != nil {
		return nil, false, err
	}
	if len(results) > 0 {
		return results[0], false, nil
	}
	values := make(map[string]interface{})
	for _, l := range lookups {
		values[l.Field] = l.Value
	}
	for k, v := range defaults {
		values[k] = v
	}
	created, err := qs.Create(ctx, values)
	if err != nil {
		return nil, false, err
	}
	return created, true, nil
}

func (qs *QuerySet) BulkCreate(ctx context.Context, records []map[string]interface{}) (int64, error) {
	if len(records) == 0 {
		return 0, nil
	}
	var totalAffected int64
	for _, record := range records {
		_, err := qs.Create(ctx, record)
		if err != nil {
			return totalAffected, err
		}
		totalAffected++
	}
	return totalAffected, nil
}

func (qs *QuerySet) executeSelect(ctx context.Context) ([]map[string]interface{}, error) {
	if qs.noop {
		return nil, nil
	}
	if qs.db == nil {
		return nil, fmt.Errorf("orm: no database connection")
	}
	sql, args := qs.compiler.CompileSelect(qs)
	rows, err := qs.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

type ValuesQuerySet struct {
	*QuerySet
}

func (vqs *ValuesQuerySet) AllRecords(ctx context.Context) ([]map[string]interface{}, error) {
	return vqs.executeSelect(ctx)
}

func (vqs *ValuesQuerySet) First(ctx context.Context) (map[string]interface{}, error) {
	limited := vqs.Limit(1)
	results, err := limited.executeSelect(ctx)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

type ValuesListQuerySet struct {
	*QuerySet
	flat bool
}

func (vlqs *ValuesListQuerySet) Flat() *ValuesListQuerySet {
	newQS := vlqs.clone()
	vlqsNew := &ValuesListQuerySet{
		QuerySet: newQS,
		flat:     true,
	}
	vlqsNew.valuesListFields = vlqs.valuesListFields
	return vlqsNew
}

func (vlqs *ValuesListQuerySet) AllRecords(ctx context.Context) ([][]interface{}, error) {
	results, err := vlqs.executeSelect(ctx)
	if err != nil {
		return nil, err
	}
	var list [][]interface{}
	for _, row := range results {
		var vals []interface{}
		for _, field := range vlqs.valuesListFields {
			colName := vlqs.model.DBColumnForField(field)
			if v, ok := row[colName]; ok {
				vals = append(vals, v)
			} else {
				vals = append(vals, nil)
			}
		}
		list = append(list, vals)
	}
	return list, nil
}

func formatLookups(lookups []Lookup) string {
	parts := make([]string, len(lookups))
	for i, l := range lookups {
		parts[i] = fmt.Sprintf("%s%s=%v", l.Field, formatOp(l.Op), l.Value)
	}
	return strings_Join(parts, ", ")
}

func formatOp(op string) string {
	if op == "" || op == "exact" {
		return ""
	}
	return "__" + op
}

var _ = strings_Join

func strings_Join(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}
	var b strings.Builder
	b.Grow(n)
	b.WriteString(elems[0])
	for i := 1; i < len(elems); i++ {
		b.WriteString(sep)
		b.WriteString(elems[i])
	}
	return b.String()
}

func scanRows(rows Rows) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	colTypes, err := rows.ColumnTypes()
	_ = colTypes
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}
	return results, rows.Err()
}