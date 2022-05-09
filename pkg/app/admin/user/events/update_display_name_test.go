package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UpdateDisplayNameSuite struct {
	suite.Suite
}

func (s *UpdateDisplayNameSuite) TestUnmarshalUpdateDisplayNameEvent() {
	bs := []byte(`{"id":"i","displayName":"o"}`)
	var e UpdateDisplayName
	require.NoError(s.T(), json.Unmarshal(bs, &e))

	// has empty id
	bs = []byte(`{"id":"","displayName":"o"}`)
	require.Error(s.T(), json.Unmarshal(bs, &e))
}

func TestUpdateDisplayNameSuite(t *testing.T) {
	suite.Run(t, new(UpdateDisplayNameSuite))
}
