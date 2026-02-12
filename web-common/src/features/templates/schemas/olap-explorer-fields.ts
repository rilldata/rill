/**
 * Shared explorer-step schema fields for OLAP connectors.
 * These define the SQL query and model name fields used in the explorer step.
 */
export function olapExplorerFields(engineName: string) {
  return {
    sql: {
      type: "string" as const,
      title: "SQL",
      description: `SQL query to run against ${engineName}`,
      "x-placeholder": "SELECT * FROM my_table",
      "x-step": "explorer" as const,
    },
    name: {
      type: "string" as const,
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_model",
      "x-step": "explorer" as const,
    },
  };
}
