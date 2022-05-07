package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UpdateStatusSuite struct {
	suite.Suite
}

func (s *UpdateStatusSuite) TestUnmarshalUpdateStatusEvent() {
	bs := []byte(`{"id":"i","status":3}`)
	var e UpdateStatus
	require.NoError(s.T(), json.Unmarshal(bs, &e))

	// has empty id
	bs = []byte(`{"id":"","status":3}`)
	require.Error(s.T(), json.Unmarshal(bs, &e))
}

func TestUpdateStatusSuite(t *testing.T) {
	suite.Run(t, new(UpdateStatusSuite))
}
