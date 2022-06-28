package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	user_events "github.com/grokloc/grokloc-server/pkg/app/admin/user/events"
	"github.com/grokloc/grokloc-server/pkg/app/jwt"
	"github.com/grokloc/grokloc-server/pkg/security"
	"github.com/stretchr/testify/require"
)

func (s *AdminSuite) TestCreateUser() {
	// org owner
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.srv.ST.Argon2Cfg)
	require.Nil(s.T(), err)

	o, err := org.Create(
		s.ctx,
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		ownerPassword,
		s.srv.ST.DBKey,
		s.srv.ST.Master,
	)
	require.Nil(s.T(), err)

	owner, err := user.Read(
		s.ctx,
		o.Owner,
		s.srv.ST.DBKey,
		s.srv.ST.RandomReplica(),
	)
	require.Nil(s.T(), err)

	password, err := security.DerivePassword(uuid.NewString(), s.srv.ST.Argon2Cfg)
	require.Nil(s.T(), err)

	event, err := user_events.NewCreate(
		s.ctx,
		uuid.NewString(), // display name
		uuid.NewString(), // email
		o.ID,
		password,
	)
	require.Nil(s.T(), err)

	bs, err := json.Marshal(event)
	require.Nil(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, s.ts.URL+UserRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	location := resp.Header.Get("location")
	require.NotEmpty(s.T(), location)

	// duplicate
	req, err = http.NewRequest(http.MethodPost, s.ts.URL+UserRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, resp.StatusCode)

	// as org owner
	req, err = http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, owner.ID)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(owner.ID+owner.APISecret))
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	respBody, err := io.ReadAll(resp.Body)
	require.Nil(s.T(), err)
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	require.Nil(s.T(), err)

	newPassword, err := security.DerivePassword(uuid.NewString(), s.srv.ST.Argon2Cfg)
	require.Nil(s.T(), err)

	newEvent, err := user_events.NewCreate(
		s.ctx,
		uuid.NewString(), // display name
		uuid.NewString(), // email
		o.ID,
		newPassword,
	)
	require.Nil(s.T(), err)

	bs, err = json.Marshal(newEvent)
	require.Nil(s.T(), err)
	req, err = http.NewRequest(http.MethodPost, s.ts.URL+UserRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, owner.ID)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(tok.Bearer))

	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	location = resp.Header.Get("location")
	require.NotEmpty(s.T(), location)
}
