package databricks_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/sources/databricks"
	"github.com/googleapis/genai-toolbox/internal/testutils"
)

func TestParseFromYamlDatabricks(t *testing.T) {
	tcs := []struct {
		desc string
		in   string
		want server.SourceConfigs
	}{
		{
			desc: "basic example",
			in: `
			sources:
			  my-databricks:
			    kind: databricks
			    host: my-host
			    token: my-token
			    http_path: my-http-path
			    catalog: my-catalog
			    schema: my-schema
			`,
			want: server.SourceConfigs{
				"my-databricks": databricks.Config{
					Name:     "my-databricks",
					Kind:     databricks.SourceKind,
					Host:     "my-host",
					Token:    "my-token",
					HttpPath: "my-http-path",
					Catalog:  "my-catalog",
					Schema:   "my-schema",
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			got := struct {
				Sources server.SourceConfigs `yaml:"sources"`
			}{}
			// Parse contents
			err := yaml.Unmarshal(testutils.FormatYaml(tc.in), &got)
			if err != nil {
				t.Fatalf("unable to unmarshal: %s", err)
			}
			if !cmp.Equal(tc.want, got.Sources) {
				t.Fatalf("incorrect parse: want %v, got %v", tc.want, got.Sources)
			}
		})
	}
}

func TestFailParseFromYamlDatabricks(t *testing.T) {
	tcs := []struct {
		desc string
		in   string
		err  string
	}{
		{
			desc: "extra field",
			in: `
			sources:
			  my-databricks:
			    kind: databricks
			    host: my-host
			    token: my-token
			    http_path: my-http-path
			    foo: bar
			`,
			err: "unable to parse as \"databricks\": [2:1] unknown field \"foo\"\n   1 | http_path: my-http-path\n>  2 | foo: bar\n       ^\n   3 | host: my-host\n   4 | kind: databricks\n   5 | token: my-token\n   6 | ",
		},
		{
			desc: "missing required field",
			in: `
			sources:
			  my-databricks:
			    kind: databricks
			    host: my-host
			    token: my-token
			`,
			err: "unable to parse as \"databricks\": Key: 'Config.HttpPath' Error:Field validation for 'HttpPath' failed on the 'required' tag",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			got := struct {
				Sources server.SourceConfigs `yaml:"sources"`
			}{}
			// Parse contents
			err := yaml.Unmarshal(testutils.FormatYaml(tc.in), &got)
			if err == nil {
				t.Fatalf("expect parsing to fail")
			}
			errStr := err.Error()
			if errStr != tc.err {
				t.Fatalf("unexpected error: got %q, want %q", errStr, tc.err)
			}
		})
	}
}
