package docs

import (
"bytes"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"github.com/stretchr/testify/suite"
)

type RendererTestSuite struct {
	suite.Suite
}

// TestEntryRendererNew tests creating a new renderer
func (suite *RendererTestSuite) TestEntryRendererNew() {
	buf := &bytes.Buffer{}
	renderer := NewEntryRenderer(buf, nil)
	suite.NotNil(renderer)
}

// TestRenderSummaryWithEntry renders summary correctly
func (suite *RendererTestSuite) TestRenderSummaryWithEntry() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Summary:   "This is a summary.",
	})

	err := renderer.RenderSummary()
	require.NoError(suite.T(), err)
}

// TestRenderDetailsWithEntry renders details correctly
func (suite *RendererTestSuite) TestRenderDetailsWithEntry() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Details:   "These are the detailed notes.",
	})

	err := renderer.RenderDetails()
	require.NoError(suite.T(), err)
}

// TestRenderExamplesWithEntry renders examples
func (suite *RendererTestSuite) TestRenderExamplesWithEntry() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Examples: []Example{
			{
				Description: "Example 1",
				Code:        "code here",
				Output:      "output here",
			},
		},
	})

	err := renderer.RenderExamples()
	require.NoError(suite.T(), err)
	content := buf.String()
	assert.Contains(suite.T(), content, "Examples")
}

// TestRenderLinksWithEntry renders links
func (suite *RendererTestSuite) TestRenderLinksWithEntry() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Links: []Link{
			{
				Target:       "pkg.OtherType",
				Relationship: RelationshipUses,
				Context:      "Used in this type",
			},
		},
	})

	err := renderer.RenderLinks()
	require.NoError(suite.T(), err)
	content := buf.String()
	assert.Contains(suite.T(), content, "Related")
}

// TestRenderFullWithCompleteEntry renders full entry
func (suite *RendererTestSuite) TestRenderFullWithCompleteEntry() {
	buf := &bytes.Buffer{}
	formatter := &DefaultLinkFormatter{}
	opts := DefaultRenderOptions(formatter)

	renderer := NewEntryRenderer(buf, opts)
	renderer.SetEntry(&DocumentationEntry{
		LocalName: "TestType",
		Summary:   "This is a summary.",
		Details:   "These are details.",
		Examples: []Example{
			{Description: "Example", Code: "code"},
		},
		Links: []Link{
			{Target: "pkg.Other", Relationship: RelationshipUses},
		},
	})

	err := renderer.RenderFull()
	require.NoError(suite.T(), err)
}

func TestRendererSuite(t *testing.T) {
	suite.Run(t, new(RendererTestSuite))
}
