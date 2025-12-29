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
  "x-display"?: "radio" | "select" | "textarea" | "file";
  "x-step"?: "connector" | "source";
  "x-secret"?: boolean;
  "x-visible-if"?: Record<string, JSONSchemaVisibleIfValue>;
  "x-enum-labels"?: string[];
  "x-enum-descriptions"?: string[];
  "x-placeholder"?: string;
  "x-hint"?: string;
  "x-accept"?: string;
  /**
   * Explicit grouping for radio/select options: maps an option value to the
   * child field keys that should render beneath that option.
   */
  "x-grouped-fields"?: Record<string, string[]>;
  // Allow custom keywords such as errorMessage or future x-extensions.
  [key: string]: unknown;
};

export type JSONSchemaCondition = {
  properties?: Record<string, { const?: string | number | boolean }>;
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

export type JSONSchemaObject = {
  $schema?: string;
  type: "object";
  title?: string;
  description?: string;
  properties?: Record<string, JSONSchemaField>;
  required?: string[];
  allOf?: JSONSchemaConditional[];
};

export type MultiStepFormSchema = JSONSchemaObject;
