package safe

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StringSuite struct {
	suite.Suite
}

func (s *StringSuite) TestStringIs() {
	require.Equal(s.T(), ErrStringLength, StringIs(""))
	require.Equal(s.T(), ErrCharsDetected, StringIs("hello'"))
	require.Equal(s.T(), ErrCharsDetected, StringIs("hello`"))
	require.NoError(s.T(), StringIs("hello"))

	for _, v := range []string{
		"insert ",
		"update ",
		"upsert ",
		"drop ",
		"create "} {
		require.Equal(s.T(), ErrSQLDetected, StringIs(fmt.Sprintf("%s ", v)))
		require.Equal(s.T(), ErrSQLDetected, StringIs(fmt.Sprintf(" %s ", v)))
		require.Equal(s.T(), ErrSQLDetected, StringIs(fmt.Sprintf("%s ", strings.ToUpper(v))))
	}
	require.Equal(s.T(), ErrCharsDetected, StringIs(" < "))
	require.Equal(s.T(), ErrCharsDetected, StringIs(" > "))
	require.Equal(s.T(), ErrHTMLDetected, StringIs("&gt;"))
	require.Equal(s.T(), ErrHTMLDetected, StringIs("&lt;"))
	require.Equal(s.T(), ErrHTMLDetected, StringIs("window.onload"))
	require.Equal(s.T(), ErrWSDetected, StringIs(`
                                      multi
                                      line
                                     `))
	require.Equal(s.T(), ErrWSDetected, StringIs("\thello\t"))
}

func (s *StringSuite) TestIDIs() {
	require.Nil(s.T(), IDIs(uuid.NewString()))
	require.Error(s.T(), IDIs(``))
	require.Error(s.T(), IDIs(uuid.NewString()+` `+uuid.NewString()))
}

func TestStringSuite(t *testing.T) {
	suite.Run(t, new(StringSuite))
}
