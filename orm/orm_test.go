package orm

import (
	"strings"
	"testing"
	"time"
)

type User struct {
	ID uint `jango:"primary_key"`
}

type Widget struct {
	ID          uint      `jango:"primary_key"`
	Name        string    `jango:"type:char,max_length:100"`
	Description string    `jango:"type:text"`
	Count       int       `jango:"type:int"`
	Active      bool      `jango:"type:boolean,default:true"`
	CreatedAt   time.Time `jango:"auto_now_add,column:created_at"`
	Author      *User     `jango:"related_name:widgets,column:author_id"`
}

func (Widget) Meta() ModelOptions {
	return ModelOptions{
		VerboseName:       "widget",
		VerboseNamePlural: "widgets",
		TableName:         "test_widget",
		DefaultOrdering:   []string{"-created_at"},
	}
}

var testMeta *ModelMeta

func init() {
	RegisterModel("test", &User{})
	testMeta = RegisterModel("test", &Widget{})
}

func TestFieldConstructors(t *testing.T) {
	tests := []struct {
		name     string
		field    FieldDef
		ftype    FieldType
		nullable bool
		unique   bool
		maxLen   int
	}{
		{"auto", AutoField("id"), AutoFieldType, false, false, 0},
		{"bigauto", BigAutoField("id"), BigAutoFieldType, false, false, 0},
		{"char", CharField("name", 100), CharFieldType, false, false, 100},
		{"char_nullable", CharField("name", 100, WithNullable), CharFieldType, true, false, 100},
		{"char_unique", CharField("name", 100, WithUnique), CharFieldType, false, true, 100},
		{"text", TextField("body"), TextFieldType, false, false, 0},
		{"slug", SlugField("slug", 50), SlugFieldType, false, false, 50},
		{"email", EmailField("email", 0), EmailFieldType, false, false, 254},
		{"url", URLField("url", 0), URLFieldType, false, false, 200},
		{"uuid", UUIDField("id"), UUIDFieldType, false, false, 0},
		{"int", IntegerField("count"), IntFieldType, false, false, 0},
		{"bigint", BigIntegerField("count"), BigIntFieldType, false, false, 0},
		{"float", FloatField("score"), FloatFieldType, false, false, 0},
		{"decimal", DecimalField("price", 10, 2), DecimalFieldType, false, false, 10},
		{"bool", BooleanField("active"), BooleanFieldType, false, false, 0},
		{"nullbool", NullBooleanField("active"), NullBooleanFieldType, true, false, 0},
		{"date", DateField("date"), DateFieldType, false, false, 0},
		{"time", TimeField("time"), TimeFieldType, false, false, 0},
		{"datetime", DateTimeField("created"), DateTimeFieldType, false, false, 0},
		{"duration", DurationField("duration"), DurationFieldType, false, false, 0},
		{"json", JSONField("data"), JSONFieldType, true, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.FieldType != tt.ftype {
				t.Errorf("expected FieldType %s, got %s", tt.ftype, tt.field.FieldType)
			}
			if tt.field.Nullable != tt.nullable {
				t.Errorf("expected Nullable %v, got %v", tt.nullable, tt.field.Nullable)
			}
			if tt.field.Unique != tt.unique {
				t.Errorf("expected Unique %v, got %v", tt.unique, tt.field.Unique)
			}
			if tt.maxLen > 0 && tt.field.MaxLength != tt.maxLen {
				t.Errorf("expected MaxLength %d, got %d", tt.maxLen, tt.field.MaxLength)
			}
		})
	}
}

func TestFieldOptions(t *testing.T) {
	f := CharField("name", 100, WithNullable, WithUnique, WithDefault("hello"), WithDBColumn("full_name"), WithDBComment("user name"))
	if !f.Nullable {
		t.Error("expected Nullable to be true")
	}
	if !f.Unique {
		t.Error("expected Unique to be true")
	}
	if f.Default != "hello" {
		t.Errorf("expected Default 'hello', got %v", f.Default)
	}
	if f.DBColumn != "full_name" {
		t.Errorf("expected DBColumn 'full_name', got %s", f.DBColumn)
	}
	if f.DBComment != "user name" {
		t.Errorf("expected DBComment 'user name', got %s", f.DBComment)
	}
}

