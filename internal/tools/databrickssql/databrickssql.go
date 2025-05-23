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

package databrickssql

import (
	"context"
	"fmt"

	databricks "github.com/databricks/databricks-sdk-go"
	"github.com/googleapis/genai-toolbox/internal/sources"
	databrickssrc "github.com/googleapis/genai-toolbox/internal/sources/databricks"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

const ToolKind string = "databricks-sql"

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
	Name         string           `yaml:"name" validate:"required"`
	Kind         string           `yaml:"kind" validate:"required"`
	Source       string           `yaml:"source" validate:"required"`
	Description  string           `yaml:"description" validate:"required"`
	Statement    string           `yaml:"statement" validate:"required"`
	AuthRequired []string         `yaml:"authRequired"`
	Parameters   tools.Parameters `yaml:"parameters"`
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

	mcpManifest := tools.McpManifest{
		Name:        cfg.Name,
		Description: cfg.Description,
		InputSchema: cfg.Parameters.McpManifest(),
	}

	// finish tool setup
	t := Tool{
		Name:         cfg.Name,
		Kind:         ToolKind,
		Parameters:   cfg.Parameters,
		Statement:    cfg.Statement,
		AuthRequired: cfg.AuthRequired,
		Client:       s.DatabricksClient(),
		HttpPath:     s.GetHttpPath(),
		Catalog:      s.GetCatalog(),
		Schema:       s.GetSchema(),
		manifest:     tools.Manifest{Description: cfg.Description, Parameters: cfg.Parameters.Manifest(), AuthRequired: cfg.AuthRequired},
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
	Statement    string
	manifest     tools.Manifest
	mcpManifest  tools.McpManifest
}

func (t Tool) Invoke(ctx context.Context, params tools.ParamValues) ([]any, error) {
	// Get the source to execute the query
	source := &databrickssrc.Source{
		Client:   t.Client,
		HttpPath: t.HttpPath,
		Catalog:  t.Catalog,
		Schema:   t.Schema,
	}
	
	// Convert parameters to the right format for the SQL query
	// For Databricks SQL, we need to use parameter substitution
	sliceParams := params.AsSlice()
	
	// Execute the SQL with parameters
	rows, err := source.ExecuteSQL(ctx, t.Statement, sliceParams)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL: %w", err)
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
