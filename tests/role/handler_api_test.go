package role_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RoleApiTestSuite struct {
	suite.Suite
}

func (s *RoleApiTestSuite) TestRoleApiLifecycle() {
	t := s.T()
	t.Skip("skipping: API tests require apigateway handler refactoring to use new pb sub-packages")
}

func TestRoleApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleApiTestSuite))
}
