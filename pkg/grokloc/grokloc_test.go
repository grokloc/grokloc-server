package grokloc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GrokLOCSuite struct {
	suite.Suite
}

func (s *GrokLOCSuite) TestContext() {
	ctx := WithRequestID(context.Background())
	id := CtxRequestID(ctx)
	require.NotEqual(s.T(), "", id)
}

func TestGrokLOCSuite(t *testing.T) {
	suite.Run(t, new(GrokLOCSuite))
}