func TestForeignKey(t *testing.T) {
	f := ForeignKey("author", "auth.User", WithOnDelete(Cascade), WithRelatedName("articles"), WithDBColumn("author_id"))
	if f.FieldType != ForeignKeyType {
		t.Errorf("expected ForeignKeyType, got %s", f.FieldType)
	}
	if f.RelatedModel != "auth.User" {
		t.Errorf("expected RelatedModel 'auth.User', got %s", f.RelatedModel)
	}
	if f.OnDelete != Cascade {
		t.Errorf("expected OnDelete Cascade, got %s", f.OnDelete)
	}
	if f.RelatedName != "articles" {
		t.Errorf("expected RelatedName 'articles', got %s", f.RelatedName)
	}
	if f.DBColumn != "author_id" {
		t.Errorf("expected DBColumn 'author_id', got %s", f.DBColumn)
	}
}

func TestOneToOneField(t *testing.T) {
	f := OneToOneField("profile", "auth.User", WithRelatedName("profile"))
	if f.FieldType != OneToOneType {
		t.Errorf("expected OneToOneType, got %s", f.FieldType)
	}
	if !f.Unique {
		t.Error("expected OneToOne to be unique")
	}
}

func TestManyToManyField(t *testing.T) {
	f := ManyToManyField("tags", "blog.Tag", WithRelatedName("articles"), WithThrough("blog.ArticleTag"))
	if f.FieldType != ManyToManyType {
		t.Errorf("expected ManyToManyType, got %s", f.FieldType)
	}
	if f.Through != "blog.ArticleTag" {
		t.Errorf("expected Through 'blog.ArticleTag', got %s", f.Through)
	}
}

func TestModelMeta(t *testing.T) {
	m := testMeta

	if m.AppLabel != "test" {
		t.Errorf("expected AppLabel 'test', got %s", m.AppLabel)
	}
	if m.ModelName != "Widget" {
		t.Errorf("expected ModelName 'Widget', got %s", m.ModelName)
	}
	if m.TableName != "test_widget" {
		t.Errorf("expected TableName 'test_widget', got %s", m.TableName)
	}
	if m.PKField != "ID" {
		t.Errorf("expected PKField 'ID', got %s", m.PKField)
	}
	if len(m.Fields) != 7 {
		t.Errorf("expected 7 fields, got %d", len(m.Fields))
	}
}

func TestModelMetaFieldByName(t *testing.T) {
	f, ok := testMeta.FieldByName("Name")
	if !ok {
		t.Fatal("expected field 'Name' to exist")
	}
	if f.FieldType != CharFieldType {
		t.Errorf("expected CharFieldType, got %s", f.FieldType)
	}
	if f.MaxLength != 100 {
		t.Errorf("expected MaxLength 100, got %d", f.MaxLength)
	}

	_, ok = testMeta.FieldByName("NonExistent")
	if ok {
		t.Error("expected non-existent field to not be found")
	}
}

func TestModelMetaDBColumnForField(t *testing.T) {
	if got := GoFieldToDBColumn("ID"); got != "id" {
		t.Errorf("expected GoFieldToDBColumn(ID) to be 'id', got '%s'", got)
	}
	if got := GoFieldToDBColumn("AuthorID"); got != "author_id" {
		t.Errorf("expected GoFieldToDBColumn(AuthorID) to be 'author_id', got '%s'", got)
	}

	col := testMeta.DBColumnForField("ID")
	if col != "id" {
		t.Errorf("expected 'id', got '%s'", col)
	}

	col = testMeta.DBColumnForField("CreatedAt")
	if col != "created_at" {
		t.Errorf("expected 'created_at', got '%s'", col)
	}

	col = testMeta.DBColumnForField("Name")
	if col != "name" {
		t.Errorf("expected 'name' for field without explicit DBColumn, got '%s'", col)
	}

	col = testMeta.DBColumnForField("author_id")
	if col != "author_id" {
		t.Errorf("expected DB column lookup to preserve 'author_id', got '%s'", col)
	}
}

func TestModelMetaIsRelation(t *testing.T) {
	if !testMeta.IsRelation("Author") {
		t.Error("expected Author to be a relation")
	}
	if testMeta.IsRelation("Name") {
		t.Error("expected Name to not be a relation")
	}
}

func TestModelMetaConcreteFields(t *testing.T) {
	cf := testMeta.ConcreteFields()
	for _, f := range cf {
		if f.FieldType == ManyToManyType {
			t.Error("concrete fields should not include ManyToMany")
		}
	}
}

