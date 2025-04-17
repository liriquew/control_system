package suite

import (
	"fmt"
	"testing"

	"github.com/liriquew/control_system/internal/lib/config"
)

type Suite struct {
	*testing.T
	TestConfig *config.AppTestConfig
}

func New(t *testing.T) *Suite {
	t.Helper()

	cfg := config.MustLoadPathTest("../config/test_config.yaml")

	return &Suite{
		TestConfig: &cfg,
	}
}

func (s *Suite) GetURL() string {
	return fmt.Sprintf("http://%s:%d", s.TestConfig.API.Host, s.TestConfig.API.Port)
}
