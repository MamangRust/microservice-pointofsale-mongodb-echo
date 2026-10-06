package category_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CategoryApiTestSuite struct {
	suite.Suite
}

func (s *CategoryApiTestSuite) TestCategoryApiLifecycle() {
	t := s.T()
	t.Skip("skipping: API tests require apigateway handler refactoring to use new pb sub-packages")
}

func TestCategoryApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CategoryApiTestSuite))
}
