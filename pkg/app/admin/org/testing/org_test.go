// Package testing provides tests for the org package
// (broken out to break import cycles)
package testing

import (
	"context"
	"database/sql"
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
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
	require.Equal(s.T(), s.st.RootOrg, o.ID)
}

func (s *OrgSuite) TestReadOrgMiss() {
	replica := s.st.RandomReplica()

	_, err := org.Read(context.Background(), uuid.NewString(), replica)
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)
}

func (s *OrgSuite) TestUpdateStatus() {
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		context.Background(),
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	err = o.UpdateStatus(context.Background(), models.StatusInactive, s.st.Master)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, o.Meta.Status)

	o_read, err := org.Read(context.Background(), o.ID, s.st.RandomReplica())
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, o.Meta.Status)
	require.Equal(s.T(), models.StatusInactive, o_read.Meta.Status)
}

func (s *OrgSuite) TestUpdateOwner() {
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		context.Background(),
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	newOwnerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	newOwner, err := user.Create(
		context.Background(),
		uuid.NewString(), // display name
		uuid.NewString(), // email
		o.ID,
		newOwnerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	// will fail - user is still unconfirmed
	err = o.UpdateOwner(
		context.Background(),
		newOwner.ID,
		s.st.Master,
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)

	err = newOwner.UpdateStatus(
		context.Background(),
		models.StatusActive,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	// new owner is now active
	err = o.UpdateOwner(
		context.Background(),
		newOwner.ID,
		s.st.Master,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), o.Owner, newOwner.ID)
}

func (s *OrgSuite) TestUpdateOwnerWrongOrg() {
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		context.Background(),
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	oOther, err := org.Create(
		context.Background(),
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	// cannot make the owner of o the owner of oOther
	err = oOther.UpdateOwner(
		context.Background(),
		o.Owner,
		s.st.Master,
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)
}

func (s *OrgSuite) TestUpdateOwnerMissing() {
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		context.Background(),
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	// prospective new owner doesn't exist
	err = o.UpdateOwner(
		context.Background(),
		uuid.NewString(),
		s.st.Master,
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgSuite))
}