func TestModelMetaRelations(t *testing.T) {
	rels := testMeta.Relations()
	if len(rels) != 1 {
		t.Errorf("expected 1 relation, got %d", len(rels))
	}
	if rels[0].Name != "Author" {
		t.Errorf("expected relation 'Author', got '%s'", rels[0].Name)
	}
}

func TestModelRegistry(t *testing.T) {
	reg := NewModelRegistry()
	meta := &ModelMeta{
		AppLabel:  "app1",
		ModelName: "TestModel",
		Fields: []FieldDef{
			AutoField("id"),
			CharField("name", 100),
		},
	}
	reg.Register("app1", "TestModel", meta)

	m, ok := reg.Get("app1", "TestModel")
	if !ok {
		t.Fatal("expected model to be registered")
	}
	if m.ModelName != "TestModel" {
		t.Errorf("expected ModelName 'TestModel', got %s", m.ModelName)
	}

	_, ok = reg.Get("app1", "NonExistent")
	if ok {
		t.Error("expected non-existent model to not be found")
	}
}

func TestModelRegistryDefaultTableName(t *testing.T) {
	reg := NewModelRegistry()
	meta := &ModelMeta{
		AppLabel:  "blog",
		ModelName: "Post",
		Fields: []FieldDef{
			AutoField("id"),
		},
	}
	reg.Register("blog", "Post", meta)

	m, _ := reg.Get("blog", "Post")
	if m.TableName != "blog_post" {
		t.Errorf("expected default table name 'blog_post', got '%s'", m.TableName)
	}
}

func TestModelRegistryDefaultManager(t *testing.T) {
	reg := NewModelRegistry()
	meta := &ModelMeta{
		AppLabel:  "app1",
		ModelName: "Item",
		Fields: []FieldDef{
			AutoField("id"),
		},
	}
	reg.Register("app1", "Item", meta)

	m, _ := reg.Get("app1", "Item")
	if _, ok := m.Managers["objects"]; !ok {
		t.Error("expected default 'objects' manager to be created")
	}
}

func TestModelRegistryDuplicate(t *testing.T) {
	reg := NewModelRegistry()
	meta := &ModelMeta{
		AppLabel:  "app1",
		ModelName: "Dup",
		Fields:    []FieldDef{AutoField("id")},
	}
	reg.Register("app1", "Dup", meta)

	// Verify the model is registered
	m, ok := reg.Get("app1", "Dup")
	if !ok {
		t.Error("expected model to be registered")
	}
	if m.ModelName != "Dup" {
		t.Errorf("expected 'Dup', got '%s'", m.ModelName)
	}

	// Re-registering same model overwrites it
	meta2 := &ModelMeta{
		AppLabel:  "app1",
		ModelName: "Dup2",
		Fields:    []FieldDef{AutoField("id")},
	}
	reg.Register("app1", "Dup", meta2)
	m, ok = reg.Get("app1", "Dup")
	if !ok {
		t.Error("expected model to still be registered after overwrite")
	}
	if m.ModelName != "Dup2" {
		t.Errorf("expected 'Dup2' after overwrite, got '%s'", m.ModelName)
	}
}

func TestGlobalRegistry(t *testing.T) {
	reg := GlobalRegistry()
	if reg == nil {
		t.Error("expected global registry to not be nil")
	}
}

func TestInferDBType(t *testing.T) {
	tests := []struct {
		field    FieldDef
		expected string
	}{
		{AutoField("id"), "SERIAL"},
		{BigAutoField("id"), "BIGSERIAL"},
		{CharField("name", 100), "VARCHAR(100)"},
		{TextField("body"), "TEXT"},
		{IntegerField("count"), "INTEGER"},
		{BigIntegerField("count"), "BIGINT"},
		{FloatField("score"), "REAL"},
		{BooleanField("active"), "BOOLEAN"},
		{DateField("date"), "DATE"},
		{DateTimeField("created"), "TIMESTAMPTZ"},
		{UUIDField("id"), "UUID"},
		{JSONField("data"), "JSONB"},
		{ForeignKey("author", "auth.User"), "BIGINT"},
		{OneToOneField("profile", "auth.User"), "BIGINT"},
	}

	for _, tt := range tests {
		result := InferDBType(tt.field)
		if result != tt.expected {
			t.Errorf("InferDBType(%s %s): expected %s, got %s", tt.field.Name, tt.field.FieldType, tt.expected, result)
		}
	}
}

