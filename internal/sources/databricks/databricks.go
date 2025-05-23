package databricks

import (
	"context"
	"fmt"
	"strings"

	databricks "github.com/databricks/databricks-sdk-go"
	dbsql "github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"go.opentelemetry.io/otel/trace"
)

const SourceKind string = "databricks"

// validate interface
var _ sources.SourceConfig = Config{}

type Config struct {
    Name     string `yaml:"name" validate:"required"`
    Kind     string `yaml:"kind" validate:"required"`
    Host     string `yaml:"host" validate:"required"`
    Token    string `yaml:"token" validate:"required"`
    HttpPath string `yaml:"http_path" validate:"required"`
    Catalog  string `yaml:"catalog"`
    Schema   string `yaml:"schema"`
}

func (r Config) SourceConfigKind() string {
    return SourceKind
}

func (r Config) Initialize(ctx context.Context, tracer trace.Tracer) (sources.Source, error) {
    // Extract warehouse ID from full HTTP path if needed
    httpPath := r.HttpPath
    if len(httpPath) > 0 && httpPath[0] == '/' {
        // Extract just the warehouse ID from the full path
        parts := []string{}
        for _, part := range strings.Split(httpPath, "/") {
            if part != "" {
                parts = append(parts, part)
            }
        }
        if len(parts) > 0 {
            httpPath = parts[len(parts)-1]
        }
    }

    client, err := initDatabricksClient(ctx, tracer, r.Name, r.Host, r.Token, httpPath, r.Catalog, r.Schema)
    if err != nil {
        return nil, fmt.Errorf("unable to create databricks client: %w", err)
    }

    s := &Source{
        Name:     r.Name,
        Kind:     SourceKind,
        Client:   client,
        HttpPath: httpPath,
        Catalog:  r.Catalog,
        Schema:   r.Schema,
    }
    return s, nil
}

var _ sources.Source = &Source{}

type Source struct {
    Name     string `yaml:"name"`
    Kind     string `yaml:"kind"`
    Client   *databricks.WorkspaceClient
    HttpPath string
    Catalog  string
    Schema   string
}

func (s *Source) SourceKind() string {
    return SourceKind
}

// DatabricksClient returns the Databricks workspace client
func (s *Source) DatabricksClient() *databricks.WorkspaceClient {
    return s.Client
}

// GetHttpPath returns the HTTP path for SQL warehouse connection
func (s *Source) GetHttpPath() string {
    return s.HttpPath
}

// GetCatalog returns the catalog name
func (s *Source) GetCatalog() string {
    return s.Catalog
}

// GetSchema returns the schema name
func (s *Source) GetSchema() string {
    return s.Schema
}

// ExecuteSQL executes a SQL query on Databricks and returns the results
func (s *Source) ExecuteSQL(ctx context.Context, statement string, params []interface{}) ([]map[string]interface{}, error) {
    // Create the SQL warehouse context with schema/catalog if provided
    // Make sure we're using a valid warehouse ID (without any path components)
    warehouseID := s.HttpPath
    if warehouseID == "" {
        return nil, fmt.Errorf("warehouse ID cannot be empty")
    }
    
    // Prepare the request
    req := dbsql.ExecuteStatementRequest{
        WarehouseId: warehouseID,
        Statement:   statement,
    }
    
    // Add catalog if specified
    if s.Catalog != "" {
        req.Catalog = s.Catalog
    }
    
    // Add schema if specified
    if s.Schema != "" {
        req.Schema = s.Schema
    }
    
    // Add parameters if any
    if len(params) > 0 {
        // Simple parameter substitution for ?-style placeholders
        // Replace ? with actual values from params
        for _, param := range params {
            switch v := param.(type) {
            case string:
                // Escape single quotes and wrap in quotes
                escapedVal := strings.Replace(v, "'", "''", -1)
                statement = strings.Replace(statement, "?", fmt.Sprintf("'%s'", escapedVal), 1)
            case int, int64, float64, bool:
                // Numbers and booleans don't need quotes
                statement = strings.Replace(statement, "?", fmt.Sprintf("%v", v), 1)
            case nil:
                statement = strings.Replace(statement, "?", "NULL", 1)
            default:
                // Try to stringify anything else
                statement = strings.Replace(statement, "?", fmt.Sprintf("%v", v), 1)
            }
        }
        req.Statement = statement
    }
    
    // Execute the query and wait for the results
    resp, err := s.Client.StatementExecution.ExecuteAndWait(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to execute SQL: %w", err)
    }

    // Convert to standard map format
    rows := make([]map[string]interface{}, 0)
    if resp.Result != nil && resp.Result.DataArray != nil && resp.Manifest != nil && resp.Manifest.Schema != nil {
        for _, row := range resp.Result.DataArray {
            rowMap := make(map[string]interface{})
            for i, col := range resp.Manifest.Schema.Columns {
                if i < len(row) {
                    rowMap[col.Name] = row[i]
                }
            }
            rows = append(rows, rowMap)
        }
    }
    
    return rows, nil
}

func initDatabricksClient(ctx context.Context, tracer trace.Tracer, name, host, token, httpPath, catalog, schema string) (*databricks.WorkspaceClient, error) {
    //nolint:all // Reassigned ctx
    ctx, span := sources.InitConnectionSpan(ctx, tracer, SourceKind, name)
    defer span.End()

    // Initialize the Databricks client using the SDK
    cfg := databricks.Config{
        Host:  host,
        Token: token,
    }
    client, err := databricks.NewWorkspaceClient(&cfg)
    if err != nil {
        return nil, fmt.Errorf("databricks.NewWorkspaceClient: %w", err)
    }
    
    return client, nil
}

// Sources are registered in the server/config.go file
// No init() function needed here
