// Package testing exists to break an import cycle -
// state imports org, therefore testing of org cannot
// be in the org pkg or a loop is created
package testing

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/models/admin/org"
	"github.com/grokloc/grokloc-server/pkg/state"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type OrgSuite struct {
	suite.Suite
	st *app.State
}

func (s *OrgSuite) SetupTest() {
	var err error
	s.st, err = state.New(env.Unit)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *OrgSuite) TestReadOrg() {
	replica := s.st.RandomReplica()

	// State initialization creates an org (and owner user)
	o, err := org.Read(context.Background(), s.st.RootOrg, replica)
	require.Nil(s.T(), err)
	require.Equal(s.T(), o.ID, s.st.RootOrg)
}

func (s *OrgSuite) TestReadOrgMiss() {
	replica := s.st.RandomReplica()

	_, err := org.Read(context.Background(), uuid.NewString(), replica)
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgSuite))
}
