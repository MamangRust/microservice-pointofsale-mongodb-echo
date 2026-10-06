package user_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserHandlerTestSuite struct {
	suite.Suite
}

func (s *UserHandlerTestSuite) TestUserApiLifecycle() {
	t := s.T()
	t.Skip("skipping: API tests require apigateway handler refactoring to use new pb sub-packages")
}

func TestUserHandlerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserHandlerTestSuite))
}
