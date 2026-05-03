package postgres

type Extension struct {
	Name        string
	Schema      string
	IfNotExists bool
	Version     string
}

func CreateExtension(name string) Extension {
	return Extension{
		Name:        name,
		IfNotExists: true,
	}
}

func CreateExtensionInSchema(name, schema string) Extension {
	return Extension{
		Name:        name,
		Schema:      schema,
		IfNotExists: true,
	}
}

func HStoreExtension() Extension {
	return CreateExtension("hstore")
}

func CITextExtension() Extension {
	return CreateExtension("citext")
}

func TrigramExtension() Extension {
	return CreateExtension("pg_trgm")
}

func UUIDExtension() Extension {
	return CreateExtension("uuid-ossp")
}

func CryptoExtension() Extension {
	return CreateExtension("pgcrypto")
}

func BTreeGinExtension() Extension {
	return CreateExtension("btree_gin")
}

func BTreeGistExtension() Extension {
	return CreateExtension("btree_gist")
}

func UnaccentExtension() Extension {
	return CreateExtension("unaccent")
}

func LTreeExtension() Extension {
	return CreateExtension("ltree")
}

func PostGISExtension() Extension {
	return CreateExtension("postgis")
}

func TableFuncExtension() Extension {
	return CreateExtension("tablefunc")
}

func XMLElementExtension() Extension {
	return CreateExtension("xml2")
}

func (e Extension) SQL() string {
	sql := "CREATE EXTENSION "
	if e.IfNotExists {
		sql += "IF NOT EXISTS "
	}
	sql += `"` + e.Name + `"`
	if e.Schema != "" {
		sql += ` WITH SCHEMA "` + e.Schema + `"`
	}
	if e.Version != "" {
		sql += ` VERSION "` + e.Version + `"`
	}
	return sql + ";"
}