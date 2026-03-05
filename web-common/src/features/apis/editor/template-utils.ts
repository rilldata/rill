export interface Template {
  label: string;
  description: string;
  content: string;
}

const header = `# API YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/apis
# Test your API endpoint at http://localhost:9009/v1/instances/default/api/<filename>

`;

export const templates: Template[] = [
  {
    label: "SQL Query",
    description:
      "Query a model or source using SQL. Use {{ .args.param }} to accept dynamic arguments.",
    content: `${header}type: api
sql: |
  SELECT * FROM model_name
`,
  },
  {
    label: "SQL Query with Limits",
    description:
      "Query a model with pagination controls. Pass 'limit' and 'offset' arguments to control page size and position.",
    content: `${header}type: api
sql: |
  SELECT * FROM model_name
  LIMIT {{ .args.limit }}
  OFFSET {{ .args.offset }}
`,
  },
  {
    label: "Metrics SQL",
    description:
      "Query a metrics view using Metrics SQL. Reference measures and dimensions defined in your metrics view.",
    content: `${header}type: api
metrics_sql: |
  SELECT measure, dimension FROM metrics_view_name
`,
  },
  {
    label: "Metrics SQL with Args",
    description:
      "Query a metrics view with dynamic filtering. Pass arguments to filter results at query time.",
    content: `${header}type: api
metrics_sql: |
  SELECT measure, dimension FROM metrics_view_name
  WHERE dimension = '{{ .args.filter }}'
`,
  },
  {
    label: "Metrics SQL with Pagination",
    description:
      "Query a metrics view with pagination controls. Pass 'limit' and 'offset' arguments to page through results.",
    content: `${header}type: api
metrics_sql: |
  SELECT measure, dimension FROM metrics_view_name
  LIMIT {{ .args.limit }}
  OFFSET {{ .args.offset }}
`,
  },
  {
    label: "OpenAPI",
    description:
      "Define an API with an OpenAPI specification. Includes request and response schemas for documentation and validation.",
    content: `${header}type: api
metrics_sql: |
  SELECT measure, dimension FROM metrics_view_name
  WHERE dimension = '{{ .args.filter }}'
openapi:
  summary: Describe your API endpoint
  request_schema:
    type: object
    properties:
      filter:
        type: string
        description: Filter by dimension value
  response_schema:
    type: object
    properties:
      measure:
        type: number
      dimension:
        type: string
`,
  },
  {
    label: "Resource Status",
    description:
      "Return the reconciliation status of resources in the project. Useful for health-check and monitoring endpoints.",
    content: `${header}type: api
resource_status:
  where_error: true
`,
  },
];