func TestLookupParsing(t *testing.T) {
	tests := []struct {
		input    string
		expOp    string
		expField string
	}{
		{"name", "exact", "name"},
		{"name__exact", "exact", "name"},
		{"name__contains", "contains", "name"},
		{"name__icontains", "icontains", "name"},
		{"name__gt", "gt", "name"},
		{"name__gte", "gte", "name"},
		{"name__lt", "lt", "name"},
		{"name__lte", "lte", "name"},
		{"name__in", "in", "name"},
		{"name__isnull", "isnull", "name"},
		{"name__startswith", "startswith", "name"},
		{"name__istartswith", "istartswith", "name"},
		{"name__endswith", "endswith", "name"},
		{"name__iendswith", "iendswith", "name"},
		{"name__regex", "regex", "name"},
		{"name__iregex", "iregex", "name"},
		{"name__range", "range", "name"},
		{"created__year", "year", "created"},
		{"created__month", "month", "created"},
		{"created__day", "day", "created"},
		{"name__search", "search", "name"},
	}

	for _, tt := range tests {
		op, field := parseLookup(tt.input)
		if op != tt.expOp {
			t.Errorf("parseLookup(%q): expected op %q, got %q", tt.input, tt.expOp, op)
		}
		if field != tt.expField {
			t.Errorf("parseLookup(%q): expected field %q, got %q", tt.input, tt.expField, field)
		}
	}
}

func TestQObjects(t *testing.T) {
	q1 := Q(L("name__contains", "test"))
	if len(q1.Lookups) != 1 {
		t.Errorf("expected 1 lookup, got %d", len(q1.Lookups))
	}
	if q1.Lookups[0].Field != "name" {
		t.Errorf("expected field 'name', got %s", q1.Lookups[0].Field)
	}
	if q1.Lookups[0].Op != "contains" {
		t.Errorf("expected op 'contains', got %s", q1.Lookups[0].Op)
	}

	q2 := QAnd(q1, Q(L("age__gt", 18)))
	if len(q2.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(q2.Children))
	}
	if q2.Connector != AND {
		t.Errorf("expected AND connector, got %s", q2.Connector)
	}

	q3 := QOr(q1, Q(L("status", "active")))
	if q3.Connector != OR {
		t.Errorf("expected OR connector, got %s", q3.Connector)
	}

	q4 := QNot(q1)
	if !q4.Negated {
		t.Error("expected Negated to be true")
	}
}

func TestFExpressions(t *testing.T) {
	f := F("count")
	if f.Field != "count" {
		t.Errorf("expected field 'count', got %s", f.Field)
	}
}

func TestAggregateExpressions(t *testing.T) {
	count := Count(F("id"))
	if count.Function != "COUNT" {
		t.Errorf("expected function 'COUNT', got %s", count.Function)
	}

	sum := Sum(F("price"))
	if sum.Function != "SUM" {
		t.Errorf("expected function 'SUM', got %s", sum.Function)
	}

	avg := Avg(F("score"))
	if avg.Function != "AVG" {
		t.Errorf("expected function 'AVG', got %s", avg.Function)
	}

	max := Max(F("created_at"))
	if max.Function != "MAX" {
		t.Errorf("expected function 'MAX', got %s", max.Function)
	}

	min := Min(F("created_at"))
	if min.Function != "MIN" {
		t.Errorf("expected function 'MIN', got %s", min.Function)
	}
}

func TestAggregateWithOptions(t *testing.T) {
	count := Count(F("id"), WithDistinct())
	if !count.Distinct {
		t.Error("expected Distinct to be true")
	}

	sum := Sum(F("price"), WithFilter(Q(L("active", true))))
	if sum.Filter == nil {
		t.Error("expected Filter to be set")
	}
}

func TestCaseExpr(t *testing.T) {
	caseExpr := Case(
		WhenClause{Condition: Q(L("age__gte", 18)), Result: ValueExpr{Value: "adult"}},
		WhenClause{Condition: Q(L("age__lt", 18)), Result: ValueExpr{Value: "minor"}},
	).Default(ValueExpr{Value: "unknown"})

	if caseExpr.ExprType() != "case" {
		t.Errorf("expected 'case', got %s", caseExpr.ExprType())
	}
	if len(caseExpr.Conditions) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(caseExpr.Conditions))
	}
}

