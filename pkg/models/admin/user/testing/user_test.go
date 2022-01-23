// Package testing provides tests for the user package
// (broken out to break import cycles)
package testing

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/models/admin/user"
	"github.com/grokloc/grokloc-server/pkg/state"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserSuite struct {
	suite.Suite
	st *app.State
}

func (s *UserSuite) SetupTest() {
	var err error
	s.st, err = state.New(env.Unit)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *UserSuite) TestReadUser() {
	replica := s.st.RandomReplica()

	// State initialization creates an org (and owner user)
	u, err := user.Read(context.Background(), s.st.RootUser, s.st.DBKey, replica)
	require.Nil(s.T(), err)
	require.Equal(s.T(), s.st.RootUser, u.ID)
	require.Equal(s.T(), s.st.RootUserAPISecret, u.APISecret)
	require.NotEqual(s.T(), 0, u.Meta.Ctime)
	require.NotEqual(s.T(), 0, u.Meta.Mtime)
}

func (s *UserSuite) TestReadUserMiss() {
	replica := s.st.RandomReplica()

	_, err := user.Read(context.Background(), uuid.NewString(), s.st.DBKey, replica)
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
