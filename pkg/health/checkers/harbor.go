package checkers

import (
	harbor "github.com/Gskill75/api2/pkg/harbor/client"
)

type HarborChecker struct {
	NameStr string
	Client  *harbor.Client // Ton type exact ici
}

func (h HarborChecker) Name() string {
	return "harbor"
}

func (h HarborChecker) Check() error {
	return h.Client.Ping()
}
