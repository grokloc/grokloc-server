package security

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SafeSuite struct {
	suite.Suite
}

func (s *SafeSuite) TestSafeStr() {
	require.False(s.T(), SafeStr(""))
	require.False(s.T(), SafeStr("hello'"))
	require.False(s.T(), SafeStr("hello`"))
	require.True(s.T(), SafeStr("hello"))

	for _, v := range []string{
		"insert",
		"update",
		"upsert",
		"drop",
		"create",
		"select"} {
		require.False(s.T(), SafeStr(fmt.Sprintf("%s ", v)))
		require.False(s.T(), SafeStr(fmt.Sprintf(" %s ", v)))
		require.False(s.T(), SafeStr(fmt.Sprintf("%s ", strings.ToUpper(v))))
	}
	require.False(s.T(), SafeStr(" < "))
	require.False(s.T(), SafeStr(" > "))
	require.False(s.T(), SafeStr("&gt;"))
	require.False(s.T(), SafeStr("&lt;"))
	require.False(s.T(), SafeStr("window.onload"))
	require.False(s.T(), SafeStr(`
                                      multi
                                      line
                                     `))
	require.False(s.T(), SafeStr("\thello\t"))
}

func TestSafeSuite(t *testing.T) {
	suite.Run(t, new(SafeSuite))
}
