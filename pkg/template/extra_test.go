package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/helmwave/helmwave/pkg/template"
	"github.com/stretchr/testify/suite"
)

type ExtraTestSuite struct {
	suite.Suite
}

func (s *ExtraTestSuite) TestToYaml() {
	data := struct {
		Field       interface{}
		nonexported interface{}
	}{
		Field:       "field",
		nonexported: 123,
	}
	yamlData := "field: field"

	y, err := template.ToYaml(data)
	s.Require().NoError(err)
	s.Require().YAMLEq(yamlData, y)
}

type raw struct{}

func (r raw) MarshalYAML() (interface{}, error) {
	return nil, os.ErrNotExist
}

func (s *ExtraTestSuite) TestToYamlNil() {
	data := raw{}
	y, err := template.ToYaml(data)
	s.Require().Equal("", y)
	s.Require().ErrorIs(err, os.ErrNotExist)
}

func (s *ExtraTestSuite) TestFromYaml() {
	tests := []struct {
		result template.Values
		yaml   string
		fails  bool
	}{
		{
			yaml:   "abc: 123",
			result: template.Values{"abc": uint64(123)},
			fails:  false,
		},
		{
			yaml:  "123!!123",
			fails: true,
		},
	}

	for i := range tests {
		v, err := template.FromYaml(tests[i].yaml)
		if tests[i].fails {
			s.Require().Error(err)
			s.Require().Empty(v)
		} else {
			s.Require().NoError(err)
			s.Require().Equal(tests[i].result, v)
		}
	}
}

func (s *ExtraTestSuite) TestExec() {
	res, err := template.Exec("pwd", []interface{}{})
	s.Require().NoError(err)

	pwd, err := os.Getwd()
	s.Require().NoError(err)

	s.Require().Equal(pwd, strings.TrimSpace(res))
}

func (s *ExtraTestSuite) TestExecInvalidArg() {
	res, err := template.Exec("pwd", []interface{}{123})
	s.Require().Error(err)
	s.Require().Empty(res)
}

func (s *ExtraTestSuite) TestExecError() {
	res, err := template.Exec(s.T().Name(), []interface{}{})
	s.Require().Error(err)
	s.Require().Empty(res)
}

func (s *ExtraTestSuite) TestExecStdin() {
	input := "123"
	res, err := template.Exec("cat", []interface{}{}, input)
	s.Require().NoError(err)
	s.Require().Equal(input, res)
}

func (s *ExtraTestSuite) TestSetValueAtPath() {
	data := template.Values{
		"a": map[string]interface{}{
			"b": "123",
		},
		"c": 123,
	}

	tests := []struct {
		result template.Values
		value  interface{}
		path   string
		fails  bool
	}{
		{
			path:  "c",
			value: 321,
			result: template.Values{
				"a": map[string]interface{}{"b": "123"},
				"c": 321,
			},
			fails: false,
		},
		{
			path:  "a.b",
			value: "321",
			result: template.Values{
				"a": map[string]interface{}{"b": "321"},
				"c": 321,
			},
			fails: false,
		},
		{
			path:  "a.c",
			value: "321",
			result: template.Values{
				"a": map[string]interface{}{"b": "321", "c": "321"},
				"c": 321,
			},
			fails: false,
		},
		{
			path:   "c.a",
			value:  "321",
			result: nil,
			fails:  true,
		},
	}

	for i := range tests {
		res, err := template.SetValueAtPath(tests[i].path, tests[i].value, data)
		if tests[i].fails {
			s.Require().Error(err)
			s.Require().Nil(res)
		} else {
			s.Require().NoError(err)
			s.Require().Equal(tests[i].result, res)
		}
	}
}

func (s *ExtraTestSuite) TestRequiredEnv() {
	name := s.T().Name()

	res, err := template.RequiredEnv(name)
	s.Require().Error(err)
	s.Require().Empty(res)

	data := "test"
	s.T().Setenv(name, data)

	res, err = template.RequiredEnv(name)
	s.Require().NoError(err)
	s.Require().Equal(data, res)
}

func (s *ExtraTestSuite) TestRequired() {
	tests := []struct {
		data  interface{}
		fails bool
	}{
		{
			data:  nil,
			fails: true,
		},
		{
			data:  4,
			fails: false,
		},
		{
			data:  "",
			fails: true,
		},
		{
			data:  "123",
			fails: false,
		},
	}

	for _, t := range tests {
		res, err := template.Required("blabla", t.data)
		if t.fails {
			s.Require().Error(err)
			s.Require().Nil(res)
		} else {
			s.Require().NoError(err)
			s.Require().Equal(t.data, res)
		}
	}
}

func (s *ExtraTestSuite) TestReadFile() {
	tmpDir := s.T().TempDir()
	tmpFile := filepath.Join(tmpDir, "blablafile")

	res, err := template.ReadFile(tmpFile)

	s.Require().Equal("", res)
	s.Require().ErrorIs(err, os.ErrNotExist)

	data := s.T().Name()

	s.Require().NoError(os.WriteFile(tmpFile, []byte(data), 0o600))
	s.Require().FileExists(tmpFile)

	res, err = template.ReadFile(tmpFile)

	s.Require().NoError(err)
	s.Require().Equal(data, res)
}

func (s *ExtraTestSuite) TestGet() {
	data := template.Values{
		"a": map[string]interface{}{
			"b": "123",
		},
		"c": 123,
	}

	tests := []struct {
		result interface{}
		path   string
		fails  bool
	}{
		{
			path:   "c",
			result: 123,
			fails:  false,
		},
		{
			path:   "a.b",
			result: "123",
			fails:  false,
		},
		{
			path:   "a.c",
			result: nil,
			fails:  true,
		},
		{
			path:   "c.a",
			result: nil,
			fails:  true,
		},
	}

	for i := range tests {
		res, err := template.Get(tests[i].path, data)
		if tests[i].fails {
			s.Require().Error(err)
			s.Require().Nil(res)
		} else {
			s.Require().NoError(err)
			s.Require().Equal(tests[i].result, res)
		}
	}
}

func (s *ExtraTestSuite) TestHasKey() {
	data := template.Values{
		"a": map[string]interface{}{
			"b": "123",
		},
		"c": 123,
	}

	tests := []struct {
		path   string
		result bool
		fails  bool
	}{
		{
			path:   "c",
			result: true,
			fails:  false,
		},
		{
			path:   "a.b",
			result: true,
			fails:  false,
		},
		{
			path:   "a.c",
			result: false,
			fails:  false,
		},
		{
			path:   "c.a",
			result: false,
			fails:  true,
		},
	}

	for _, test := range tests {
		res, err := template.HasKey(test.path, data)
		s.Require().Equal(test.result, res)

		if test.fails {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)
		}
	}
}

func TestExtraTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ExtraTestSuite))
}
