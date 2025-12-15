export type AddDataFormType = "source" | "connector";

export type ConnectorType = "parameters" | "dsn";

export type AuthOption = {
  value: string;
  label: string;
  description: string;
  hint?: string;
};

export type AuthField =
  | {
      type: "credentials";
      id: string;
      hint?: string;
      optional?: boolean;
      accept?: string;
    }
  | {
      type: "input";
      id: string;
      label: string;
      placeholder?: string;
      optional?: boolean;
      secret?: boolean;
      hint?: string;
    };

type JSONSchemaVisibleIfValue =
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
};

export type JSONSchemaCondition = {
  properties?: Record<string, { const?: string | number | boolean }>;
};

export type JSONSchemaConstraint = {
  required?: string[];
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

export type MultiStepFormConfig = {
  schema: MultiStepFormSchema;
  authMethodKey: string;
  authOptions: AuthOption[];
  clearFieldsByMethod: Record<string, string[]>;
  excludedKeys: string[];
  authFieldGroups: Record<string, AuthField[]>;
  requiredFieldsByMethod: Record<string, string[]>;
  fieldLabels: Record<string, string>;
  defaultAuthMethod?: string;
};