func TestQuerySetImmutability(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)

	filtered := qs.Filter(L("name", "test"))
	if len(qs.filters) != 0 {
		t.Error("original queryset should not be modified by Filter")
	}
	if len(filtered.filters) != 1 {
		t.Error("cloned queryset should have the filter")
	}

	excluded := qs.Exclude(L("name", "test"))
	if len(qs.exclude) != 0 {
		t.Error("original queryset should not be modified by Exclude")
	}
	if len(excluded.exclude) != 1 {
		t.Error("cloned queryset should have the exclude")
	}

	ordered := qs.OrderBy("name", "-created_at")
	if len(qs.orderBy) != 0 {
		t.Error("original queryset should not be modified by OrderBy")
	}
	if len(ordered.orderBy) != 2 {
		t.Error("cloned queryset should have the ordering")
	}
}

func TestQuerySetChaining(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)

	result := qs.Filter(L("name", "test")).Exclude(L("active", false)).OrderBy("-created_at").Limit(10).Offset(20)

	if len(result.filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(result.filters))
	}
	if len(result.exclude) != 1 {
		t.Errorf("expected 1 exclude, got %d", len(result.exclude))
	}
	if len(result.orderBy) != 1 {
		t.Errorf("expected 1 order, got %d", len(result.orderBy))
	}
	if result.limit != 10 {
		t.Errorf("expected limit 10, got %d", result.limit)
	}
	if result.offset != 20 {
		t.Errorf("expected offset 20, got %d", result.offset)
	}
}

func TestQuerySetDistinct(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)
	distinct := qs.Distinct()
	if !distinct.distinct {
		t.Error("expected distinct to be true")
	}
	if qs.distinct {
		t.Error("original should not have distinct")
	}
}

func TestQuerySetOnlyDefer(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)

	only := qs.Only("name", "created_at")
	if len(only.onlyFields) != 2 {
		t.Errorf("expected 2 only fields, got %d", len(only.onlyFields))
	}

	deferred := qs.Defer("description")
	if len(deferred.deferFields) != 1 {
		t.Errorf("expected 1 defer field, got %d", len(deferred.deferFields))
	}
}

func TestQuerySetSelectRelated(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)
	related := qs.SelectRelated("author")
	if len(related.selectRelated) != 1 {
		t.Errorf("expected 1 select_related field, got %d", len(related.selectRelated))
	}
	if related.selectRelated[0] != "author" {
		t.Errorf("expected 'author', got %s", related.selectRelated[0])
	}
}

func TestQuerySetPrefetchRelated(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)
	prefetched := qs.PrefetchRelated("tags")
	if len(prefetched.prefetchRelated) != 1 {
		t.Errorf("expected 1 prefetch_related field, got %d", len(prefetched.prefetchRelated))
	}
	if prefetched.prefetchRelated[0].Field != "tags" {
		t.Errorf("expected 'tags', got %s", prefetched.prefetchRelated[0].Field)
	}
}

func TestQuerySetAnnotate(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)
	annotated := qs.Annotate(
		Annotate("total", Count(F("id"))),
		Annotate("avg_score", Avg(F("score"))),
	)
	if len(annotated.annotations) != 2 {
		t.Errorf("expected 2 annotations, got %d", len(annotated.annotations))
	}
}

func TestQuerySetForUpdate(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)
	locked := qs.ForUpdate()
	if !locked._forUpdate {
		t.Error("expected forUpdate to be true")
	}

	lockedNowait := qs.ForUpdate(NoWait)
	if !lockedNowait._forUpdateNoWait {
		t.Error("expected forUpdateNoWait to be true")
	}

	lockedSkip := qs.ForUpdate(SkipLocked)
	if !lockedSkip._forUpdateSkipLocked {
		t.Error("expected forUpdateSkipLocked to be true")
	}
}

func TestQuerySetNone(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)
	none := qs.None()
	if !none.noop {
		t.Error("expected noop to be true")
	}
}

func TestQuerySetValues(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)
	vqs := qs.Values("name", "count")
	if len(vqs.valuesFields) != 2 {
		t.Errorf("expected 2 values fields, got %d", len(vqs.valuesFields))
	}
}

func TestQuerySetValuesList(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil)
	vlqs := qs.ValuesList("name", "count")
	if len(vlqs.valuesListFields) != 2 {
		t.Errorf("expected 2 values_list fields, got %d", len(vlqs.valuesListFields))
	}
}

func TestSQLCompilerSelectBasic(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil)

	sql, args := compiler.CompileSelect(qs)
	if sql == "" {
		t.Error("expected non-empty SQL")
	}
	if !strings.Contains(sql, "SELECT") {
		t.Error("expected SELECT in SQL")
	}
	if !strings.Contains(sql, "FROM") {
		t.Error("expected FROM in SQL")
	}
	if !strings.Contains(sql, `"test_widget"`) {
		t.Errorf("expected table name in SQL, got: %s", sql)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args, got %d", len(args))
	}
}

