package postgres

import "github.com/iMerica/jango/orm"

func ExclusionConstraint(name string, condition string) orm.ConstraintDef {
	return orm.ConstraintDef{
		Name:  name,
		Check: condition,
	}
}

func UniqueConstraint(name string, fields []string, condition string, opts ...ConstraintOption) orm.ConstraintDef {
	c := orm.ConstraintDef{
		Name:      name,
		Unique:    fields,
		Condition: condition,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func CheckConstraint(name string, check string, opts ...ConstraintOption) orm.ConstraintDef {
	c := orm.ConstraintDef{
		Name:  name,
		Check: check,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func DeferrableConstraint(name string, deferred bool, violates string) orm.ConstraintDef {
	return orm.ConstraintDef{
		Name:       name,
		Deferrable: "DEFERRABLE",
		Deferred:   deferred,
		Violated:   violates,
	}
}

func InitiallyDeferredConstraint(name string, fields []string) orm.ConstraintDef {
	return orm.ConstraintDef{
		Name:       name,
		Unique:     fields,
		Deferrable: "DEFERRABLE",
		Deferred:   true,
	}
}

type ConstraintOption func(*orm.ConstraintDef)

func WithDeferrable() ConstraintOption {
	return func(c *orm.ConstraintDef) { c.Deferrable = "DEFERRABLE" }
}

func WithInitiallyDeferred() ConstraintOption {
	return func(c *orm.ConstraintDef) {
		c.Deferrable = "DEFERRABLE"
		c.Deferred = true
	}
}

func WithCondition(condition string) ConstraintOption {
	return func(c *orm.ConstraintDef) { c.Condition = condition }
}
