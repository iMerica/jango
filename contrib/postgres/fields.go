package postgres

import "github.com/iMerica/jango/orm"

func ArrayField(name string, elementType orm.FieldType, maxLength int) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		MaxLength: maxLength,
		DBComment: string(elementType) + "[]",
		Nullable:  true,
		Editable:  true,
	}
}

func IntegerArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "INTEGER[]",
		Nullable:  true,
		Editable:  true,
	}
}

func BigIntegerArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "BIGINT[]",
		Nullable:  true,
		Editable:  true,
	}
}

func SmallIntegerArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "SMALLINT[]",
		Nullable:  true,
		Editable:  true,
	}
}

func CharArrayField(name string, maxLength int) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		MaxLength: maxLength,
		DBComment: "VARCHAR[]",
		Nullable:  true,
		Editable:  true,
	}
}

func TextArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "TEXT[]",
		Nullable:  true,
		Editable:  true,
	}
}

func FloatArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "REAL[]",
		Nullable:  true,
		Editable:  true,
	}
}

func DoubleArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "DOUBLE PRECISION[]",
		Nullable:  true,
		Editable:  true,
	}
}

func BooleanArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "BOOLEAN[]",
		Nullable:  true,
		Editable:  true,
	}
}

func UUIDArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "UUID[]",
		Nullable:  true,
		Editable:  true,
	}
}

func DateArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "DATE[]",
		Nullable:  true,
		Editable:  true,
	}
}

func DateTimeArrayField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.ArrayFieldType,
		DBComment: "TIMESTAMPTZ[]",
		Nullable:  true,
		Editable:  true,
	}
}

func HStoreField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.FieldType("hstore"),
		DBComment: "HSTORE",
		Nullable:  true,
		Editable:  true,
	}
}

func JSONBField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.JSONFieldType,
		DBColumn:  name,
		DBComment: "JSONB",
		Nullable:  true,
		Editable:  true,
	}
}

func CICharField(name string, maxLength int) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.FieldType("citext"),
		MaxLength: maxLength,
		DBComment: "CITEXT",
		Nullable:  true,
		Editable:  true,
	}
}

func CITextField(name string) orm.FieldDef {
	return orm.FieldDef{
		Name:      name,
		FieldType: orm.FieldType("citext"),
		DBComment: "CITEXT",
		Nullable:  true,
		Editable:  true,
	}
}
