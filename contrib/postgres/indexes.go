package postgres

import "github.com/iMerica/jango/orm"

func GinIndex(name string, fields []string, opts ...IndexOption) orm.IndexDef {
	idx := orm.IndexDef{
		Name:      name,
		Fields:    fields,
		Opclasses: []string{"gin_trgm_ops"},
	}
	for _, opt := range opts {
		opt(&idx)
	}
	return idx
}

func GinIndexWithOptions(name string, fields []string, opclasses []string, opts ...IndexOption) orm.IndexDef {
	idx := orm.IndexDef{
		Name:      name,
		Fields:    fields,
		Opclasses: opclasses,
	}
	for _, opt := range opts {
		opt(&idx)
	}
	return idx
}

func GistIndex(name string, fields []string, opts ...IndexOption) orm.IndexDef {
	idx := orm.IndexDef{
		Name:      name,
		Fields:    fields,
		Opclasses: []string{"gist_trgm_ops"},
	}
	for _, opt := range opts {
		opt(&idx)
	}
	return idx
}

func SpGistIndex(name string, fields []string, opts ...IndexOption) orm.IndexDef {
	idx := orm.IndexDef{
		Name:   name,
		Fields: fields,
	}
	for _, opt := range opts {
		opt(&idx)
	}
	return idx
}

func BrinIndex(name string, fields []string, opts ...IndexOption) orm.IndexDef {
	idx := orm.IndexDef{
		Name:   name,
		Fields: fields,
	}
	for _, opt := range opts {
		opt(&idx)
	}
	return idx
}

func PartialIndex(name string, fields []string, condition string, opts ...IndexOption) orm.IndexDef {
	idx := orm.IndexDef{
		Name:      name,
		Fields:    fields,
		Condition: condition,
	}
	for _, opt := range opts {
		opt(&idx)
	}
	return idx
}

func IncludeIndex(name string, fields []string, include []string, opts ...IndexOption) orm.IndexDef {
	idx := orm.IndexDef{
		Name:    name,
		Fields:  fields,
		Include: include,
	}
	for _, opt := range opts {
		opt(&idx)
	}
	return idx
}

func ConcurrentIndex(name string, fields []string, opts ...IndexOption) orm.IndexDef {
	idx := orm.IndexDef{
		Name:         name,
		Fields:       fields,
		Concurrently: true,
	}
	for _, opt := range opts {
		opt(&idx)
	}
	return idx
}

type IndexOption func(*orm.IndexDef)

func WithIndexCondition(condition string) IndexOption {
	return func(idx *orm.IndexDef) { idx.Condition = condition }
}

func WithInclude(include ...string) IndexOption {
	return func(idx *orm.IndexDef) { idx.Include = include }
}

func WithOpclasses(opclasses ...string) IndexOption {
	return func(idx *orm.IndexDef) { idx.Opclasses = opclasses }
}

func WithUnique() IndexOption {
	return func(idx *orm.IndexDef) { idx.Unique = true }
}

func WithConcurrently() IndexOption {
	return func(idx *orm.IndexDef) { idx.Concurrently = true }
}
