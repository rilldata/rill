import type { MultiStepFormSchema } from "../../templates/schemas/types";

export type {
  JSONSchemaCondition,
  JSONSchemaConditional,
  JSONSchemaConstraint,
  JSONSchemaField,
  JSONSchemaObject,
  MultiStepFormSchema,
} from "../../templates/schemas/types";

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
