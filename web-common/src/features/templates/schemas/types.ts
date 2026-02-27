import type { ComponentType, SvelteComponent } from "svelte";

export type ConnectorIcon = ComponentType<SvelteComponent>;

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
  /** Render style override for the field (e.g. radio buttons, tabs, file picker). */
  "x-display"?: "radio" | "select" | "textarea" | "file" | "tabs" | "key-value";
  /** Visual style for select fields. "rich" renders with icons and colored cards. */
  "x-select-style"?: "standard" | "rich";
  /** Render the field value in a monospace font. */
  "x-monospace"?: boolean;
  /** Which modal step this field belongs to. */
  "x-step"?: "connector" | "source" | "explorer";
  /** Field holds a secret value that should be stored in .env, not in YAML. */
  "x-secret"?: boolean;
  /** Show this field only when other fields match the given values. */
  "x-visible-if"?: Record<string, JSONSchemaVisibleIfValue>;
  /** Disable this field (read-only) when other fields match the given values. */
  "x-disabled-if"?: Record<string, JSONSchemaVisibleIfValue>;
  /** Human-readable labels for each enum option, in the same order as `enum`. */
  "x-enum-labels"?: string[];
  /** Descriptive text for each enum option, in the same order as `enum`. */
  "x-enum-descriptions"?: string[];
  /** Icon identifiers for each enum option, in the same order as `enum`. */
  "x-enum-icons"?: string[];
  /** Placeholder text shown in the input when empty. */
  "x-placeholder"?: string;
  /** Helper text displayed below the input. */
  "x-hint"?: string;
  /** @deprecated Use "x-file-accept" instead. */
  "x-accept"?: string;
  /** Accepted file types for file inputs (e.g. ".json,.pem"). */
  "x-file-accept"?: string;
  /** How to encode file content: base64, json (parse+stringify), or raw (pass-through). */
  "x-file-encoding"?: "base64" | "json" | "raw";
  /** Extract values from parsed file content into other form fields. Maps form field key to JSON property name. */
  "x-file-extract"?: Record<string, string>;
  /** Field is read-only and shown for informational purposes only. */
  "x-informational"?: boolean;
  /** URL to external documentation for this field, shown as a help link. */
  "x-docs-url"?: string;
  /** Field controls UI behavior only and is excluded from generated YAML. */
  "x-ui-only"?: boolean;
  /**
   * Explicit grouping for radio/select options: maps an option value to the
   * child field keys that should render beneath that option.
   */
  "x-grouped-fields"?: Record<string, string[]>;
  /**
   * Group fields under tab options for enum-driven tab layouts.
   */
  "x-tab-group"?: Record<string, string[]>;
  /**
   * Explicit environment variable name for secret fields.
   * When set, this name is used instead of computing it from driver + property key.
   */
  "x-env-var-name"?: string;
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
   * Form width for the add data modal.
   * "wide" = wider form for connectors with templates or more content
   * "default" = standard form width
   */
  "x-form-width"?: "default" | "wide";
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
  /** Full-size icon component for the connector (used in add-data grid). */
  "x-icon"?: ConnectorIcon;
  /** Small icon component for the connector (used in nav, cards, dialogs). */
  "x-small-icon"?: ConnectorIcon;
};

export type MultiStepFormSchema = JSONSchemaObject;
