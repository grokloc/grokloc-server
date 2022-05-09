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
	bs := []byte(`{"display_name":"d",
                       "email":"e",
                       "org":"o",
                       "password":"p"}`)
	var e Create
	require.NoError(s.T(), json.Unmarshal(bs, &e))

	// has empty org
	bs = []byte(`{"display_name":"d",
                      "email":"e",
                      "org":"",
                      "password":"p"}`)
	require.Error(s.T(), json.Unmarshal(bs, &e))
}

func TestCreateSuite(t *testing.T) {
	suite.Run(t, new(CreateSuite))
}
