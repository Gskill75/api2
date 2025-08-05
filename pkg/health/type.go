package health

type ReadinessChecker interface {
	Name() string
	Check() error
}
