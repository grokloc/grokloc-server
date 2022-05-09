package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UpdatePasswordSuite struct {
	suite.Suite
}

func (s *UpdatePasswordSuite) TestUnmarshalUpdatePasswordEvent() {
	bs := []byte(`{"id":"i","password":"o"}`)
	var e UpdatePassword
	require.NoError(s.T(), json.Unmarshal(bs, &e))

	// has empty id
	bs = []byte(`{"id":"","password":"o"}`)
	require.Error(s.T(), json.Unmarshal(bs, &e))
}

func TestUpdatePasswordSuite(t *testing.T) {
	suite.Run(t, new(UpdatePasswordSuite))
}
