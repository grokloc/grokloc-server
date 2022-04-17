package state

import (
	"testing"

	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StateSuite struct {
	suite.Suite
}

func (s *StateSuite) TestUnit() {
	_, err := New(env.Unit)
	require.Nil(s.T(), err)
}

func TestStateSuite(t *testing.T) {
	suite.Run(t, new(StateSuite))
}
