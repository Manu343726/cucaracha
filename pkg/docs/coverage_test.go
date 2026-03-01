package docs

import (
"testing"

"github.com/stretchr/testify/suite"
)

// BuilderCoverageTestSuite tests the builder
type BuilderCoverageTestSuite struct {
	suite.Suite
}

// TestBuilderCoverageInitialized tests that the builder initializes
func (suite *BuilderCoverageTestSuite) TestBuilderCoverageInitialized() {
	builder := NewBuilder()
	suite.NotNil(builder)
}

func TestBuilderCoverageSuite(t *testing.T) {
	suite.Run(t, new(BuilderCoverageTestSuite))
}
