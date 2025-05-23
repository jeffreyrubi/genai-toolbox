// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package databricks

import (
	"context"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/googleapis/genai-toolbox/tests"
)

var (
	DATABRICKS_SOURCE_KIND = "databricks"
	DATABRICKS_TOOL_KIND   = "databricks-execute-sql"
	DATABRICKS_HOST        = os.Getenv("DATABRICKS_HOST")
	DATABRICKS_TOKEN       = os.Getenv("DATABRICKS_TOKEN")
	DATABRICKS_HTTP_PATH   = os.Getenv("DATABRICKS_HTTP_PATH")
	DATABRICKS_CATALOG     = os.Getenv("DATABRICKS_CATALOG")
	DATABRICKS_SCHEMA      = os.Getenv("DATABRICKS_SCHEMA")
)

// getDatabricksVars returns the Databricks configuration variables and validates they're set
func getDatabricksVars(t *testing.T) map[string]any {
	// Check required variables
	switch "" {
	case DATABRICKS_HOST:
		t.Fatal("'DATABRICKS_HOST' not set")
	case DATABRICKS_TOKEN:
		t.Fatal("'DATABRICKS_TOKEN' not set")
	case DATABRICKS_HTTP_PATH:
		t.Fatal("'DATABRICKS_HTTP_PATH' not set")
	}

	// Catalog and schema are optional, default to empty which means use the warehouse default
	return map[string]any{
		"kind":      DATABRICKS_SOURCE_KIND,
		"host":      DATABRICKS_HOST,
		"token":     DATABRICKS_TOKEN,
		"http_path": DATABRICKS_HTTP_PATH,
		"catalog":   DATABRICKS_CATALOG,
		"schema":    DATABRICKS_SCHEMA,
	}
}

func TestDatabricks(t *testing.T) {
	sourceConfig := getDatabricksVars(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var args []string

	// Provide a basic tool statement for the Databricks execute-sql tool (YAML, not JSON)
	toolStatement := `
	name: my-exec-sql-tool
	kind: databricks-execute-sql
	source: databricks
	description: Execute arbitrary SQL on Databricks.
	statement: SELECT 1 AS test
	`
	// Pass a comment as the 4th argument to avoid writing an invalid tool config
	toolsFile := tests.GetToolsConfig(sourceConfig, DATABRICKS_TOOL_KIND, toolStatement, "#")

	cmd, cleanup, err := tests.StartCmd(ctx, toolsFile, args...)
	if err != nil {
		t.Fatalf("command initialization returned an error: %s", err)
	}
	defer cleanup()

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := cmd.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`))
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %s", err)
	}

	tests.RunToolGetTest(t)

	// Databricks-specific wants and params
	select1Want := `[{"test":1}]`
	failInvocationWant := `{"jsonrpc":"2.0","id":"invoke-fail-tool","result":{"content":[{"type":"text","text":"unable to execute query: Syntax error: Unexpected identifier` // partial match for error
	invokeParamWant := `[{"test":1}]`
	mcpInvokeParamWant := `{"jsonrpc":"2.0","id":"my-param-tool","result":{"content":[{"type":"text","text":"{\"test\":1}"}]}}`

	tests.RunToolInvokeTest(t, select1Want, invokeParamWant)
	tests.RunPgExecuteSqlToolInvokeTest(t, select1Want)
	tests.RunMCPToolCallMethod(t, mcpInvokeParamWant, failInvocationWant)
}

func TestDatabricksIntegration(t *testing.T) {
	sourceConfig := getDatabricksVars(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var args []string

	// Prepare tool statements (YAML, not JSON)
	toolStatement1 := `
  name: my-param-tool
  kind: databricks-execute-sql
  source: databricks
  description: Parametrized query tool
  statement: SELECT 1 AS test
`
	toolStatement2 := `
  name: my-auth-tool
  kind: databricks-execute-sql
  source: databricks
  description: Auth required query tool
  statement: SELECT 1 AS test
`

	toolsFile := tests.GetToolsConfig(sourceConfig, DATABRICKS_TOOL_KIND, toolStatement1, toolStatement2)

	cmd, cleanup, err := tests.StartCmd(ctx, toolsFile, args...)
	if err != nil {
		t.Fatalf("command initialization returned an error: %s", err)
	}
	defer cleanup()

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := cmd.WaitForString(waitCtx, regexp.MustCompile(`Server ready to serve`))
	if err != nil {
		t.Logf("toolbox command logs: \n%s", out)
		t.Fatalf("toolbox didn't start successfully: %s", err)
	}

	tests.RunToolGetTest(t)

	// Use expected values for Databricks
	select1Want := `[{"test":1}]`
	failInvocationWant := `{"jsonrpc":"2.0","id":"invoke-fail-tool","result":{"content":[{"type":"text","text":"unable to execute query: Syntax error: Unexpected identifier` // partial match for error
	invokeParamWant := `[{"test":1}]`
	mcpInvokeParamWant := `{"jsonrpc":"2.0","id":"my-param-tool","result":{"content":[{"type":"text","text":"{\"test\":1}"}]}}`

	tests.RunToolInvokeTest(t, select1Want, invokeParamWant)
	tests.RunPgExecuteSqlToolInvokeTest(t, select1Want)
	tests.RunMCPToolCallMethod(t, mcpInvokeParamWant, failInvocationWant)
}
