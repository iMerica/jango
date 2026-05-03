package orm

import (
	"fmt"
	"strings"
	"sync"
)

var globalRegistry = NewModelRegistry()

type ModelRegistry struct {
	mu      sync.RWMutex
	models  map[string]*ModelMeta
	ordered []string
}

func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]*ModelMeta),
	}
}

func GlobalRegistry() *ModelRegistry {
	return globalRegistry
}

func RegisterModel(appLabel string, modelName string, meta *ModelMeta) *ModelMeta {
	return globalRegistry.Register(appLabel, modelName, meta)
}

func (mr *ModelRegistry) Register(appLabel string, modelName string, meta *ModelMeta) *ModelMeta {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	key := appLabel + "." + modelName

	if meta.AppLabel == "" {
		meta.AppLabel = appLabel
	}
	if meta.ModelName == "" {
		meta.ModelName = modelName
	}
	if meta.TableName == "" {
		if meta.Options.TableName != "" {
			meta.TableName = meta.Options.TableName
		} else {
			meta.TableName = appLabel + "_" + strings.ToLower(modelName)
		}
	}
	if meta.PKField == "" {
		for _, f := range meta.Fields {
			if f.PrimaryKey {
				meta.PKField = f.Name
				break
			}
		}
		if meta.PKField == "" {
			meta.PKField = "id"
		}
	}
	if meta.DefaultOrdering == nil {
		if len(meta.Options.DefaultOrdering) > 0 {
			meta.DefaultOrdering = meta.Options.DefaultOrdering
		}
		meta.DefaultOrdering = meta.Options.Ordering
	}
	if meta.Managers == nil {
		meta.Managers = make(map[string]*ManagerDef)
	}
	if _, exists := meta.Managers["objects"]; !exists {
		meta.Managers["objects"] = &ManagerDef{Name: "objects"}
	}

	mr.models[key] = meta
	mr.ordered = append(mr.ordered, key)
	return meta
}

func (mr *ModelRegistry) Get(appLabel, modelName string) (*ModelMeta, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	key := appLabel + "." + modelName
	m, ok := mr.models[key]
	return m, ok
}

func (mr *ModelRegistry) MustGet(appLabel, modelName string) *ModelMeta {
	m, ok := mr.Get(appLabel, modelName)
	if !ok {
		panic(fmt.Sprintf("orm: model %s.%s not registered", appLabel, modelName))
	}
	return m
}

func (mr *ModelRegistry) AllModels() []*ModelMeta {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	result := make([]*ModelMeta, 0, len(mr.ordered))
	for _, key := range mr.ordered {
		if m, ok := mr.models[key]; ok {
			result = append(result, m)
		}
	}
	return result
}

func (mr *ModelRegistry) ModelsForApp(appLabel string) []*ModelMeta {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	var result []*ModelMeta
	for _, key := range mr.ordered {
		if m, ok := mr.models[key]; ok && m.AppLabel == appLabel {
			result = append(result, m)
		}
	}
	return result
}

func (mr *ModelRegistry) FieldForLookup(appLabel, modelName, fieldPath string) (FieldDef, bool) {
	m, ok := mr.Get(appLabel, modelName)
	if !ok {
		return FieldDef{}, false
	}
	parts := strings.SplitN(fieldPath, "__", 2)
	fieldName := parts[0]
	f, ok := m.FieldByName(fieldName)
	if !ok {
		return FieldDef{}, false
	}
	return f, true
}

func (mr *ModelRegistry) Reset() {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.models = make(map[string]*ModelMeta)
	mr.ordered = nil
}

func DefaultPKField(name string) FieldDef {
	return AutoField(name)
}

func DefaultBigPKField(name string) FieldDef {
	return BigAutoField(name)
}