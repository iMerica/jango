package postgres

import (
	"strings"
	"testing"

	"github.com/iMerica/jango/orm"
)

func TestArrayFields(t *testing.T) {
	tests := []struct {
		name     string
		field    orm.FieldDef
		ftype    orm.FieldType
		nullable bool
	}{
		{"integer_array", IntegerArrayField("tags"), orm.ArrayFieldType, true},
		{"bigint_array", BigIntegerArrayField("ids"), orm.ArrayFieldType, true},
		{"smallint_array", SmallIntegerArrayField("flags"), orm.ArrayFieldType, true},
		{"char_array", CharArrayField("names", 100), orm.ArrayFieldType, true},
		{"text_array", TextArrayField("descriptions"), orm.ArrayFieldType, true},
		{"float_array", FloatArrayField("scores"), orm.ArrayFieldType, true},
		{"double_array", DoubleArrayField("values"), orm.ArrayFieldType, true},
		{"bool_array", BooleanArrayField("flags"), orm.ArrayFieldType, true},
		{"uuid_array", UUIDArrayField("uuids"), orm.ArrayFieldType, true},
		{"date_array", DateArrayField("dates"), orm.ArrayFieldType, true},
		{"datetime_array", DateTimeArrayField("timestamps"), orm.ArrayFieldType, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.FieldType != tt.ftype {
				t.Errorf("expected FieldType %s, got %s", tt.ftype, tt.field.FieldType)
			}
			if tt.field.Nullable != tt.nullable {
				t.Errorf("expected Nullable %v, got %v", tt.nullable, tt.field.Nullable)
			}
		})
	}
}

func TestJSONBField(t *testing.T) {
	f := JSONBField("data")
	if f.FieldType != orm.JSONFieldType {
		t.Errorf("expected JSONFieldType, got %s", f.FieldType)
	}
	if !f.Nullable {
		t.Error("expected JSONB field to be nullable")
	}
}

func TestHStoreField(t *testing.T) {
	f := HStoreField("props")
	if f.FieldType != orm.FieldType("hstore") {
		t.Errorf("expected hstore FieldType, got %s", f.FieldType)
	}
}

func TestCITextField(t *testing.T) {
	f := CITextField("name")
	if f.FieldType != orm.FieldType("citext") {
		t.Errorf("expected citext FieldType, got %s", f.FieldType)
	}
	if !f.Nullable {
		t.Error("expected CIText field to be nullable")
	}
}

func TestSearchVectorField(t *testing.T) {
	f := SearchVectorField("body_search")
	if f.FieldType != orm.FieldType("searchvector") {
		t.Errorf("expected searchvector FieldType, got %s", f.FieldType)
	}
	if f.DBColumn != "body_search" {
		t.Errorf("expected DBColumn 'body_search', got %s", f.DBColumn)
	}
}

func TestSearchFunctions(t *testing.T) {
	sv := SearchVector("title", "body")
	if sv.Name != "TO_TSVECTOR" {
		t.Errorf("expected TO_TSVECTOR, got %s", sv.Name)
	}

	sq := SearchQuery("english", "test query")
	if sq.Name != "TO_TSQUERY" {
		t.Errorf("expected TO_TSQUERY, got %s", sq.Name)
	}

	psq := SearchPlainQuery("english", "test query")
	if psq.Name != "PLAINTO_TSQUERY" {
		t.Errorf("expected PLAINTO_TSQUERY, got %s", psq.Name)
	}

	phsq := SearchPhraseQuery("english", "test phrase")
	if phsq.Name != "PHRASETO_TSQUERY" {
		t.Errorf("expected PHRASETO_TSQUERY, got %s", phsq.Name)
	}

	wsq := SearchWebQuery("english", "test & web")
	if wsq.Name != "WEBSEARCH_TO_TSQUERY" {
		t.Errorf("expected WEBSEARCH_TO_TSQUERY, got %s", wsq.Name)
	}

	rank := SearchRank("body_search", "query")
	if rank.Name != "TS_RANK" {
		t.Errorf("expected TS_RANK, got %s", rank.Name)
	}

	rankCD := SearchRankCD("body_search", "query")
	if rankCD.Name != "TS_RANK_CD" {
		t.Errorf("expected TS_RANK_CD, got %s", rankCD.Name)
	}
}

func TestGinIndex(t *testing.T) {
	idx := GinIndex("test_gin_idx", []string{"body"}, WithUnique())
	if idx.Name != "test_gin_idx" {
		t.Errorf("expected name 'test_gin_idx', got %s", idx.Name)
	}
	if len(idx.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(idx.Fields))
	}
	if !idx.Unique {
		t.Error("expected Unique to be true")
	}
}

