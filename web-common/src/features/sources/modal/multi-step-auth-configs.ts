import type {
  AuthField,
  AuthOption,
  JSONSchemaConditional,
  JSONSchemaField,
  MultiStepFormConfig,
  MultiStepFormSchema,
} from "./types";

type VisibleIf = Record<
  string,
  string | number | boolean | Array<string | number | boolean>
>;

export const multiStepFormSchemas: Record<string, MultiStepFormSchema> = {
  s3: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    properties: {
      auth_method: {
        type: "string",
        title: "Authentication method",
        description: "Choose how to authenticate to S3",
        enum: ["access_keys", "public"],
        default: "access_keys",
        "x-display": "radio",
        "x-enum-labels": ["Access keys", "Public"],
        "x-enum-descriptions": [
          "Use AWS access key ID and secret access key.",
          "Access publicly readable buckets without credentials.",
        ],
        "x-step": "connector",
      },
      aws_access_key_id: {
        type: "string",
        title: "Access Key ID",
        description: "AWS access key ID for the bucket",
        "x-placeholder": "Enter AWS access key ID",
        "x-secret": true,
        "x-step": "connector",
        "x-visible-if": { auth_method: "access_keys" },
      },
      aws_secret_access_key: {
        type: "string",
        title: "Secret Access Key",
        description: "AWS secret access key for the bucket",
        "x-placeholder": "Enter AWS secret access key",
        "x-secret": true,
        "x-step": "connector",
        "x-visible-if": { auth_method: "access_keys" },
      },
      region: {
        type: "string",
        title: "Region",
        description:
          "Rill uses your default AWS region unless you set it explicitly.",
        "x-placeholder": "us-east-1",
        "x-step": "connector",
        "x-visible-if": { auth_method: "access_keys" },
      },
      endpoint: {
        type: "string",
        title: "Endpoint",
        description:
          "Override the S3 endpoint (for S3-compatible services like R2/MinIO).",
        "x-placeholder": "https://s3.example.com",
        "x-step": "connector",
        "x-visible-if": { auth_method: "access_keys" },
      },
      path: {
        type: "string",
        title: "S3 URI",
        description: "Path to your S3 bucket or prefix",
        pattern: "^s3://",
        "x-step": "source",
      },
      name: {
        type: "string",
        title: "Model name",
        description: "Name for the source model",
        pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
        "x-step": "source",
      },
    },
    allOf: [
      {
        if: { properties: { auth_method: { const: "access_keys" } } },
        then: {
          required: ["aws_access_key_id", "aws_secret_access_key"],
        },
      },
    ],
  },
  gcs: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    properties: {
      auth_method: {
        type: "string",
        title: "Authentication method",
        enum: ["credentials", "hmac", "public"],
        default: "credentials",
        description: "Choose how to authenticate to GCS",
        "x-display": "radio",
        "x-enum-labels": ["GCP credentials", "HMAC keys", "Public"],
        "x-enum-descriptions": [
          "Upload a JSON key file for a service account with GCS access.",
          "Use HMAC access key and secret for S3-compatible authentication.",
          "Access publicly readable buckets without credentials.",
        ],
        "x-step": "connector",
      },
      google_application_credentials: {
        type: "string",
        title: "Service account key",
        description:
          "Upload a JSON key file for a service account with GCS access.",
        format: "file",
        "x-display": "file",
        "x-accept": ".json",
        "x-step": "connector",
        "x-visible-if": { auth_method: "credentials" },
      },
      key_id: {
        type: "string",
        title: "Access Key ID",
        description: "HMAC access key ID for S3-compatible authentication",
        "x-placeholder": "Enter your HMAC access key ID",
        "x-secret": true,
        "x-step": "connector",
        "x-visible-if": { auth_method: "hmac" },
      },
      secret: {
        type: "string",
        title: "Secret Access Key",
        description: "HMAC secret access key for S3-compatible authentication",
        "x-placeholder": "Enter your HMAC secret access key",
        "x-secret": true,
        "x-step": "connector",
        "x-visible-if": { auth_method: "hmac" },
      },
      path: {
        type: "string",
        title: "GCS URI",
        description: "Path to your GCS bucket or prefix",
        pattern: "^gs://",
        "x-step": "source",
      },
      name: {
        type: "string",
        title: "Model name",
        description: "Name for the source model",
        pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
        "x-step": "source",
      },
    },
    allOf: [
      {
        if: { properties: { auth_method: { const: "credentials" } } },
        then: { required: ["google_application_credentials"] },
      },
      {
        if: { properties: { auth_method: { const: "hmac" } } },
        then: { required: ["key_id", "secret"] },
      },
    ],
  },
  azure: {
    $schema: "http://json-schema.org/draft-07/schema#",
    type: "object",
    properties: {
      auth_method: {
        type: "string",
        title: "Authentication method",
        enum: ["connection_string", "account_key", "sas_token", "public"],
        default: "connection_string",
        description: "Choose how to authenticate to Azure Blob Storage",
        "x-display": "radio",
        "x-enum-labels": [
          "Connection String",
          "Storage Account Key",
          "SAS Token",
          "Public",
        ],
        "x-enum-descriptions": [
          "Provide a full Azure Storage connection string.",
          "Provide the storage account name and access key.",
          "Provide the storage account name and SAS token.",
          "Access publicly readable blobs without credentials.",
        ],
        "x-step": "connector",
      },
      azure_storage_connection_string: {
        type: "string",
        title: "Connection string",
        description: "Paste an Azure Storage connection string",
        "x-placeholder": "Enter Azure storage connection string",
        "x-secret": true,
        "x-step": "connector",
        "x-visible-if": { auth_method: "connection_string" },
      },
      azure_storage_account: {
        type: "string",
        title: "Storage account",
        description: "The name of the Azure storage account",
        "x-placeholder": "Enter Azure storage account",
        "x-step": "connector",
        "x-visible-if": { auth_method: ["account_key", "sas_token"] },
      },
      azure_storage_key: {
        type: "string",
        title: "Access key",
        description: "Primary or secondary access key for the storage account",
        "x-placeholder": "Enter Azure storage access key",
        "x-secret": true,
        "x-step": "connector",
        "x-visible-if": { auth_method: "account_key" },
      },
      azure_storage_sas_token: {
        type: "string",
        title: "SAS token",
        description:
          "Shared Access Signature token for the storage account (starting with ?sv=)",
        "x-placeholder": "Enter Azure SAS token",
        "x-secret": true,
        "x-step": "connector",
        "x-visible-if": { auth_method: "sas_token" },
      },
      path: {
        type: "string",
        title: "Blob URI",
        description:
          "URI to the Azure blob container or directory (e.g., https://<account>.blob.core.windows.net/container)",
        pattern: "^https?://",
        "x-step": "source",
      },
      name: {
        type: "string",
        title: "Model name",
        description: "Name for the source model",
        pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
        "x-step": "source",
      },
    },
    allOf: [
      {
        if: { properties: { auth_method: { const: "connection_string" } } },
        then: { required: ["azure_storage_connection_string"] },
      },
      {
        if: { properties: { auth_method: { const: "account_key" } } },
        then: { required: ["azure_storage_account", "azure_storage_key"] },
      },
      {
        if: { properties: { auth_method: { const: "sas_token" } } },
        then: {
          required: ["azure_storage_account", "azure_storage_sas_token"],
        },
      },
    ],
  },
};

