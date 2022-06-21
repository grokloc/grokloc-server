package jwt

import (
	"context"
	"testing"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	"github.com/grokloc/grokloc-server/pkg/app/state"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/grokloc"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type JWTSuite struct {
	suite.Suite
	st *app.State
}

func (s *JWTSuite) SetupTest() {
	var err error
	s.st, err = state.New(env.Unit)
	if err != nil {
		zap.L().Fatal("setup",
			zap.Error(err),
		)
	}
}

func (s *JWTSuite) TestJWT() {
	// make a new org and user as owner
	ctx := grokloc.context.Background()
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

	claims, err := New(u.ID, u.EmailDigest, u.Org)
	require.Nil(s.T(), err)
	token := jwt_go.NewWithClaims(jwt_go.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(u.ID + string(s.st.TokenKey)))
	require.Nil(s.T(), err)
	claimsDecoded, err := Decode(u.ID, signedToken, s.st.TokenKey)
	require.Nil(s.T(), err)
	require.Equal(s.T(), u.ID, claimsDecoded.Id)
	require.Equal(s.T(), u.Org, claimsDecoded.Org)

	// wrong user
	password, err := security.DerivePassword(uuid.NewString(), s.st.Argon2Cfg)
	require.Nil(s.T(), err)
	uOther, err := user.Encrypted(
		ctx,
		uuid.NewString(),
		uuid.NewString(),
		o.ID,
		password,
		s.st.DBKey,
	)
	require.Nil(s.T(), err)
	uOther.Meta.Status = models.StatusActive
	err = uOther.Insert(context.Background(), s.st.Master)
	require.Nil(s.T(), err)
	_, err = Decode(uOther.ID, signedToken, s.st.TokenKey)
	require.Error(s.T(), err)

	// bad JWT
	_, err = Decode(u.ID, uuid.NewString(), s.st.TokenKey)
	require.Error(s.T(), err)

	// bad signing key
	otherSigningKey, err := security.MakeKey(uuid.NewString())
	require.Nil(s.T(), err)
	_, err = Decode(u.ID, signedToken, otherSigningKey)
	require.Error(s.T(), err)
}

func (s *JWTSuite) TestHeaderVal() {
	token := uuid.NewString() // it just needs to be some string
	require.Equal(s.T(), token, FromHeaderVal(ToHeaderVal(token)))
	require.Equal(s.T(), token, FromHeaderVal(token))
}

func TestJWTSuite(t *testing.T) {
	suite.Run(t, new(JWTSuite))
}
