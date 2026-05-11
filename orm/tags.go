package orm

import (
	"reflect"
	"strconv"
	"strings"
)

// ParseModel generates a ModelMeta from a struct instance.
func ParseModel(appLabel, modelName string, model interface{}) *ModelMeta {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		panic("orm: ParseModel requires a struct or pointer to struct")
	}
	typ := val.Type()

	meta := &ModelMeta{
		AppLabel:  appLabel,
		ModelName: modelName,
		Fields:    []FieldDef{},
	}

	if meta.ModelName == "" {
		meta.ModelName = typ.Name()
	}

	// Check for Meta interface
	if metaProvider, ok := model.(interface{ Meta() ModelOptions }); ok {
		meta.Options = metaProvider.Meta()
	}

	hasExplicitPK := false

	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)

		// Ignore unexported fields
		if structField.PkgPath != "" {
			continue
		}

		// Parse jango tag
		tag := structField.Tag.Get("jango")
		if tag == "-" {
			continue
		}

		fieldDef := parseField(structField, tag)
		if fieldDef.PrimaryKey {
			hasExplicitPK = true
			meta.PKField = fieldDef.Name
		}
		meta.Fields = append(meta.Fields, fieldDef)
	}

	// If no explicit PK is found but there's a field named "ID", make it the PK
	if !hasExplicitPK {
		for i, f := range meta.Fields {
			if strings.EqualFold(f.Name, "ID") {
				meta.Fields[i].PrimaryKey = true
				meta.PKField = f.Name
				if f.FieldType == IntFieldType || f.FieldType == PositiveIntFieldType {
					meta.Fields[i].FieldType = AutoFieldType
					meta.Fields[i].Auto = true
					meta.Fields[i].Editable = false
				} else if f.FieldType == BigIntFieldType || f.FieldType == PositiveBigIntFieldType {
					meta.Fields[i].FieldType = BigAutoFieldType
					meta.Fields[i].Auto = true
					meta.Fields[i].Editable = false
				}
				break
			}
		}
	}

	return meta
}

func parseField(sf reflect.StructField, tag string) FieldDef {
	name := sf.Name
	colName := GoFieldToDBColumn(name)

	f := FieldDef{
		Name:     name,
		DBColumn: colName,
		Nullable: sf.Type.Kind() == reflect.Ptr,
		Editable: true,
	}

	// Infer type from Go type
	f.FieldType = inferFieldType(sf.Type)

	// Process tags
	if tag != "" {
		parts := strings.Split(tag, ",")
		for _, p := range parts {
			kv := strings.SplitN(p, ":", 2)
			key := kv[0]
			val := ""
			if len(kv) > 1 {
				val = kv[1]
			}
			switch key {
			case "primary_key":
				f.PrimaryKey = true
			case "column":
				f.DBColumn = val
			case "unique":
				f.Unique = true
			case "index":
				f.DBIndex = true
			case "null":
				f.Nullable = true
			case "auto_now":
				f.AutoNow = true
			case "auto_now_add":
				f.AutoNowAdd = true
			case "type":
				f.FieldType = FieldType(val)
			case "auto":
				f.Auto = true
				f.Editable = false
			case "related_name":
				f.RelatedName = val
			case "related_model":
				f.RelatedModel = val
			case "max_length":
				if l, err := strconv.Atoi(val); err == nil {
					f.MaxLength = l
				}
			}
		}
	}

	if f.PrimaryKey {
		if f.FieldType == IntFieldType || f.FieldType == PositiveIntFieldType {
			f.FieldType = AutoFieldType
			f.Auto = true
			f.Editable = false
		} else if f.FieldType == BigIntFieldType || f.FieldType == PositiveBigIntFieldType {
			f.FieldType = BigAutoFieldType
			f.Auto = true
			f.Editable = false
		}
	}

	// Relation defaults based on type/tags
	if sf.Type.Kind() == reflect.Ptr && sf.Type.Elem().Kind() == reflect.Struct {
		if f.RelatedModel == "" {
			// e.g. *User -> "User"
			f.RelatedModel = sf.Type.Elem().Name()
		}
		// If tag didn't specify type, assume ForeignKey
		if f.FieldType == TextFieldType {
			f.FieldType = ForeignKeyType
		}
	} else if sf.Type.Kind() == reflect.Slice && sf.Type.Elem().Kind() == reflect.Ptr {
		if f.RelatedModel == "" {
			f.RelatedModel = sf.Type.Elem().Elem().Name()
		}
		if f.FieldType == TextFieldType {
			f.FieldType = ManyToManyType
		}
	}

	if f.FieldType == ForeignKeyType || f.FieldType == OneToOneType || f.FieldType == ManyToManyType {
		if f.RelatedName == "" {
			f.RelatedName = "+" // Django default
		}
		if f.DBColumn == colName && f.FieldType != ManyToManyType {
			f.DBColumn = colName + "_id"
		}
	}

	return f
}

func inferFieldType(t reflect.Type) FieldType {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return CharFieldType
	case reflect.Int, reflect.Int32:
		return IntFieldType
	case reflect.Int64:
		return BigIntFieldType
	case reflect.Int16:
		return SmallIntFieldType
	case reflect.Uint, reflect.Uint32:
		return PositiveIntFieldType
	case reflect.Uint64:
		return PositiveBigIntFieldType
	case reflect.Uint16:
		return PositiveSmallIntFieldType
	case reflect.Bool:
		return BooleanFieldType
	case reflect.Float32:
		return FloatFieldType
	case reflect.Float64:
		return DoubleFieldType
	}

	// Handle specific structs like time.Time
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return DateTimeFieldType
	}
	if t.PkgPath() == "database/sql" {
		switch t.Name() {
		case "NullString":
			return CharFieldType
		case "NullInt64":
			return BigIntFieldType
		case "NullBool":
			return BooleanFieldType
		case "NullFloat64":
			return DoubleFieldType
		case "NullTime":
			return DateTimeFieldType
		}
	}

	return TextFieldType
}