export function getMultiStepFormConfig(
  connectorName: string,
): MultiStepFormConfig | null {
  const schema =
    multiStepFormSchemas[connectorName as keyof typeof multiStepFormSchemas];
  if (!schema?.properties) return null;

  const authMethodKey = findAuthMethodKey(schema);
  if (!authMethodKey) return null;

  const authProperty = schema.properties[authMethodKey];
  const authOptions = buildAuthOptions(authProperty);
  if (!authOptions.length) return null;

  const defaultAuthMethod =
    authProperty.default !== undefined && authProperty.default !== null
      ? String(authProperty.default)
      : authOptions[0]?.value;

  const requiredByMethod = buildRequiredByMethod(
    schema,
    authMethodKey,
    authOptions.map((o) => o.value),
  );
  const authFieldGroups = buildAuthFieldGroups(
    schema,
    authMethodKey,
    authOptions,
    requiredByMethod,
  );
  const excludedKeys = buildExcludedKeys(
    schema,
    authMethodKey,
    authFieldGroups,
  );
  const clearFieldsByMethod = buildClearFieldsByMethod(
    schema,
    authMethodKey,
    authOptions,
  );

  return {
    schema,
    authMethodKey,
    authOptions,
    defaultAuthMethod: defaultAuthMethod || undefined,
    clearFieldsByMethod,
    excludedKeys,
    authFieldGroups,
    requiredFieldsByMethod: requiredByMethod,
    fieldLabels: buildFieldLabels(schema),
  };
}

function findAuthMethodKey(schema: MultiStepFormSchema): string | null {
  if (!schema.properties) return null;
  for (const [key, value] of Object.entries(schema.properties)) {
    if (value.enum && value["x-display"] === "radio") {
      return key;
    }
  }
  return schema.properties.auth_method ? "auth_method" : null;
}

