package server

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app/jwt"
	"github.com/grokloc/grokloc-server/pkg/security"
	"github.com/stretchr/testify/require"

	org_events "github.com/grokloc/grokloc-server/pkg/app/admin/org/events"
)

func (s *AdminSuite) TestCreateOrg() {
	ownerPassword, err := security.DerivePassword(uuid.NewString(), s.srv.ST.Argon2Cfg)
	require.Nil(s.T(), err)

	event, err := org_events.NewCreate(
		s.ctx,
		uuid.NewString(), // name
		uuid.NewString(), // owner display name
		uuid.NewString(), // owner email
		ownerPassword,
	)
	require.Nil(s.T(), err)

	bs, err := json.Marshal(event)
	require.Nil(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, s.ts.URL+OrgRoute, bytes.NewBuffer(bs))
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(s.token.Bearer))
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	location := resp.Header.Get("location")
	require.NotEmpty(s.T(), location)

	// duplicate
	resp, err = s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, resp.StatusCode)
}
