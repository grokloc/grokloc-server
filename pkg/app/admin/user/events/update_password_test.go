package events

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UpdatePasswordSuite struct {
	suite.Suite
}

func (s *UpdatePasswordSuite) TestUnmarshalUpdatePasswordEvent() {
	bs := []byte(fmt.Sprintf(`{"id":"%s","password":"%s"}`,
		uuid.NewString(), uuid.NewString()))
	var e UpdatePassword
	require.NoError(s.T(), json.Unmarshal(bs, &e))

	// has empty id
	bs = []byte(`{"id":"","password":"o"}`)
	require.Error(s.T(), json.Unmarshal(bs, &e))
}

func TestUpdatePasswordSuite(t *testing.T) {
	suite.Run(t, new(UpdatePasswordSuite))
}
