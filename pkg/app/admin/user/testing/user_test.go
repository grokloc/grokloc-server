// Package testing provides tests for the user package
// (broken out to break import cycles)
package testing

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user/events"
	"github.com/grokloc/grokloc-server/pkg/app/state"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type UserSuite struct {
	suite.Suite
	st *app.State
}

func (s *UserSuite) SetupTest() {
	var err error
	s.st, err = state.New(env.Unit)
	if err != nil {
		zap.L().Fatal("setup",
			zap.Error(err),
		)
	}
}

func (s *UserSuite) TestReadUser() {
	replica := s.st.RandomReplica()

	// State initialization creates an org (and owner user)
	u, err := user.Read(
		context.Background(),
		s.st.RootUser,
		s.st.DBKey,
		replica,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), s.st.RootUser, u.ID)
	require.Equal(s.T(), s.st.RootUserAPISecret, u.APISecret)
	require.NotEqual(s.T(), 0, u.Meta.Ctime)
	require.NotEqual(s.T(), 0, u.Meta.Mtime)
}

func (s *UserSuite) TestReadUserMiss() {
	replica := s.st.RandomReplica()

	_, err := user.Read(
		context.Background(),
		uuid.NewString(),
		s.st.DBKey,
		replica,
	)
	require.Error(s.T(), err)
	require.Equal(s.T(), sql.ErrNoRows, err)
}

func (s *UserSuite) TestUpdateDisplayName() {
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

	u, err := user.Read(
		ctx,
		o.Owner,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)

	newDisplayName := uuid.NewString()
	newDisplayNameDigest := security.EncodedSHA256(newDisplayName)

	err = u.UpdateDisplayName(
		ctx,
		newDisplayName,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newDisplayName, u.DisplayName)
	require.Equal(s.T(), newDisplayNameDigest, u.DisplayNameDigest)

	u_read, err := user.Read(
		ctx,
		o.Owner,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newDisplayName, u_read.DisplayName)
	require.Equal(s.T(), newDisplayNameDigest, u_read.DisplayNameDigest)
}

func (s *UserSuite) TestUpdatePassword() {
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

	u, err := user.Read(
		ctx,
		o.Owner,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)

	newPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	err = u.UpdatePassword(
		ctx,
		newPassword,
		s.st.Master,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newPassword, u.Password)

	u_read, err := user.Read(
		ctx,
		o.Owner,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newPassword, u_read.Password)
}

func (s *UserSuite) TestUpdateStatus() {
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

	newUserPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	u, err := user.Create(
		ctx,
		uuid.NewString(), // display name
		uuid.NewString(), // email
		o.ID,
		newUserPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	err = u.UpdateStatus(
		ctx,
		models.StatusInactive,
		s.st.Master,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, u.Meta.Status)

	u_read, err := user.Read(
		ctx,
		u.ID,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, u.Meta.Status)
	require.Equal(s.T(), models.StatusInactive, u_read.Meta.Status)
}

func (s *UserSuite) TestDuplicateInsert() {
	ctx := context.Background()
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	email := uuid.NewString()
	o, err := org.Create(
		ctx,
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		email,
		ownerPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	_, err = user.Read(ctx, o.Owner, s.st.DBKey, s.st.Master)
	require.Nil(s.T(), err)

	userPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)
	u, err := user.Encrypted(
		ctx,
		uuid.NewString(), // display name
		email,            // RE-USED -> conflict
		o.ID,
		userPassword,
		s.st.DBKey,
	)
	require.Nil(s.T(), err)

	err = u.Insert(ctx, s.st.Master)
	require.Equal(s.T(), models.ErrConflict, err)
}

func (s *UserSuite) TestCreateEvent() {
	ctx := context.Background()
	c, err := user.NewController(ctx, s.st)
	require.Nil(s.T(), err)

	password, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	event, err := events.NewCreate(
		ctx,
		uuid.NewString(), // display name
		uuid.NewString(), // email
		s.st.RootOrg,
		password,
	)
	require.Nil(s.T(), err)

	u, err := c.Create(ctx, *event)
	require.Nil(s.T(), err)

	_, err = c.Read(ctx, u.ID)
	require.Nil(s.T(), err)
}

func (s *UserSuite) TestUpdateDisplayNameEvent() {
	ctx := context.Background()
	c, err := user.NewController(ctx, s.st)
	require.Nil(s.T(), err)

	password, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	event, err := events.NewCreate(
		ctx,
		uuid.NewString(), // display name
		uuid.NewString(), // email
		s.st.RootOrg,
		password,
	)
	require.Nil(s.T(), err)

	u, err := c.Create(ctx, *event)
	require.Nil(s.T(), err)

	newDisplayName := uuid.NewString()

	updateDisplayNameEvent, err := events.NewUpdateDisplayName(
		ctx,
		u.ID,
		newDisplayName,
	)
	require.Nil(s.T(), err)
	uUpdate, err := c.UpdateDisplayName(ctx, *updateDisplayNameEvent)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newDisplayName, uUpdate.DisplayName)
}

func (s *UserSuite) TestUpdatePasswordEvent() {
	ctx := context.Background()
	c, err := user.NewController(ctx, s.st)
	require.Nil(s.T(), err)

	password, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	event, err := events.NewCreate(
		ctx,
		uuid.NewString(), // display name
		uuid.NewString(), // email
		s.st.RootOrg,
		password,
	)
	require.Nil(s.T(), err)

	u, err := c.Create(ctx, *event)
	require.Nil(s.T(), err)

	newPassword := uuid.NewString()

	updatePasswordEvent, err := events.NewUpdatePassword(
		ctx,
		u.ID,
		newPassword,
	)
	require.Nil(s.T(), err)
	uUpdate, err := c.UpdatePassword(ctx, *updatePasswordEvent)
	require.Nil(s.T(), err)
	match, err := security.VerifyPassword(newPassword, uUpdate.Password)
	require.Nil(s.T(), err)
	require.True(s.T(), match)
}

func (s *UserSuite) TestUpdateStatusEvent() {
	ctx := context.Background()
	c, err := user.NewController(ctx, s.st)
	require.Nil(s.T(), err)

	password, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	event, err := events.NewCreate(
		ctx,
		uuid.NewString(), // display name
		uuid.NewString(), // email
		s.st.RootOrg,
		password,
	)
	require.Nil(s.T(), err)

	u, err := c.Create(ctx, *event)
	require.Nil(s.T(), err)

	_, err = events.NewUpdateStatus(
		ctx,
		u.ID,
		999, // not a valid status int
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrDisallowedValue, err)

	_, err = events.NewUpdateStatus(
		ctx,
		u.ID,
		int(models.StatusUnconfirmed), // unconfirmed not allowed as a set status
	)
	require.NotNil(s.T(), err)
	require.Equal(s.T(), models.ErrStatus, err)

	updateStatusEvent, err := events.NewUpdateStatus(
		ctx,
		u.ID,
		int(models.StatusInactive),
	)
	require.Nil(s.T(), err)

	uUpdate, err := c.UpdateStatus(ctx, *updateStatusEvent)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, uUpdate.Meta.Status)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
