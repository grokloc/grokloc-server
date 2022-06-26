// Package testing provides tests for the org package
// (broken out to break import cycles)
package testing

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org/events"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	"github.com/grokloc/grokloc-server/pkg/app/state"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type OrgSuite struct {
	suite.Suite
	st *app.State
}

func (s *OrgSuite) SetupTest() {
	var err error
	s.st, err = state.New(env.Unit)
	if err != nil {
		zap.L().Fatal("setup",
			zap.Error(err),
		)
	}
}

func (s *OrgSuite) TestReadOrg() {
	replica := s.st.RandomReplica()

	// State initialization creates an org (and owner user)
	o, err := org.Read(
		context.Background(),
		s.st.RootOrg,
		replica,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), s.st.RootOrg, o.ID)
}

func (s *OrgSuite) TestReadOrgMiss() {
	replica := s.st.RandomReplica()

	_, err := org.Read(
		context.Background(),
		uuid.NewString(),
		replica,
	)
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)
}

func (s *OrgSuite) TestUpdateStatus() {
	ctx := context.Background()
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		ctx,
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	err = o.UpdateStatus(
		ctx,
		models.StatusInactive,
		s.st.Master,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, o.Meta.Status)

	o_read, err := org.Read(
		ctx,
		o.ID,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, o.Meta.Status)
	require.Equal(s.T(), models.StatusInactive, o_read.Meta.Status)
}

func (s *OrgSuite) TestUpdateOwner() {
	ctx := context.Background()
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		ctx,
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
		ctx,
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
		ctx,
		newOwner.ID,
		s.st.Master,
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)

	err = newOwner.UpdateStatus(
		ctx,
		models.StatusActive,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	// new owner is now active
	err = o.UpdateOwner(
		ctx,
		newOwner.ID,
		s.st.Master,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), o.Owner, newOwner.ID)
}

func (s *OrgSuite) TestUpdateOwnerWrongOrg() {
	ctx := context.Background()
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		ctx,
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	oOther, err := org.Create(
		ctx,
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
		ctx,
		o.Owner,
		s.st.Master,
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)
}

func (s *OrgSuite) TestUpdateOwnerMissing() {
	ctx := context.Background()
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		ctx,
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
		ctx,
		uuid.NewString(),
		s.st.Master,
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrRelatedUser, err)
}

func (s *OrgSuite) TestDuplicateInsert() {
	ctx := context.Background()
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	name := uuid.NewString()
	_, err = org.Create(
		ctx,
		name,
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	ownerPassword2, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)
	_, err = org.Create(
		ctx,
		name,             // RE-USED -> conflict
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword2,
		s.st.DBKey,
		s.st.Master,
	)
	require.Equal(s.T(), models.ErrConflict, err)
}

func (s *OrgSuite) TestCreateEvent() {
	ctx := context.Background()
	c, err := org.NewController(ctx, s.st)
	require.Nil(s.T(), err)

	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	event, err := events.NewCreate(
		ctx,
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
	)
	require.Nil(s.T(), err)

	o, err := c.Create(ctx, *event)
	require.Nil(s.T(), err)

	_, err = c.Read(ctx, o.ID)
	require.Nil(s.T(), err)
}

func (s *OrgSuite) TestUpdateOwnerEvent() {
	ctx := context.Background()
	c, err := org.NewController(ctx, s.st)
	require.Nil(s.T(), err)

	// use direct db api to create a new org
	o, err := org.Create(
		ctx,
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		uuid.NewString(), // org owner password
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	newOwnerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	// use direct db api to create a user in o
	newOwner, err := user.Create(
		ctx,
		uuid.NewString(), // display name
		uuid.NewString(), // email
		o.ID,
		newOwnerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	// use direct db api to make user active
	err = newOwner.UpdateStatus(
		ctx,
		models.StatusActive,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	updateOwnerEvent, err := events.NewUpdateOwner(
		ctx,
		o.ID,
		newOwner.ID,
	)
	require.Nil(s.T(), err)

	oUpdate, err := c.UpdateOwner(ctx, *updateOwnerEvent)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newOwner.ID, oUpdate.Owner)
}

func (s *OrgSuite) TestUpdateStatusEvent() {
	ctx := context.Background()
	c, err := org.NewController(ctx, s.st)
	require.Nil(s.T(), err)

	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	createEvent, err := events.NewCreate(
		ctx,
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
	)
	require.Nil(s.T(), err)

	o, err := c.Create(ctx, *createEvent)
	require.Nil(s.T(), err)

	_, err = events.NewUpdateStatus(
		ctx,
		o.ID,
		999, // not a valid status int
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrDisallowedValue, err)

	_, err = events.NewUpdateStatus(
		ctx,
		o.ID,
		int(models.StatusUnconfirmed), // unconfirmed not allowed as a set status
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrStatus, err)

	updateStatusEvent, err := events.NewUpdateStatus(
		ctx,
		o.ID,
		int(models.StatusInactive),
	)
	require.Nil(s.T(), err)

	oUpdate, err := c.UpdateStatus(ctx, *updateStatusEvent)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, oUpdate.Meta.Status)
}

func TestOrgSuite(t *testing.T) {
	suite.Run(t, new(OrgSuite))
}
