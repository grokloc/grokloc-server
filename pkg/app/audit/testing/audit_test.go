// Package testing provides tests for the audit package
// (broken out to break import cycles)
package testing

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/audit"
	"github.com/grokloc/grokloc-server/pkg/app/state"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/grokloc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuditSuite struct {
	suite.Suite
	st *app.State
}

func (s *AuditSuite) SetupTest() {
	var err error
	s.st, err = state.New(env.Unit)
	if err != nil {
		zap.L().Fatal("setup",
			zap.Error(err),
		)
	}
}

func (s *AuditSuite) TestInsert() {
	err := audit.Insert(
		grokloc.context.Background(),
		audit.USER_INSERT,
		uuid.NewString(),
		uuid.NewString(),
		s.st.Master,
	)
	require.Nil(s.T(), err)
}

func TestAuditSuite(t *testing.T) {
	suite.Run(t, new(AuditSuite))
}
