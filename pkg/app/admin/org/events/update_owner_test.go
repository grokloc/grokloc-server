package events

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UpdateOwnerSuite struct {
	suite.Suite
}

func (s *UpdateOwnerSuite) TestUnmarshalUpdateOwnerEvent() {
	bs := []byte(fmt.Sprintf(`{"id":"%s","owner":"%s"}`,
		uuid.NewString(), uuid.NewString()))
	var e UpdateOwner
	require.NoError(s.T(), json.Unmarshal(bs, &e))

	// has empty id
	bs = []byte(`{"id":"","owner":"o"}`)
	require.Error(s.T(), json.Unmarshal(bs, &e))
}

func TestUpdateOwnerSuite(t *testing.T) {
	suite.Run(t, new(UpdateOwnerSuite))
}
