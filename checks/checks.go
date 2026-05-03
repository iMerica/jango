package checks

type CheckResult struct {
	ID      string
	Status  string
	Message string
	Hint    string
}

type CheckFunc func() []CheckResult

var registry []CheckFunc

func Register(fn CheckFunc) {
	registry = append(registry, fn)
}

func RunAll() []CheckResult {
	var results []CheckResult
	for _, fn := range registry {
		results = append(results, fn()...)
	}
	return results
}

const (
	StatusOK   = "OK"
	StatusWarn  = "Warning"
	StatusErr   = "Error"
	StatusInfo  = "Info"
)

func OK(id, message string) CheckResult {
	return CheckResult{ID: id, Status: StatusOK, Message: message}
}

func Warn(id, message, hint string) CheckResult {
	return CheckResult{ID: id, Status: StatusWarn, Message: message, Hint: hint}
}

func Error(id, message, hint string) CheckResult {
	return CheckResult{ID: id, Status: StatusErr, Message: message, Hint: hint}
}

func Info(id, message string) CheckResult {
	return CheckResult{ID: id, Status: StatusInfo, Message: message}
}