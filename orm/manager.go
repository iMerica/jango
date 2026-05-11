package orm

import "context"

type Manager struct {
	model *ModelMeta
	db    *DB
}

func NewManager(model *ModelMeta, db *DB) *Manager {
	return &Manager{model: model, db: db}
}

func (m *Manager) All() *BaseQuerySet {
	return NewBaseQuerySet(m.model, m.db).All()
}

func (m *Manager) Filter(lookups ...Lookup) *BaseQuerySet {
	return NewBaseQuerySet(m.model, m.db).Filter(lookups...)
}

func (m *Manager) Exclude(lookups ...Lookup) *BaseQuerySet {
	return NewBaseQuerySet(m.model, m.db).Exclude(lookups...)
}

func (m *Manager) OrderBy(fields ...string) *BaseQuerySet {
	return NewBaseQuerySet(m.model, m.db).OrderBy(fields...)
}

func (m *Manager) Get(ctx context.Context, lookups ...Lookup) (map[string]interface{}, error) {
	return NewBaseQuerySet(m.model, m.db).Get(ctx, lookups...)
}

func (m *Manager) Create(ctx context.Context, values map[string]interface{}) (map[string]interface{}, error) {
	return NewBaseQuerySet(m.model, m.db).Create(ctx, values)
}

func (m *Manager) GetOrCreate(ctx context.Context, defaults map[string]interface{}, lookups ...Lookup) (map[string]interface{}, bool, error) {
	return NewBaseQuerySet(m.model, m.db).GetOrCreate(ctx, defaults, lookups...)
}

func (m *Manager) Count(ctx context.Context) (int64, error) {
	return NewBaseQuerySet(m.model, m.db).Count(ctx)
}

func (m *Manager) First(ctx context.Context) (map[string]interface{}, error) {
	return NewBaseQuerySet(m.model, m.db).First(ctx)
}

func (m *Manager) Last(ctx context.Context) (map[string]interface{}, error) {
	return NewBaseQuerySet(m.model, m.db).Last(ctx)
}

func (m *Manager) None() *BaseQuerySet {
	return NewBaseQuerySet(m.model, m.db).None()
}

func (m *Manager) Using(db *DB) *BaseQuerySet {
	return NewBaseQuerySet(m.model, db)
}

func (m *Manager) SelectRelated(fields ...string) *BaseQuerySet {
	return NewBaseQuerySet(m.model, m.db).SelectRelated(fields...)
}

func (m *Manager) PrefetchRelated(field string, opts ...PrefetchOption) *BaseQuerySet {
	return NewBaseQuerySet(m.model, m.db).PrefetchRelated(field, opts...)
}

func (m *Manager) Annotate(annotations ...Annotation) *BaseQuerySet {
	return NewBaseQuerySet(m.model, m.db).Annotate(annotations...)
}

func (m *Manager) Values(fields ...string) *ValuesQuerySet {
	return NewBaseQuerySet(m.model, m.db).Values(fields...)
}

func (m *Manager) ValuesList(fields ...string) *ValuesListQuerySet {
	return NewBaseQuerySet(m.model, m.db).ValuesList(fields...)
}

func (m *Manager) BulkCreate(ctx context.Context, records []map[string]interface{}) (int64, error) {
	return NewBaseQuerySet(m.model, m.db).BulkCreate(ctx, records)
}

type CustomManager struct {
	*Manager
	getQuerySet func() *BaseQuerySet
}

func NewCustomManager(model *ModelMeta, db *DB, getQuerySet func() *BaseQuerySet) *CustomManager {
	return &CustomManager{
		Manager:     NewManager(model, db),
		getQuerySet: getQuerySet,
	}
}

func (cm *CustomManager) All() *BaseQuerySet {
	if cm.getQuerySet != nil {
		return cm.getQuerySet()
	}
	return cm.Manager.All()
}