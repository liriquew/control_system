package suite

import (
	"fmt"
	"testing"

	"github.com/liriquew/control_system/internal/config"
)

type Suite struct {
	*testing.T
	TestConfig *config.AppConfig
}

func New(t *testing.T) *Suite {
	t.Helper()

	cfg := config.MustLoadPath("../config/test_config.yaml")

	return &Suite{
		TestConfig: &cfg,
	}
}

func (s *Suite) GetURL() string {
	return fmt.Sprintf("http://%s:%d", s.TestConfig.APIService.Host, s.TestConfig.APIService.Port)
}