func TestGistIndex(t *testing.T) {
	idx := GistIndex("test_gist_idx", []string{"body"})
	if idx.Name != "test_gist_idx" {
		t.Errorf("expected name 'test_gist_idx', got %s", idx.Name)
	}
}

func TestPartialIndex(t *testing.T) {
	idx := PartialIndex("test_partial_idx", []string{"is_active"}, "is_active = true")
	if idx.Condition != "is_active = true" {
		t.Errorf("expected condition 'is_active = true', got %s", idx.Condition)
	}
}

func TestUniqueConstraint(t *testing.T) {
	c := UniqueConstraint("unique_name", []string{"name"}, "active = true")
	if len(c.Unique) != 1 {
		t.Errorf("expected 1 unique field, got %d", len(c.Unique))
	}
	if c.Condition != "active = true" {
		t.Errorf("expected condition 'active = true', got %s", c.Condition)
	}
}

func TestCheckConstraint(t *testing.T) {
	c := CheckConstraint("check_age", "age >= 0")
	if c.Check != "age >= 0" {
		t.Errorf("expected check 'age >= 0', got %s", c.Check)
	}
}

func TestExtensions(t *testing.T) {
	tests := []struct {
		ext  Extension
		name string
	}{
		{HStoreExtension(), "hstore"},
		{CITextExtension(), "citext"},
		{TrigramExtension(), "pg_trgm"},
		{UUIDExtension(), "uuid-ossp"},
		{CryptoExtension(), "pgcrypto"},
		{BTreeGinExtension(), "btree_gin"},
		{BTreeGistExtension(), "btree_gist"},
		{UnaccentExtension(), "unaccent"},
		{LTreeExtension(), "ltree"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ext.Name != tt.name {
				t.Errorf("expected extension name %s, got %s", tt.name, tt.ext.Name)
			}
			if !tt.ext.IfNotExists {
				t.Error("expected IfNotExists to be true")
			}
			sql := tt.ext.SQL()
			if !strings.Contains(sql, "CREATE EXTENSION IF NOT EXISTS") {
				t.Errorf("expected IF NOT EXISTS in SQL, got %s", sql)
			}
			if !strings.Contains(sql, tt.name) {
				t.Errorf("expected extension name in SQL, got %s", sql)
			}
		})
	}
}

func TestExtensionWithSchema(t *testing.T) {
	ext := CreateExtensionInSchema("hstore", "public")
	if ext.Schema != "public" {
		t.Errorf("expected schema 'public', got %s", ext.Schema)
	}
	sql := ext.SQL()
	if !strings.Contains(sql, "WITH SCHEMA") {
		t.Errorf("expected WITH SCHEMA in SQL, got %s", sql)
	}
}

func TestAggregateFunctions(t *testing.T) {
	sa := StringAgg("name", ",")
	if sa.Function != "STRING_AGG" {
		t.Errorf("expected STRING_AGG, got %s", sa.Function)
	}

	aa := ArrayAgg("name")
	if aa.Function != "ARRAY_AGG" {
		t.Errorf("expected ARRAY_AGG, got %s", aa.Function)
	}

	ba := BoolAnd("active")
	if ba.Function != "BOOL_AND" {
		t.Errorf("expected BOOL_AND, got %s", ba.Function)
	}

	bo := BoolOr("active")
	if bo.Function != "BOOL_OR" {
		t.Errorf("expected BOOL_OR, got %s", bo.Function)
	}
}

func TestWindowFunctions(t *testing.T) {
	rn := RowNumber()
	if rn.Name != "ROW_NUMBER" {
		t.Errorf("expected ROW_NUMBER, got %s", rn.Name)
	}

	lag := Lag("value", 1, nil)
	if lag.Name != "LAG" {
		t.Errorf("expected LAG, got %s", lag.Name)
	}

	lead := Lead("value", 1, 0)
	if lead.Name != "LEAD" {
		t.Errorf("expected LEAD, got %s", lead.Name)
	}
}

func TestDateFunctions(t *testing.T) {
	now := NowExpr()
	if now.Name != "NOW" {
		t.Errorf("expected NOW, got %s", now.Name)
	}

	cd := CurrentDateExpr()
	if cd.Name != "CURRENT_DATE" {
		t.Errorf("expected CURRENT_DATE, got %s", cd.Name)
	}

	ct := CurrentTimeExpr()
	if ct.Name != "CURRENT_TIME" {
		t.Errorf("expected CURRENT_TIME, got %s", ct.Name)
	}

	cts := CurrentTimestampExpr()
	if cts.Name != "CURRENT_TIMESTAMP" {
		t.Errorf("expected CURRENT_TIMESTAMP, got %s", cts.Name)
	}
}
