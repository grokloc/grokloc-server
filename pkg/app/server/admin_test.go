package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/security"
	"github.com/stretchr/testify/suite"
)

// AdminSuite is responsible for admin endpoint testing
type AdminSuite struct {
	suite.Suite
	srv   *Instance
	ctx   context.Context
	ts    *httptest.Server
	c     *http.Client
	token *Token
}

func (s *AdminSuite) SetupTest() {
	var err error
	s.srv, err = New(env.Unit)
	if err != nil {
		log.Fatal(err.Error())
	}

	s.ctx = context.Background()
	s.ts = httptest.NewServer(s.srv.Router())
	s.c = &http.Client{}

	// for making authenticated requests, get a token
	// (these steps are already run through real tests in token_test)
	req, err := http.NewRequest(http.MethodPut, s.ts.URL+TokenRoute, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	req.Header.Add(IDHeader, s.srv.ST.RootUser)
	req.Header.Add(TokenRequestHeader, security.EncodedSHA256(s.srv.ST.RootUser+s.srv.ST.RootUserAPISecret))
	resp, err := s.c.Do(req)

	if err != nil {
		log.Fatal(err.Error())
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	var tok Token
	err = json.Unmarshal(respBody, &tok)
	if err != nil {
		log.Fatal(err.Error())
	}
	s.token = &tok
}
