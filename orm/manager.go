package orm

import "context"

type Manager struct {
	model *ModelMeta
	db    *DB
}

func NewManager(model *ModelMeta, db *DB) *Manager {
	return &Manager{model: model, db: db}
}

func (m *Manager) All() *QuerySet {
	return NewQuerySet(m.model, m.db).All()
}

func (m *Manager) Filter(lookups ...Lookup) *QuerySet {
	return NewQuerySet(m.model, m.db).Filter(lookups...)
}

func (m *Manager) Exclude(lookups ...Lookup) *QuerySet {
	return NewQuerySet(m.model, m.db).Exclude(lookups...)
}

func (m *Manager) OrderBy(fields ...string) *QuerySet {
	return NewQuerySet(m.model, m.db).OrderBy(fields...)
}

func (m *Manager) Get(ctx context.Context, lookups ...Lookup) (map[string]interface{}, error) {
	return NewQuerySet(m.model, m.db).Get(ctx, lookups...)
}

func (m *Manager) Create(ctx context.Context, values map[string]interface{}) (map[string]interface{}, error) {
	return NewQuerySet(m.model, m.db).Create(ctx, values)
}

func (m *Manager) GetOrCreate(ctx context.Context, defaults map[string]interface{}, lookups ...Lookup) (map[string]interface{}, bool, error) {
	return NewQuerySet(m.model, m.db).GetOrCreate(ctx, defaults, lookups...)
}

func (m *Manager) Count(ctx context.Context) (int64, error) {
	return NewQuerySet(m.model, m.db).Count(ctx)
}

func (m *Manager) First(ctx context.Context) (map[string]interface{}, error) {
	return NewQuerySet(m.model, m.db).First(ctx)
}

func (m *Manager) Last(ctx context.Context) (map[string]interface{}, error) {
	return NewQuerySet(m.model, m.db).Last(ctx)
}

func (m *Manager) None() *QuerySet {
	return NewQuerySet(m.model, m.db).None()
}

func (m *Manager) Using(db *DB) *QuerySet {
	return NewQuerySet(m.model, db)
}

func (m *Manager) SelectRelated(fields ...string) *QuerySet {
	return NewQuerySet(m.model, m.db).SelectRelated(fields...)
}

func (m *Manager) PrefetchRelated(field string, opts ...PrefetchOption) *QuerySet {
	return NewQuerySet(m.model, m.db).PrefetchRelated(field, opts...)
}

func (m *Manager) Annotate(annotations ...Annotation) *QuerySet {
	return NewQuerySet(m.model, m.db).Annotate(annotations...)
}

func (m *Manager) Values(fields ...string) *ValuesQuerySet {
	return NewQuerySet(m.model, m.db).Values(fields...)
}

func (m *Manager) ValuesList(fields ...string) *ValuesListQuerySet {
	return NewQuerySet(m.model, m.db).ValuesList(fields...)
}

func (m *Manager) BulkCreate(ctx context.Context, records []map[string]interface{}) (int64, error) {
	return NewQuerySet(m.model, m.db).BulkCreate(ctx, records)
}

type CustomManager struct {
	*Manager
	getQuerySet func() *QuerySet
}

func NewCustomManager(model *ModelMeta, db *DB, getQuerySet func() *QuerySet) *CustomManager {
	return &CustomManager{
		Manager:     NewManager(model, db),
		getQuerySet: getQuerySet,
	}
}

func (cm *CustomManager) All() *QuerySet {
	if cm.getQuerySet != nil {
		return cm.getQuerySet()
	}
	return cm.Manager.All()
}