# Databricks Tools Guide

This guide explains how to use the Databricks SQL tools added to the genai-toolbox.

## Available Tools

### 1. Predefined SQL Queries
- **databricks-list-tables**: Lists all tables in the current Databricks schema
- **databricks-query-sample**: Executes a query against a specified table with a limit

### 2. Dynamic SQL Execution
- **databricks-execute-sql**: Allows executing arbitrary SQL queries at runtime

## Running the Server

### Debug Mode (VS Code)

The launch.json file contains configurations for running the server with Databricks tools in debug mode.

To start debugging:
1. Open VS Code
2. Go to the "Run and Debug" sidebar (Ctrl+Shift+D or Cmd+Shift+D)
3. Select one of the launch configurations from the dropdown
4. Click the green play button or press F5

1. **Run Server with Databricks Tools**:
   - Runs the server with both Postgres and Databricks tools available
   - Access toolsets via `http://localhost:8888/api/toolset/databricks`

2. **Run Server with Databricks Toolset Only**:
   - Sets the default toolset to only include Databricks tools
   - Access tools via `http://localhost:8889/api/toolset`

3. **Debug Databricks Execute SQL**:
   - Specifically for debugging the dynamic SQL execution
   - Includes HTTP tracing for detailed debugging
   - Uses port 5001 to avoid conflicts with other instances

### Command Line

Run the server with Databricks support:

```bash
export DATABRICKS_HOST="https://adb-6856225216501813.13.azuredatabricks.net:8082"
export DATABRICKS_TOKEN="dapicfd042a91921a5b6d42435589de0abb0"
export DATABRICKS_HTTP_PATH="/sql/1.0/warehouses/84c487324df116f0"

go run . --tools-file tools.yaml --port 8888
```

## Testing the Tools

### List Tables
```
curl -X POST http://localhost:8888/api/tool/databricks-list-tables/invoke \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Query with Parameters
```
curl -X POST http://localhost:8888/api/tool/databricks-query-sample/invoke \
  -H "Content-Type: application/json" \
  -d '{"table_name": "your_table", "limit": 10}'
```

### Execute Dynamic SQL
```
curl -X POST http://localhost:8888/api/tool/databricks-execute-sql/invoke \
  -H "Content-Type: application/json" \
  -d '{"sql": "SELECT * FROM your_table LIMIT 5"}'
```

### Using Databricks Toolset (Port 8889)
```
curl -X POST http://localhost:8889/api/tool/databricks-execute-sql/invoke \
  -H "Content-Type: application/json" \
  -d '{"sql": "SHOW TABLES"}'
```

### Get Available Toolsets
```
curl http://localhost:8888/api/toolset
```

### Get Specific Toolset
```
curl http://localhost:8888/api/toolset/databricks
```

## Integration Tests

For integration testing, use the VS Code launch configuration:

- **Run Databricks Execute SQL Integration Test**: Tests the dynamic SQL execution feature
- **Run Databricks Integration Test**: Tests the predefined SQL tool
- **Run All Integration Tests**: Runs all tests including Databricks tests

## Troubleshooting

If you encounter issues:

1. **Warehouse ID Format**:
   - Ensure your warehouse ID is correctly extracted from the HTTP path

2. **Authentication Issues**:
   - Verify your Databricks token has access to the specified warehouse

3. **Warehouse Status**:
   - Ensure the Databricks SQL warehouse is running and not in a stopped state

4. **HTTP Path**:
   - The full path should look like `/sql/1.0/warehouses/84c487324df116f0`
   - The tool will extract just the warehouse ID automatically
