package orm

import "context"

func Init(ctx context.Context, dbConfig *DBConfig) (*DB, error) {
	db, err := OpenDB(ctx, dbConfig)
	if err != nil {
		return nil, err
	}
	SetDefaultDB(db)
	return db, nil
}

var defaultDB *DB

func SetDefaultDB(db *DB) {
	defaultDB = db
}

func DefaultDB() *DB {
	return defaultDB
}

func NewManagerForModel(appLabel, modelName string) (*Manager, error) {
	meta, ok := GlobalRegistry().Get(appLabel, modelName)
	if !ok {
		return nil, &DoesNotExist{ModelName: modelName, Filter: "registry lookup"}
	}
	return NewManager(meta, defaultDB), nil
}

func BaseObjects(appLabel, modelName string) *BaseQuerySet {
	meta := GlobalRegistry().MustGet(appLabel, modelName)
	return NewBaseQuerySet(meta, defaultDB)
}