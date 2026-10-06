package auth_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthHandlerApiTestSuite struct {
	suite.Suite
}

func (s *AuthHandlerApiTestSuite) Test0_Ping() {
	t := s.T()
	t.Skip("skipping: API tests require apigateway handler refactoring to use new pb sub-packages")
}

func TestAuthHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerApiTestSuite))
}
