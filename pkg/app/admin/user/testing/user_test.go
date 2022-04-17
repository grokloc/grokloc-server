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
	"github.com/grokloc/grokloc-server/pkg/app/admin/org"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	"github.com/grokloc/grokloc-server/pkg/app/state"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
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

func (s *UserSuite) TestUpdateDisplayName() {
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

	u, err := user.Read(
		context.Background(),
		o.Owner,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)

	newDisplayName := uuid.NewString()
	newDisplayNameDigest := security.EncodedSHA256(newDisplayName)

	err = u.UpdateDisplayName(
		context.Background(),
		newDisplayName,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newDisplayName, u.DisplayName)
	require.Equal(s.T(), newDisplayNameDigest, u.DisplayNameDigest)

	u_read, err := user.Read(
		context.Background(),
		o.Owner,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newDisplayName, u_read.DisplayName)
	require.Equal(s.T(), newDisplayNameDigest, u_read.DisplayNameDigest)

}

func (s *UserSuite) TestUpdatePassword() {
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

	u, err := user.Read(
		context.Background(),
		o.Owner,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)

	newPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	err = u.UpdatePassword(
		context.Background(),
		newPassword,
		s.st.Master,
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newPassword, u.Password)

	u_read, err := user.Read(
		context.Background(),
		o.Owner,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), newPassword, u_read.Password)
}

func (s *UserSuite) TestUpdateStatus() {
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

	newUserPassword, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)

	u, err := user.Create(
		context.Background(),
		uuid.NewString(), // display name
		uuid.NewString(), // email
		o.ID,
		newUserPassword,
		s.st.DBKey,
		s.st.Master,
	)
	require.Nil(s.T(), err)

	err = u.UpdateStatus(context.Background(), models.StatusInactive, s.st.Master)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, u.Meta.Status)

	u_read, err := user.Read(
		context.Background(),
		u.ID,
		s.st.DBKey,
		s.st.RandomReplica(),
	)
	require.Nil(s.T(), err)
	require.Equal(s.T(), models.StatusInactive, u.Meta.Status)
	require.Equal(s.T(), models.StatusInactive, u_read.Meta.Status)
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}