func TestSQLCompilerSelectWithFilter(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).Filter(L("name", "test"))

	sql, args := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "WHERE") {
		t.Error("expected WHERE in SQL")
	}
	if len(args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(args))
	}
}

func TestSQLCompilerSelectWithMultipleFilters(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).Filter(L("name", "test"), L("active", true))

	sql, args := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "WHERE") {
		t.Error("expected WHERE in SQL")
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestSQLCompilerSelectWithExclude(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).Exclude(L("name", "test"))

	sql, _ := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "NOT") {
		t.Error("expected NOT in SQL for exclude")
	}
}

func TestSQLCompilerSelectWithOrderBy(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).OrderBy("name", "-created_at")

	sql, _ := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "ORDER BY") {
		t.Error("expected ORDER BY in SQL")
	}
}

func TestSQLCompilerSelectWithLimitOffset(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).Limit(10).Offset(20)

	sql, _ := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "LIMIT 10") {
		t.Error("expected LIMIT 10 in SQL")
	}
	if !strings.Contains(sql, "OFFSET 20") {
		t.Error("expected OFFSET 20 in SQL")
	}
}

func TestSQLCompilerSelectDistinct(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).Distinct()

	sql, _ := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "DISTINCT") {
		t.Error("expected DISTINCT in SQL")
	}
}

func TestSQLCompilerInsert(t *testing.T) {
	compiler := NewSQLCompiler()
	values := map[string]interface{}{
		"name":  "test widget",
		"count": 5,
	}

	sql, args := compiler.CompileInsert(testMeta, values)
	if !strings.Contains(sql, "INSERT INTO") {
		t.Error("expected INSERT INTO in SQL")
	}
	if !strings.Contains(sql, "VALUES") {
		t.Error("expected VALUES in SQL")
	}
	if !strings.Contains(sql, "RETURNING") {
		t.Error("expected RETURNING for auto PK in SQL")
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestSQLCompilerUpdate(t *testing.T) {
	compiler := NewSQLCompiler()
	values := map[string]interface{}{
		"name":  "updated",
		"count": 10,
	}

	sql, args := compiler.CompileUpdate(testMeta, values, nil, nil)
	if !strings.Contains(sql, "UPDATE") {
		t.Error("expected UPDATE in SQL")
	}
	if !strings.Contains(sql, "SET") {
		t.Error("expected SET in SQL")
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}
}

func TestSQLCompilerDelete(t *testing.T) {
	compiler := NewSQLCompiler()

	sql, _ := compiler.CompileDelete(testMeta, []QNode{Q(L("id", 1))}, nil)
	if !strings.Contains(sql, "DELETE FROM") {
		t.Error("expected DELETE FROM in SQL")
	}
	if !strings.Contains(sql, "WHERE") {
		t.Error("expected WHERE in SQL")
	}
}

func TestSQLCompilerCount(t *testing.T) {
	compiler := NewSQLCompiler()

	sql, _ := compiler.CompileCount(testMeta, nil, nil, false)
	if !strings.Contains(sql, "SELECT COUNT(") {
		t.Error("expected SELECT COUNT in SQL")
	}
	if strings.Contains(sql, `"ID"`) {
		t.Errorf("expected count SQL to use DB column, got: %s", sql)
	}
	if !strings.Contains(sql, `"id"`) {
		t.Errorf("expected count SQL to include id column, got: %s", sql)
	}
}

func TestSQLCompilerUsesDBColumns(t *testing.T) {
	compiler := NewSQLCompiler()

	insertSQL, _ := compiler.CompileInsert(testMeta, map[string]interface{}{"Name": "test"})
	if !strings.Contains(insertSQL, `"name"`) || !strings.Contains(insertSQL, `RETURNING "id"`) {
		t.Errorf("expected insert SQL to use DB columns, got: %s", insertSQL)
	}
	if strings.Contains(insertSQL, `"Name"`) || strings.Contains(insertSQL, `"ID"`) {
		t.Errorf("expected insert SQL not to use Go field names, got: %s", insertSQL)
	}

	selectSQL, _ := compiler.CompileSelect(NewBaseQuerySet(testMeta, nil).
		Filter(L("CreatedAt__year", 2026)).
		OrderBy("-CreatedAt"))
	if !strings.Contains(selectSQL, `"created_at"`) {
		t.Errorf("expected select SQL to use created_at column, got: %s", selectSQL)
	}
	if strings.Contains(selectSQL, `"CreatedAt"`) {
		t.Errorf("expected select SQL not to use CreatedAt field name, got: %s", selectSQL)
	}
}

func TestSQLCompilerLookupOperators(t *testing.T) {
	compiler := NewSQLCompiler()

	tests := []struct {
		lookup Lookup
		op     string
	}{
		{L("name", "test"), "exact"},
		{L("name__contains", "test"), "contains"},
		{L("name__icontains", "test"), "icontains"},
		{L("name__startswith", "test"), "startswith"},
		{L("name__istartswith", "test"), "istartswith"},
		{L("name__endswith", "test"), "endswith"},
		{L("name__iendswith", "test"), "iendswith"},
		{L("name__regex", "test"), "regex"},
		{L("name__iregex", "test"), "iregex"},
		{L("count__gt", 5), "gt"},
		{L("count__gte", 5), "gte"},
		{L("count__lt", 5), "lt"},
		{L("count__lte", 5), "lte"},
		{L("name__isnull", true), "isnull"},
	}

	for _, tt := range tests {
		_, args := compiler.compileLookup(testMeta, tt.lookup)
		if tt.lookup.Op != tt.op {
			t.Errorf("lookup %v: expected op %s, got %s", tt.lookup, tt.op, tt.lookup.Op)
		}
		_ = args
	}
}

func TestSQLCompilerQNode(t *testing.T) {
	compiler := NewSQLCompiler()

	q1 := Q(L("name", "test"))
	q2 := Q(L("active", true))
	andQ := QAnd(q1, q2)

	sql, args := compiler.compileQNode(testMeta, andQ)
	if !strings.Contains(sql, "AND") {
		t.Error("expected AND in SQL")
	}
	if len(args) != 2 {
		t.Errorf("expected 2 args, got %d", len(args))
	}

	orQ := QOr(q1, q2)
	sql, args = compiler.compileQNode(testMeta, orQ)
	if !strings.Contains(sql, "OR") {
		t.Error("expected OR in SQL")
	}

	notQ := QNot(q1)
	sql, args = compiler.compileQNode(testMeta, notQ)
	if !strings.Contains(sql, "NOT") {
		t.Error("expected NOT in SQL")
	}
}

func TestSQLCompilerAnnotate(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).Annotate(
		Annotate("total", CountStar()),
		Annotate("average", Avg(F("count"))),
	)

	sql, _ := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "COUNT") {
		t.Error("expected COUNT in annotated SQL")
	}
	if !strings.Contains(sql, "AVG") {
		t.Error("expected AVG in annotated SQL")
	}
}

