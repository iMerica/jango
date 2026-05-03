package checks

var deployChecks []CheckFunc

func RegisterDeploy(fn CheckFunc) {
	deployChecks = append(deployChecks, fn)
}

func RunDeploy() []CheckResult {
	var results []CheckResult
	for _, fn := range deployChecks {
		results = append(results, fn()...)
	}
	return results
}

func RunAllDeploy(results []CheckResult) []CheckResult {
	for _, r := range results {
		if r.Status == StatusErr {
			return results
		}
	}
	return results
}