function buildAuthOptions(authProperty: JSONSchemaField): AuthOption[] {
  if (!authProperty.enum) return [];
  const labels = authProperty["x-enum-labels"] ?? [];
  const descriptions = authProperty["x-enum-descriptions"] ?? [];
  return authProperty.enum.map((value, idx) => ({
    value: String(value),
    label: labels[idx] ?? String(value),
    description:
      descriptions[idx] ?? authProperty.description ?? "Choose an option",
    hint: authProperty["x-hint"],
  }));
}

function buildRequiredByMethod(
  schema: MultiStepFormSchema,
  authMethodKey: string,
  methods: string[],
): Record<string, string[]> {
  const conditionals = schema.allOf ?? [];
  const baseRequired = new Set(schema.required ?? []);
  const result: Record<string, string[]> = {};

  for (const method of methods) {
    const required = new Set<string>(baseRequired);
    for (const conditional of conditionals) {
      if (!matchesAuthMethod(conditional, authMethodKey, method)) {
        conditional.else?.required?.forEach((field) => required.add(field));
        continue;
      }
      conditional.then?.required?.forEach((field) => required.add(field));
    }
    result[method] = Array.from(required);
  }

  return result;
}

function matchesAuthMethod(
  conditional: JSONSchemaConditional,
  authMethodKey: string,
  method: string,
) {
  const constValue =
    conditional.if?.properties?.[authMethodKey as keyof VisibleIf]?.const;
  if (constValue === undefined || constValue === null) return false;
  return String(constValue) === method;
}

function buildAuthFieldGroups(
  schema: MultiStepFormSchema,
  authMethodKey: string,
  authOptions: AuthOption[],
  requiredByMethod: Record<string, string[]>,
): Record<string, AuthField[]> {
  const groups: Record<string, AuthField[]> = {};
  const properties = schema.properties ?? {};

  for (const option of authOptions) {
    const required = new Set(requiredByMethod[option.value] ?? []);
    for (const [key, prop] of Object.entries(properties)) {
      if (key === authMethodKey) continue;
      if (!isConnectorStep(prop)) continue;
      if (!isVisibleForMethod(prop, authMethodKey, option.value)) continue;

      const field: AuthField = toAuthField(key, prop, required.has(key));
      groups[option.value] = [...(groups[option.value] ?? []), field];
    }
  }

  return groups;
}

function buildClearFieldsByMethod(
  schema: MultiStepFormSchema,
  authMethodKey: string,
  authOptions: AuthOption[],
): Record<string, string[]> {
  const properties = schema.properties ?? {};
  const clear: Record<string, string[]> = {};

  for (const option of authOptions) {
    const fields: string[] = [];
    for (const [key, prop] of Object.entries(properties)) {
      if (key === authMethodKey) continue;
      if (!isVisibleForMethod(prop, authMethodKey, option.value)) {
        fields.push(key);
      }
    }
    clear[option.value] = fields;
  }

  return clear;
}

function buildExcludedKeys(
  schema: MultiStepFormSchema,
  authMethodKey: string,
  authFieldGroups: Record<string, AuthField[]>,
): string[] {
  const excluded = new Set<string>([authMethodKey]);
  const properties = schema.properties ?? {};
  const groupedFieldKeys = new Set(
    Object.values(authFieldGroups)
      .flat()
      .map((field) => field.id),
  );

  for (const [key, prop] of Object.entries(properties)) {
    const step = prop["x-step"];
    if (step === "source") excluded.add(key);
    if (groupedFieldKeys.has(key)) excluded.add(key);
  }

  return Array.from(excluded);
}

function buildFieldLabels(schema: MultiStepFormSchema): Record<string, string> {
  const labels: Record<string, string> = {};
  for (const [key, prop] of Object.entries(schema.properties ?? {})) {
    if (prop.title) labels[key] = prop.title;
  }
  return labels;
}

function isConnectorStep(prop: JSONSchemaField): boolean {
  return (prop["x-step"] ?? "connector") === "connector";
}

function isVisibleForMethod(
  prop: JSONSchemaField,
  authMethodKey: string,
  method: string,
): boolean {
  const conditions = prop["x-visible-if"];
  if (!conditions) return true;

  const authCondition = conditions[authMethodKey];
  if (authCondition === undefined) return true;
  if (Array.isArray(authCondition)) {
    return authCondition.map(String).includes(method);
  }
  return String(authCondition) === method;
}

function toAuthField(
  key: string,
  prop: JSONSchemaField,
  isRequired: boolean,
): AuthField {
  const base = {
    id: key,
    optional: !isRequired,
    hint: prop.description ?? prop["x-hint"],
  };

  if (prop["x-display"] === "file" || prop.format === "file") {
    return {
      type: "credentials",
      accept: prop["x-accept"],
      ...base,
    };
  }

  return {
    type: "input",
    label: prop.title ?? key,
    placeholder: prop["x-placeholder"],
    secret: prop["x-secret"],
    ...base,
  };
}