func TestSQLCompilerForUpdate(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).ForUpdate()

	sql, _ := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "FOR UPDATE") {
		t.Error("expected FOR UPDATE in SQL")
	}
}

func TestSQLCompilerForUpdateNoWait(t *testing.T) {
	compiler := NewSQLCompiler()
	qs := NewBaseQuerySet(testMeta, nil).ForUpdate(NoWait)

	sql, _ := compiler.CompileSelect(qs)
	if !strings.Contains(sql, "FOR UPDATE NOWAIT") {
		t.Error("expected FOR UPDATE NOWAIT in SQL")
	}
}

func TestDBConfig(t *testing.T) {
	cfg := DefaultDBConfig()
	if cfg.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("expected Port 5432, got %d", cfg.Port)
	}
	if cfg.MaxConns != 10 {
		t.Errorf("expected MaxConns 10, got %d", cfg.MaxConns)
	}

	dsn := cfg.DSN()
	if !strings.Contains(dsn, "host=localhost") {
		t.Error("expected host in DSN")
	}
	if !strings.Contains(dsn, "sslmode=prefer") {
		t.Error("expected sslmode in DSN")
	}
}

func TestDBConfigFromSettings(t *testing.T) {
	cfg := DBConfigFromSettings("localhost", 5432, "mydb", "user", "pass")
	if cfg.Name != "mydb" {
		t.Errorf("expected Name 'mydb', got %s", cfg.Name)
	}
}

