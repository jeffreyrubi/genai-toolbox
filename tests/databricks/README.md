# Databricks Integration

This directory contains integration tests for the Databricks data source and SQL tool.

## Running Integration Tests

To run the integration tests, you need valid Databricks credentials:

1. **DATABRICKS_HOST**: Your Databricks workspace URL (e.g., `https://your-workspace.cloud.databricks.com`)
2. **DATABRICKS_TOKEN**: Personal access token for authentication
3. **DATABRICKS_HTTP_PATH**: The SQL warehouse HTTP path

Note that the HTTP_PATH should be just the warehouse ID, not the full path. The full path typically looks like:
```
/sql/1.0/warehouses/abcdef123456789
```

But you should use just the ID part: `abcdef123456789`

### Running the test

```bash
# Run the test
RUN_INTEGRATION_TESTS=1 \
DATABRICKS_HOST="https://your-workspace.cloud.databricks.com" \
DATABRICKS_TOKEN="your-token" \
DATABRICKS_HTTP_PATH="your-warehouse-id" \
go test -v ./tests/databricks
```

## Implementation Details

The Databricks integration consists of:

- Source implementation in `/internal/sources/databricks`
- SQL tool implementation in `/internal/tools/databrickssql`

The integration test verifies that the source can connect to Databricks and execute a simple query.

## Common Issues

If you encounter the error `/sql/1.0/warehouses/<id> is not a valid endpoint id`, make sure you're using just the warehouse ID, not the full path.
