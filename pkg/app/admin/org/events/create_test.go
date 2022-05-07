package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CreateSuite struct {
	suite.Suite
}

func (s *CreateSuite) TestUnmarshalCreateEvent() {
	bs := []byte(`{"name":"n",
                       "owner_display_name":"d",
                       "owner_email":"e",
                       "owner_password":"p"}`)
	var e Create
	require.NoError(s.T(), json.Unmarshal(bs, &e))

	// has empty name
	bs = []byte(`{"name":"",
                      "owner_display_name":"d",
                      "owner_email":"e",
                      "owner_password":"p"}`)
	require.Error(s.T(), json.Unmarshal(bs, &e))
}

func TestCreateSuite(t *testing.T) {
	suite.Run(t, new(CreateSuite))
}
