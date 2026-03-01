package docs

import (
"testing"

"github.com/stretchr/testify/suite"
)

type RendererFormatLinkTestSuite struct {
	suite.Suite
}

// TestDefaultLinkFormatterExists tests that link formatter is initialized
func (suite *RendererFormatLinkTestSuite) TestDefaultLinkFormatterExists() {
	formatter := &DefaultLinkFormatter{}
	suite.NotNil(formatter)
}

func TestRendererFormatLinkSuite(t *testing.T) {
	suite.Run(t, new(RendererFormatLinkTestSuite))
}
