---
title: "Databricks"
type: docs
weight: 11
description: > 
  Configure a Databricks source to query SQL warehouse data.
---

The Databricks source allows you to connect to Databricks workspaces and execute SQL queries against Databricks SQL warehouses.

## Configuration

| Field | Required | Description |
|-------|----------|-------------|
| kind | Yes | Must be `databricks` |
| host | Yes | The Databricks workspace URL, e.g., `https://your-workspace.cloud.databricks.com` |
| token | Yes | A [personal access token](https://docs.databricks.com/dev-tools/api/latest/authentication.html) for the Databricks workspace |
| http_path | Yes | The HTTP path to the SQL warehouse. Can be either the full path (e.g., `/sql/1.0/warehouses/abcdef123456789`) or just the warehouse ID (e.g., `abcdef123456789`). If the full path is provided, the tool will extract just the warehouse ID automatically. |
| catalog | No | Default catalog to use for SQL queries (defaults to `hive_metastore`) |
| schema | No | Default schema to use for SQL queries (defaults to `default`) |

## Example Configuration

```yaml
sources:
  my-databricks-source:
    kind: databricks
    host: ${DATABRICKS_HOST}
    token: ${DATABRICKS_TOKEN}
    http_path: ${DATABRICKS_HTTP_PATH}
    catalog: hive_metastore
    schema: default
```

## Available Tools

### databricks-sql

This tool allows you to execute predefined SQL queries against your Databricks SQL warehouse.

Example tool configuration:

```yaml
tools:
  list-tables:
    kind: databricks-sql
    source: my-databricks-source
    description: "List tables in the current database"
    statement: "SHOW TABLES"
  
  query-with-parameters:
    kind: databricks-sql
    source: my-databricks-source
    description: "Query data with parameters"
    statement: "SELECT * FROM your_table WHERE id > :min_id LIMIT :limit"
    parameters:
      - name: min_id
        type: integer
        description: "Minimum ID to filter by"
      - name: limit
        type: integer
```

### databricks-execute-sql

This tool allows you to execute dynamic SQL queries against your Databricks SQL warehouse. Unlike `databricks-sql` which uses a predefined SQL statement, this tool accepts the SQL query as a parameter at runtime.

Example tool configuration:

```yaml
tools:
  execute-dynamic-sql:
    kind: databricks-execute-sql
    source: my-databricks-source
    description: "Execute dynamic SQL queries against Databricks"
```

When this tool is called, it expects a single parameter named `sql` that contains the SQL statement to execute:

```json
{
  "sql": "SELECT * FROM your_table WHERE created_date > '2023-01-01' LIMIT 10"
}
        description: "Maximum number of rows to return"
```

## Finding Your HTTP Path

To find the HTTP path for your SQL warehouse:

1. Go to your Databricks workspace
2. Navigate to SQL Warehouses
3. Select your warehouse
4. Go to the "Connection details" tab
5. Copy the "HTTP Path" value

### Troubleshooting HTTP Path Issues

If you encounter connection errors, verify that:

1. The warehouse ID is correctly extracted from the full HTTP path.
2. The warehouse is started and not in a stopped or terminated state.
3. Your access token has permissions to access the specified warehouse.
4. The warehouse ID doesn't include any extra path components or leading/trailing slashes.

The tools will automatically extract just the warehouse ID from the full HTTP path. For example, if your HTTP path is `/sql/1.0/warehouses/abcdef123456789`, the tool will use `abcdef123456789` as the warehouse ID.

## Authentication

The Databricks source uses personal access tokens (PAT) for authentication. To generate a token:

1. Go to your Databricks workspace
2. Click on your user profile in the top right corner
3. Select "User Settings"
4. Go to the "Access Tokens" tab
5. Click "Generate New Token"
6. Provide a name and expiration for your token
7. Copy the generated token value

For security, it's recommended to use environment variable substitution (`${DATABRICKS_TOKEN}`) in your configuration file rather than hardcoding the token.

curl -X GET \
  -H "Authorization: Bearer <your-pat>" \
  https://<databricks-instance>/api/2.0/unity-catalog/catalogs