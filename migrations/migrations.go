package migrations

import (
	"fmt"
	"sync"
)

type Dependency struct {
	AppLabel string
	Name     string
}

func (d Dependency) String() string {
	if d.AppLabel == "" {
		return d.Name
	}
	return d.AppLabel + "." + d.Name
}

type Migration struct {
	AppLabel     string
	Name         string
	Dependencies []Dependency
	Operations   []Operation
	Replaces     []string
	IsInitial    bool
}

func (m *Migration) Key() string {
	return m.AppLabel + "." + m.Name
}

type Operation interface {
	StateForwards(appLabel string, state *ProjectState) error
	StateBackwards(appLabel string, state *ProjectState) error
	DatabaseForwards(appLabel string, state *ProjectState, editor SchemaEditor) error
	DatabaseBackwards(appLabel string, state *ProjectState, editor SchemaEditor) error
	Describe() string
}

type migrationRegistry struct {
	mu          sync.RWMutex
	migrations  map[string][]*Migration
	migrationMap map[string]*Migration
}

var globalRegistry = &migrationRegistry{
	migrations:   make(map[string][]*Migration),
	migrationMap: make(map[string]*Migration),
}

func RegisterMigration(m Migration) {
	globalRegistry.register(&m)
}

func (r *migrationRegistry) register(m *Migration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := m.Key()
	r.migrations[m.AppLabel] = append(r.migrations[m.AppLabel], m)
	r.migrationMap[key] = m
}

func GetMigrationsForApp(appLabel string) []*Migration {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return globalRegistry.migrations[appLabel]
}

func GetAllMigrations() map[string][]*Migration {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	result := make(map[string][]*Migration, len(globalRegistry.migrations))
	for k, v := range globalRegistry.migrations {
		result[k] = v
	}
	return result
}

func GetMigration(appLabel, name string) (*Migration, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	key := appLabel + "." + name
	m, ok := globalRegistry.migrationMap[key]
	return m, ok
}

func ResetMigrations() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.migrations = make(map[string][]*Migration)
	globalRegistry.migrationMap = make(map[string]*Migration)
}

type Graph struct {
	nodes map[string]*Migration
	edges map[string][]string
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Migration),
		edges: make(map[string][]string),
	}
}

func (g *Graph) Add(m *Migration) {
	key := m.Key()
	g.nodes[key] = m
	for _, dep := range m.Dependencies {
		depKey := dep.String()
		g.edges[key] = append(g.edges[key], depKey)
	}
}

func (g *Graph) Get(appLabel, name string) (*Migration, bool) {
	key := appLabel + "." + name
	m, ok := g.nodes[key]
	return m, ok
}

func (g *Graph) All() []*Migration {
	result := make([]*Migration, 0, len(g.nodes))
	for _, m := range g.nodes {
		result = append(result, m)
	}
	return result
}

func (g *Graph) TopologicalSort() ([]*Migration, error) {
	inDegree := make(map[string]int)
	for key := range g.nodes {
		inDegree[key] = 0
	}
	for key, deps := range g.edges {
		inDegree[key] = len(deps)
	}

	var queue []*Migration
	for key, deg := range inDegree {
		if deg == 0 {
			if m, ok := g.nodes[key]; ok {
				queue = append(queue, m)
			}
		}
	}

	var result []*Migration
	for len(queue) > 0 {
		m := queue[0]
		queue = queue[1:]
		result = append(result, m)

		for key, deps := range g.edges {
			for _, dep := range deps {
				if dep == m.Key() {
					inDegree[key]--
					if inDegree[key] == 0 {
						if node, ok := g.nodes[key]; ok {
							queue = append(queue, node)
						}
					}
				}
			}
		}
	}

	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("migrations: circular dependency detected in migration graph")
	}

	return result, nil
}

func BuildGraph() *Graph {
	graph := NewGraph()
	for _, migrations := range globalRegistry.migrations {
		for _, m := range migrations {
			graph.Add(m)
		}
	}
	return graph
}