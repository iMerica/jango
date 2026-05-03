package rest

type Serializer interface {
	Serialize(data interface{}) (map[string]interface{}, error)
	Deserialize(input map[string]interface{}) (interface{}, error)
	Fields() []string
}

type ViewSet interface {
	List() error
	Retrieve() error
	Create() error
	Update() error
	Destroy() error
}

type Router struct {
	prefix    string
	endpoints []Endpoint
}

type Endpoint struct {
	Path        string
	Method      string
	Handler     interface{}
	Name        string
}

func NewRouter(prefix string) *Router {
	return &Router{prefix: prefix}
}

func (r *Router) AddEndpoint(ep Endpoint) {
	r.endpoints = append(r.endpoints, ep)
}

func (r *Router) Endpoints() []Endpoint {
	return r.endpoints
}