func TestOnDeleteConstants(t *testing.T) {
	if Cascade != "cascade" {
		t.Errorf("expected Cascade 'cascade', got %s", Cascade)
	}
	if Protect != "protect" {
		t.Errorf("expected Protect 'protect', got %s", Protect)
	}
	if SetNull != "set_null" {
		t.Errorf("expected SetNull 'set_null', got %s", SetNull)
	}
	if SetDefault != "set_default" {
		t.Errorf("expected SetDefault 'set_default', got %s", SetDefault)
	}
	if DoNothing != "do_nothing" {
		t.Errorf("expected DoNothing 'do_nothing', got %s", DoNothing)
	}
	if Restrict != "restrict" {
		t.Errorf("expected Restrict 'restrict', got %s", Restrict)
	}
}

func TestErrorTypes(t *testing.T) {
	dne := &DoesNotExist{ModelName: "Widget", Filter: "id=1"}
	if !strings.Contains(dne.Error(), "Widget") {
		t.Error("DoesNotExist should contain model name")
	}
	if !strings.Contains(dne.Error(), "does not exist") {
		t.Error("DoesNotExist should contain 'does not exist'")
	}

	mor := &MultipleObjectsReturned{ModelName: "Widget", Filter: "name=test"}
	if !strings.Contains(mor.Error(), "multiple") {
		t.Error("MultipleObjectsReturned should contain 'multiple'")
	}

	ve := &ValidationError{Field: "name", Message: "required"}
	if !strings.Contains(ve.Error(), "name") {
		t.Error("ValidationError should contain field name")
	}

	ie := &IntegrityError{Constraint: "unique_name", Message: "duplicate"}
	if !strings.Contains(ie.Error(), "unique_name") {
		t.Error("IntegrityError should contain constraint")
	}
}

func TestIndexDef(t *testing.T) {
	idx := IndexDef{
		Name:      "test_idx",
		Fields:    []string{"name", "created_at"},
		Unique:    true,
		Condition: "active = true",
	}
	if idx.Name != "test_idx" {
		t.Errorf("expected Name 'test_idx', got %s", idx.Name)
	}
	if !idx.Unique {
		t.Error("expected Unique to be true")
	}
	if idx.Condition != "active = true" {
		t.Errorf("expected Condition 'active = true', got %s", idx.Condition)
	}
}

func TestConstraintDef(t *testing.T) {
	c := ConstraintDef{
		Name:      "unique_name",
		Unique:    []string{"name", "app_label"},
		Condition: "active = true",
	}
	if c.Name != "unique_name" {
		t.Errorf("expected Name 'unique_name', got %s", c.Name)
	}
	if len(c.Unique) != 2 {
		t.Errorf("expected 2 unique fields, got %d", len(c.Unique))
	}
}

func TestModelOptions(t *testing.T) {
	opts := DefaultModelOptions()
	if !opts.Managed {
		t.Error("expected Managed to be true")
	}
	if opts.DefaultManagerName != "objects" {
		t.Errorf("expected DefaultManagerName 'objects', got %s", opts.DefaultManagerName)
	}
}

func TestSubqueryExpr(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil).Filter(L("active", true))
	subq := Subquery(qs)
	if subq.ExprType() != "subquery" {
		t.Errorf("expected 'subquery', got %s", subq.ExprType())
	}
}

func TestExistsExpr(t *testing.T) {
	qs := NewBaseQuerySet(testMeta, nil).Filter(L("active", true))
	exists := Exists(qs)
	if exists.ExprType() != "exists" {
		t.Errorf("expected 'exists', got %s", exists.ExprType())
	}
}

func TestFuncExpr(t *testing.T) {
	f := Func("LOWER", ValueExpr{Value: "test"})
	if f.Name != "LOWER" {
		t.Errorf("expected 'LOWER', got %s", f.Name)
	}
	if f.ExprType() != "func" {
		t.Errorf("expected 'func', got %s", f.ExprType())
	}
}

func TestRawExpr(t *testing.T) {
	r := Raw("NOW()", 1, 2)
	if r.SQL != "NOW()" {
		t.Errorf("expected 'NOW()', got %s", r.SQL)
	}
	if len(r.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(r.Args))
	}
}

func TestValueExpr(t *testing.T) {
	v := Value(42)
	if v.ExprType() != "value" {
		t.Errorf("expected 'value', got %s", v.ExprType())
	}
}

func TestCoalesceExpr(t *testing.T) {
	c := Coalesce(F("name"), Value("default"))
	if c.ExprType() != "coalesce" {
		t.Errorf("expected 'coalesce', got %s", c.ExprType())
	}
	if len(c.Exprs) != 2 {
		t.Errorf("expected 2 exprs, got %d", len(c.Exprs))
	}
}
