package signals

type Signal struct {
	receivers []Receiver
}

type Receiver func(sender interface{}, kwargs map[string]interface{})

func NewSignal() *Signal {
	return &Signal{}
}

func (s *Signal) Connect(fn Receiver) {
	s.receivers = append(s.receivers, fn)
}

func (s *Signal) Send(sender interface{}, kwargs map[string]interface{}) {
	for _, r := range s.receivers {
		r(sender, kwargs)
	}
}

var (
	PreSave      = NewSignal()
	PostSave     = NewSignal()
	PreDelete    = NewSignal()
	PostDelete   = NewSignal()
	RequestStart = NewSignal()
	RequestEnd    = NewSignal()
	PostMigrate  = NewSignal()
)