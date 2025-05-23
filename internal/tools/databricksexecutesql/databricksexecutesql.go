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

package databricksexecutesql

import (
	"context"
	"fmt"
	"strings"

	databricks "github.com/databricks/databricks-sdk-go"
	"github.com/googleapis/genai-toolbox/internal/sources"
	databrickssrc "github.com/googleapis/genai-toolbox/internal/sources/databricks"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

const ToolKind string = "databricks-execute-sql"

type compatibleSource interface {
	DatabricksClient() *databricks.WorkspaceClient
	GetHttpPath() string
	GetCatalog() string
	GetSchema() string
}

// validate compatible sources are still compatible
var _ compatibleSource = &databrickssrc.Source{}

var compatibleSources = [...]string{databrickssrc.SourceKind}

type Config struct {
	Name         string   `yaml:"name" validate:"required"`
	Kind         string   `yaml:"kind" validate:"required"`
	Source       string   `yaml:"source" validate:"required"`
	Description  string   `yaml:"description" validate:"required"`
	AuthRequired []string `yaml:"authRequired"`
}

// validate interface
var _ tools.ToolConfig = Config{}

func (cfg Config) ToolConfigKind() string {
	return ToolKind
}

func (cfg Config) Initialize(srcs map[string]sources.Source) (tools.Tool, error) {
	// verify source exists
	rawS, ok := srcs[cfg.Source]
	if !ok {
		return nil, fmt.Errorf("no source named %q configured", cfg.Source)
	}

	// verify the source is compatible
	s, ok := rawS.(compatibleSource)
	if !ok {
		return nil, fmt.Errorf("invalid source for %q tool: source kind must be one of %q", ToolKind, compatibleSources)
	}

	// Get the warehouse ID and clean it if necessary
	httpPath := s.GetHttpPath()
	if httpPath == "" {
		return nil, fmt.Errorf("warehouse ID is empty for source %q", cfg.Source)
	}

	// Create a SQL parameter that will be filled in dynamically
	sqlParameter := tools.NewStringParameter("sql", "The SQL to execute.")
	parameters := tools.Parameters{sqlParameter}

	mcpManifest := tools.McpManifest{
		Name:        cfg.Name,
		Description: cfg.Description,
		InputSchema: parameters.McpManifest(),
	}

	// finish tool setup
	t := Tool{
		Name:         cfg.Name,
		Kind:         ToolKind,
		Parameters:   parameters,
		AuthRequired: cfg.AuthRequired,
		Client:       s.DatabricksClient(),
		HttpPath:     httpPath,
		Catalog:      s.GetCatalog(),
		Schema:       s.GetSchema(),
		manifest:     tools.Manifest{Description: cfg.Description, Parameters: parameters.Manifest(), AuthRequired: cfg.AuthRequired},
		mcpManifest:  mcpManifest,
	}
	return t, nil
}

// validate interface
var _ tools.Tool = Tool{}

type Tool struct {
	Name         string           `yaml:"name"`
	Kind         string           `yaml:"kind"`
	AuthRequired []string         `yaml:"authRequired"`
	Parameters   tools.Parameters `yaml:"parameters"`

	Client       *databricks.WorkspaceClient
	HttpPath     string
	Catalog      string
	Schema       string
	manifest     tools.Manifest
	mcpManifest  tools.McpManifest
}

func (t Tool) Invoke(ctx context.Context, params tools.ParamValues) ([]any, error) {
	sliceParams := params.AsSlice()
	sql, ok := sliceParams[0].(string)
	if !ok {
		return nil, fmt.Errorf("unable to cast parameter to string: %v", sliceParams[0])
	}

	// Added logging to help with debugging
	fmt.Printf("Databricks Execute SQL - Using warehouse ID: %s\n", t.HttpPath)
	fmt.Printf("Databricks Execute SQL - Executing SQL: %s\n", sql)
	
	// Make sure we're using a clean warehouse ID (just the ID, not the full path)
	warehouseID := t.HttpPath
	if len(warehouseID) > 0 && warehouseID[0] == '/' {
		parts := []string{}
		for _, part := range strings.Split(warehouseID, "/") {
			if part != "" {
				parts = append(parts, part)
			}
		}
		if len(parts) > 0 {
			warehouseID = parts[len(parts)-1]
		}
	}
	
	// Create a source to execute the query
	source := &databrickssrc.Source{
		Client:   t.Client,
		HttpPath: warehouseID, // Use cleaned warehouse ID
		Catalog:  t.Catalog,
		Schema:   t.Schema,
	}
	
	// Execute the SQL statement with robust error handling
	rows, err := source.ExecuteSQL(ctx, sql, nil)
	if err != nil {
		// Add more context to the error
		return nil, fmt.Errorf("failed to execute SQL with warehouse ID '%s': %w", warehouseID, err)
	}
	
	// Convert result rows to any type for return
	result := make([]any, len(rows))
	for i, row := range rows {
		result[i] = row
	}
	
	return result, nil
}

func (t Tool) ParseParams(data map[string]any, claims map[string]map[string]any) (tools.ParamValues, error) {
	return tools.ParseParams(t.Parameters, data, claims)
}

func (t Tool) Manifest() tools.Manifest {
	return t.manifest
}

func (t Tool) McpManifest() tools.McpManifest {
	return t.mcpManifest
}

func (t Tool) Authorized(verifiedAuthServices []string) bool {
	return tools.IsAuthorized(t.AuthRequired, verifiedAuthServices)
}
