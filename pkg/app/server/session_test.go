package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
	"github.com/matthewhartstonge/argon2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const authLevel = "authlevel"

// SessionSuite is responsible fo testing interior methods used in auth
type SessionSuite struct {
	suite.Suite
	srv *Instance
	ctx context.Context
	ts  *httptest.Server
	c   *http.Client
}

func (s *SessionSuite) SetupTest() {
	var err error
	s.srv, err = New(env.Unit)
	if err != nil {
		log.Fatal(err.Error())
	}

	// returns the auth leve in a header, and "OK" as a body
	okHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		lvl, ok := ctx.Value(authLevelCtxKey).(int)
		if !ok {
			http.Error(w, "authLevel", http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		// tests can read out the auth level
		w.Header().Set(authLevel, fmt.Sprintf("%d", lvl))
		_, err := w.Write([]byte("OK"))
		if err != nil {
			panic(err.Error())
		}
	}

	rtr := chi.NewRouter()
	rtr.Route("/", func(rtr chi.Router) {
		rtr.Use(s.srv.WithSession)

		// OK handler that only wants a valid user/org
		rtr.Get("/", okHandler)

		// "/token" runs the token generation handler
		rtr.Put("/token", s.srv.NewToken)

		rtr.Route("/verify", func(rtr chi.Router) {
			rtr.Use(s.srv.WithToken)

			// OK handler that also wants a valid token header
			rtr.Get("/", okHandler)
		})
	})
	s.ctx = context.Background()
	s.ts = httptest.NewServer(rtr)
	s.c = &http.Client{}
}

func (s *SessionSuite) TestFoundAndActiveRoot() {
	// root
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	authLevelVal := resp.Header.Get(authLevel)
	require.NotEqual(s.T(), "", authLevelVal)
	require.Equal(s.T(), fmt.Sprintf("%d", AuthRoot), authLevelVal)
}

func (s *SessionSuite) TestFoundAndActiveOwner() {
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

	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, owner.ID)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	authLevelVal := resp.Header.Get(authLevel)
	require.NotEqual(s.T(), "", authLevelVal)
	require.Equal(s.T(), fmt.Sprintf("%d", AuthOrg), authLevelVal)
}

func (s *SessionSuite) TestFoundAndActiveUser() {
	// org user
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

	password, err := security.DerivePassword(uuid.NewString(), argon2.DefaultConfig())
	require.Nil(s.T(), err)
	u, err := user.Encrypted(
		s.ctx,
		uuid.NewString(),
		uuid.NewString(),
		o.ID,
		password,
		s.srv.ST.DBKey,
	)
	require.Nil(s.T(), err)

	u.Meta.Status = models.StatusActive
	err = u.Insert(s.ctx, s.srv.ST.Master)
	require.Nil(s.T(), err)

	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	authLevelVal := resp.Header.Get(authLevel)
	require.NotEqual(s.T(), "", authLevelVal)
	require.Equal(s.T(), fmt.Sprintf("%d", AuthUser), authLevelVal)
}

func (s *SessionSuite) TestUserNotFound() {
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, uuid.NewString())
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *SessionSuite) TestUserInactive() {
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

	u, err := user.Read(
		s.ctx,
		o.Owner,
		s.srv.ST.DBKey,
		s.srv.ST.RandomReplica(),
	)
	require.Nil(s.T(), err)

	err = u.UpdateStatus(s.ctx, models.StatusInactive, s.srv.ST.Master)
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func (s *SessionSuite) TestOrgInactive() {
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

	u, err := user.Read(
		s.ctx,
		o.Owner,
		s.srv.ST.DBKey,
		s.srv.ST.RandomReplica(),
	)
	require.Nil(s.T(), err)

	err = o.UpdateStatus(s.ctx, models.StatusInactive, s.srv.ST.Master)
	require.Nil(s.T(), err)
	req, err := http.NewRequest(http.MethodGet, s.ts.URL+"/", nil)
	require.Nil(s.T(), err)
	req.Header.Add(IDHeader, u.ID)
	resp, err := s.c.Do(req)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

func TestSessionSuite(t *testing.T) {
	suite.Run(t, new(SessionSuite))
}
