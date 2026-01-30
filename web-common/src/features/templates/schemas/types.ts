export type JSONSchemaVisibleIfValue =
  | string
  | number
  | boolean
  | Array<string | number | boolean>;

export type JSONSchemaField = {
  type?: "string" | "number" | "boolean" | "object";
  title?: string;
  description?: string;
  enum?: Array<string | number | boolean>;
  const?: string | number | boolean;
  default?: string | number | boolean;
  pattern?: string;
  format?: string;
  errorMessage?: {
    pattern?: string;
    format?: string;
  };
  properties?: Record<string, JSONSchemaField>;
  required?: string[];
  "x-display"?: "radio" | "select" | "textarea" | "file" | "tabs";
  "x-monospace"?: boolean;
  "x-step"?: "connector" | "source" | "explorer";
  "x-secret"?: boolean;
  "x-visible-if"?: Record<string, JSONSchemaVisibleIfValue>;
  "x-enum-labels"?: string[];
  "x-enum-descriptions"?: string[];
  "x-placeholder"?: string;
  "x-hint"?: string;
  "x-accept"?: string;
  "x-informational"?: boolean;
  "x-docs-url"?: string;
  "x-internal"?: boolean;
  /**
   * Explicit grouping for radio/select options: maps an option value to the
   * child field keys that should render beneath that option.
   */
  "x-grouped-fields"?: Record<string, string[]>;
  /**
   * Group fields under tab options for enum-driven tab layouts.
   */
  "x-tab-group"?: Record<string, string[]>;
  // Allow custom keywords such as errorMessage or future x-extensions.
  [key: string]: unknown;
};

export type JSONSchemaCondition = {
  properties?: Record<
    string,
    { const?: string | number | boolean; minLength?: number }
  >;
  required?: string[];
  not?: JSONSchemaCondition;
};

export type JSONSchemaConstraint = {
  required?: string[];
  properties?: Record<string, JSONSchemaField>;
  // Allow custom keywords or overrides in constraints
  [key: string]: unknown;
};

export type JSONSchemaConditional = {
  if?: JSONSchemaCondition;
  then?: JSONSchemaConstraint;
  else?: JSONSchemaConstraint;
};

export type ConnectorCategory =
  | "sqlStore"
  | "olap"
  | "objectStore"
  | "fileStore"
  | "warehouse";

export type ButtonLabels = {
  idle: string;
  loading: string;
};

export type JSONSchemaObject = {
  $schema?: string;
  type: "object";
  title?: string;
  description?: string;
  properties?: Record<string, JSONSchemaField>;
  required?: string[];
  allOf?: JSONSchemaConditional[];
  oneOf?: JSONSchemaConstraint[];
  /**
   * Connector category for UI enumeration.
   * "source" = data sources (databases, cloud storage, etc.)
   * "olap" = OLAP engines (ClickHouse, DuckDB, etc.)
   */
  "x-category"?: ConnectorCategory;
  /**
   * Form height for the add data modal.
   * "tall" = larger form for connectors with more fields
   * "default" = standard form height
   */
  "x-form-height"?: "default" | "tall";
  /**
   * Backend connector name when different from schema name.
   * Used when a UI variant (e.g., "clickhousecloud") should map
   * to a different backend driver (e.g., "clickhouse").
   */
  "x-driver"?: string;
  /**
   * Custom button labels per field value.
   * Maps field key -> value -> button labels.
   * Example: { "connector_type": { "rill-managed": { idle: "Connect", loading: "Connecting..." } } }
   */
  "x-button-labels"?: Record<string, Record<string, ButtonLabels>>;
};

export type MultiStepFormSchema = JSONSchemaObject;
