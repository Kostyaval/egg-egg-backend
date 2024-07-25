package service

import (
	"github.com/stretchr/testify/suite"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"gitlab.com/egg-be/egg-backend/internal/mocks"
	"io"
	"log"
	"os"
	"testing"
)

func TestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &Suite{})
}

type Suite struct {
	suite.Suite
	srv      *Service
	cfg      *config.Config
	dbMocks  *mocks.DBInterface
	rdbMocks *mocks.RedisInterface
}

func (s *Suite) SetupSuite() {
	var err error

	log.SetOutput(io.Discard)

	if err := os.Setenv("RUNTIME", "production"); err != nil {
		s.T().Fatal(err)
	}

	if err := os.Setenv("MONGODB_URI", "required but no need for tests"); err != nil {
		s.T().Fatal(err)
	}

	if err := os.Setenv("REDIS_URI", "required but no need for tests"); err != nil {
		s.T().Fatal(err)
	}

	if err := os.Setenv("TELEGRAM_TOKEN", "required but no need for tests"); err != nil {
		s.T().Fatal(err)
	}

	if err := os.Setenv("JWT_PRIVATE_KEY", `{"alg":"ES256","crv":"P-256","d":"oDCWjLaVOhbTdQsl-Q6MlsXRT1m2B12S2_2gOrPgPyc","key_ops":["sign","verify"],"kty":"EC","x":"pX9L5m15AKX-GohuANh_rNmIOguNJGMP8To_QM5e6Ks","y":"lRKZJTBz2TrDVxununJso42y17j81CQoNZMks7FTJYY"}`); err != nil {
		s.T().Fatal(err)
	}

	if err := os.Setenv("JWT_PUBLIC_KEY", `{"alg":"ES256","crv":"P-256","key_ops":["verify"],"kty":"EC","x":"pX9L5m15AKX-GohuANh_rNmIOguNJGMP8To_QM5e6Ks","y":"lRKZJTBz2TrDVxununJso42y17j81CQoNZMks7FTJYY"}`); err != nil {
		s.T().Fatal(err)
	}

	if err := os.Setenv("RULES_PATH", "../../rules.yml"); err != nil {
		s.T().Fatal(err)
	}

	s.cfg, err = config.NewConfig()
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *Suite) SetupTest() {
	s.dbMocks = mocks.NewDBInterface(s.T())
	s.rdbMocks = mocks.NewRedisInterface(s.T())
	s.srv = NewService(s.cfg, s.dbMocks, s.rdbMocks)
}
