import { QueryClient } from "@tanstack/svelte-query";
import {
  type ConnectorDriverProperty,
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceGetInstanceQueryKey,
  runtimeServiceAnalyzeConnectors,
  runtimeServiceGetFile,
  runtimeServiceGetInstance,
  runtimeServicePutFile,
  type V1ConnectorDriver,
} from "../../runtime-client";
import type { RuntimeClient } from "../../runtime-client/v2";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import {
  getName,
  isNonStandardIdentifier,
} from "@rilldata/web-common/features/entity-management/name-utils";
import {
  getDriverNameForConnector,
  makeSufficientlyQualifiedTableName,
} from "./connectors-utils";
import { getDocsCategory } from "../sources/modal/connector-schemas";
import type { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";
import {
  applyDuckLakeFormPipeline,
  injectDuckLakeAttach,
} from "@rilldata/web-common/features/templates/schemas/ducklake-utils.ts";
import { filterSchemaValuesForSubmit } from "@rilldata/web-common/features/templates/schema-utils.ts";
import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";

function yamlModelTemplate(driverName: string) {
  return `# Model YAML
# Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/${driverName}

type: model
materialize: true

connector: {{ connector }}

sql: {{ sql }}{{ dev_section }}

output:
  connector: {{ output_connector }}
`;
}

const SENSITIVE_HEADER_PATTERN =
  /^(authorization|x-api-key|api-key|token|x-token|x-auth|x-secret|proxy-authorization)$/i;

/**
 * Returns true when a header key likely carries a secret value (e.g. tokens,
 * API keys). Only these headers are stored in `.env`; the rest stay as plain
 * text in the connector YAML.
 */
function isSensitiveHeaderKey(headerKey: string): boolean {
  return SENSITIVE_HEADER_PATTERN.test(headerKey.trim());
}

/**
 * Sanitize a header key into a valid .env variable segment.
 * Lowercases, replaces non-alphanumeric characters with underscores, and
 * collapses consecutive underscores.
 */
function headerKeyToEnvSegment(headerKey: string): string {
  return headerKey
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_|_$/g, "");
}

/**
 * Common HTTP authentication scheme prefixes.
 * When a sensitive header value starts with one of these (case-insensitive),
 * only the token portion after the prefix is stored in `.env`, while the
 * scheme keyword is kept as plain text in the connector YAML.
 */
const AUTH_SCHEME_PREFIXES = ["Bearer ", "Basic ", "Token ", "Bot "];

/**
 * If `value` begins with a recognised auth scheme prefix (e.g. "Bearer "),
 * returns `{ scheme, secret }` where `scheme` includes the trailing space.
 * Returns `null` when no known prefix is detected.
 */
function splitAuthSchemePrefix(
  value: string,
): { scheme: string; secret: string } | null {
  for (const prefix of AUTH_SCHEME_PREFIXES) {
    if (
      value.length > prefix.length &&
      value.slice(0, prefix.length).toLowerCase() === prefix.toLowerCase()
    ) {
      return {
        scheme: value.slice(0, prefix.length),
        secret: value.slice(prefix.length),
      };
    }
  }
  return null;
}

/**
 * Convert header entries into a YAML map block.
 * Accepts an array of {key, value} objects (new key-value input) or a legacy
 * multi-line "Header-Name: value" string. Returns empty string when there are
 * no valid entries.
 *
 * When `driverName` is provided, header values are replaced with
 * `{{ .env.connector.<name>.<header_key> }}` references so that secrets are
 * stored in `.env` rather than in the connector YAML file.
 */
export function formatHeadersAsYamlMap(
  value: Array<{ key: string; value: string }> | string,
  envEditSession: EnvEditSession,
): string {
  if (typeof value === "string") {
    // Legacy textarea format: parse "Key: Value" lines
    const lines = value
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line.includes(":"));
    if (lines.length === 0) return "";
    const entries = lines.map((line) => {
      const idx = line.indexOf(":");
      const k = line.substring(0, idx).trim().replace(/^"|"$/g, "");
      const raw = line
        .substring(idx + 1)
        .trim()
        .replace(/^"|"$/g, "");
      let v: string;
      if (isSensitiveHeaderKey(k)) {
        const split = splitAuthSchemePrefix(raw);
        const entry = envEditSession.acquire(
          headerKeyToEnvSegment(k),
          split ? split.secret : raw,
        );
        const envRef = `{{ .env.${entry.mappedEnvVarName} }}`;
        v = split ? `${split.scheme}${envRef}` : envRef;
      } else {
        v = raw;
      }
      return `    "${k}": "${v}"`;
    });
    return `headers:\n${entries.join("\n")}`;
  }

  // Array of {key, value} objects from key-value input
  const valid = value.filter((e) => e.key.trim() !== "");
  if (valid.length === 0) return "";
  const entries = valid.map((e) => {
    const k = e.key.trim();
    let v: string;
    const trimmedVal = e.value.trim();
    if (isSensitiveHeaderKey(k)) {
      const split = splitAuthSchemePrefix(trimmedVal);
      const entry = envEditSession.acquire(
        headerKeyToEnvSegment(k),
        split ? split.secret : trimmedVal,
      );
      const envRef = `{{ .env.${entry.mappedEnvVarName} }}`;
      v = split ? `${split.scheme}${envRef}` : envRef;
    } else {
      v = trimmedVal;
    }
    return `    "${k}": "${v}"`;
  });
  return `headers:\n${entries.join("\n")}`;
}

export function compileConnectorYAML(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  envEditSession: EnvEditSession,
  options?: {
    fieldFilter?: (
      property:
        | ConnectorDriverProperty
        | { key?: string; type?: string; secret?: boolean; internal?: boolean },
    ) => boolean;
    orderedProperties?: Array<
      | ConnectorDriverProperty
      | { key?: string; type?: string; secret?: boolean; internal?: boolean }
    >;
    connectorInstanceName?: string;
    secretKeys?: string[];
    stringKeys?: string[];
    schema?: MultiStepFormSchema;
  },
) {
  // Add instructions to the top of the file
  const driverName = getDriverNameForConnector(connector.name as string);
  const category = connector.implementsAi
    ? "ai"
    : connector.implementsOlap
      ? "olap"
      : undefined;
  const topOfFile = `# Connector YAML
# Reference documentation: https://docs.rilldata.com/developers/build/connectors/${getDocsCategory(category)}/${driverName}

type: connector

driver: ${driverName}`;

  // Use the provided orderedProperties if available.
  let properties = options?.orderedProperties ?? [];

  // Optionally filter properties
  if (options?.fieldFilter) {
    properties = properties.filter(options.fieldFilter);
  }

  // Get the secret property keys
  const secretPropertyKeys = options?.secretKeys ?? [];
  envEditSession.startEdit();

  // Apply ducklake transforms
  formValues = applyDuckLakeFormPipeline(options?.schema, formValues, {
    connectorName: connector.name ?? "",
    envEditSession,
  });
  formValues = options?.schema
    ? injectDuckLakeAttach(
        options.schema,
        filterSchemaValuesForSubmit(options.schema, formValues, {
          step: "connector",
        }),
        formValues,
      )
    : formValues;

  // Get the string property keys
  const stringPropertyKeys = options?.stringKeys ?? [];

  // Compile key value pairs in the order of properties
  const compiledKeyValues = properties
    .filter((property) => {
      if (!property.key) return false;
      const value = formValues[property.key];
      if (value === undefined) return false;
      // Filter out empty strings for optional fields
      if (typeof value === "string" && value.trim() === "") return false;
      // Filter out empty arrays (e.g. key-value inputs with no entries)
      if (Array.isArray(value) && value.length === 0) return false;
      // For ClickHouse, exclude managed: false as it's the default behavior
      // When managed=false, it's the default self-managed mode and doesn't need to be explicit
      if (
        connector.name === "clickhouse" &&
        property.key === "managed" &&
        value === false
      )
        return false;
      // For advanced fields, skip values that match the field's effective default.
      const schemaProp = options?.schema?.properties?.[property.key];
      if (schemaProp?.["x-advanced"]) {
        const typeDefault =
          schemaProp.type === "boolean"
            ? false
            : schemaProp.type === "number" ||
                (schemaProp.type as any) === "integer"
              ? 0
              : schemaProp.type === "string"
                ? ""
                : undefined;
        const effectiveDefault =
          schemaProp.default !== undefined ? schemaProp.default : typeDefault;
        if (effectiveDefault !== undefined && value === effectiveDefault)
          return false;
      }
      return true;
    })
    .map((property) => {
      const key = property.key as string;
      const value = formValues[key];

      if (key === "headers") {
        return formatHeadersAsYamlMap(
          value as Array<{ key: string; value: string }> | string,
          envEditSession,
        );
      }

      const isSecretProperty = secretPropertyKeys.includes(key);
      if (isSecretProperty) {
        const entry = envEditSession.acquire(key, String(value));
        return `${key}: "{{ .env.${entry.mappedEnvVarName} }}"`; // uses standard Go template syntax
      }

      // For boolean fields with x-yaml-value, emit the mapped value instead of true/false.
      // Object form maps both toggle states so each round-trips to YAML; scalar form
      // emits only when the toggle is checked.
      const schemaPropForMap = options?.schema?.properties?.[key];
      const yamlValueRule = schemaPropForMap?.["x-yaml-value"];
      if (
        yamlValueRule !== null &&
        typeof yamlValueRule === "object" &&
        typeof value === "boolean"
      ) {
        const mapped = yamlValueRule[value ? "true" : "false"];
        if (mapped !== undefined) return `${key}: ${mapped}`;
      } else if (yamlValueRule !== undefined && value === true) {
        return `${key}: ${yamlValueRule}`;
      }

      const isStringProperty = stringPropertyKeys.includes(key);
      if (isStringProperty) {
        return `${key}: "${value}"`;
      }

      return `${key}: ${value}`;
    })
    .filter((line) => line !== "")
    .join("\n");

  // Return the compiled YAML
  return `${topOfFile}\n` + compiledKeyValues;
}

export async function updateRillYAMLWithOlapConnector(
  client: RuntimeClient,
  queryClient: QueryClient,
  newConnector: string,
): Promise<string> {
  // Get the existing rill.yaml file
  const file = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetFileQueryKey(client.instanceId, {
      path: "rill.yaml",
    }),
    queryFn: () => runtimeServiceGetFile(client, { path: "rill.yaml" }),
  });
  const blob = file.blob || "";

  // Update the blob with the new OLAP connector
  return replaceOlapConnectorInYAML(blob, newConnector);
}

/**
 * Update the `olap_connector` key in a YAML file.
 * This function uses a regex approach to preserve comments and formatting.
 */
export function replaceOlapConnectorInYAML(
  blob: string,
  newConnector: string,
): string {
  const olapConnectorRegex = /^olap_connector: .+$/m;

  if (olapConnectorRegex.test(blob)) {
    return blob.replace(olapConnectorRegex, `olap_connector: ${newConnector}`);
  } else {
    return `${blob}${blob !== "" ? "\n" : ""}olap_connector: ${newConnector}\n`;
  }
}

export function maybeUnsetOlapConnectorInYaml(
  blob: string,
  connectorName: string,
): [boolean, string] {
  const olapConnectorRegex = new RegExp(
    `^\\s*olap_connector:\\s+${connectorName}\\s*$`,
    "m",
  );
  if (!olapConnectorRegex.test(blob)) return [false, blob];
  return [true, blob.replace(olapConnectorRegex, "")];
}

export async function updateRillYAMLWithAiConnector(
  client: RuntimeClient,
  queryClient: QueryClient,
  newConnector: string,
): Promise<string> {
  const file = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetFileQueryKey(client.instanceId, {
      path: "rill.yaml",
    }),
    queryFn: () => runtimeServiceGetFile(client, { path: "rill.yaml" }),
  });
  const blob = file.blob || "";
  return replaceAiConnectorInYAML(blob, newConnector);
}

/**
 * Update the `ai_connector` key in a YAML file.
 * This function uses a regex approach to preserve comments and formatting.
 */
export function replaceAiConnectorInYAML(
  blob: string,
  newConnector: string,
): string {
  const aiConnectorRegex = /^ai_connector: .+$/m;

  if (aiConnectorRegex.test(blob)) {
    return blob.replace(aiConnectorRegex, `ai_connector: ${newConnector}`);
  } else {
    return `${blob}${blob !== "" ? "\n" : ""}ai_connector: ${newConnector}\n`;
  }
}

export async function createYamlModelFromTable(
  client: RuntimeClient,
  queryClient: QueryClient,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
): Promise<[string, string]> {
  // Get driver name for makeSufficientlyQualifiedTableName
  const analyzeConnectorsQueryKey = getRuntimeServiceAnalyzeConnectorsQueryKey(
    client.instanceId,
    {},
  );
  const analyzeConnectorsQueryFn = async () =>
    runtimeServiceAnalyzeConnectors(client, {});
  const connectors = await queryClient.fetchQuery({
    queryKey: analyzeConnectorsQueryKey,
    queryFn: analyzeConnectorsQueryFn,
  });
  const analyzedConnector = connectors?.connectors?.find(
    (c) => c.name === connector,
  );
  if (!analyzedConnector) {
    throw new Error(`Could not find connector ${connector}`);
  }
  const driverName = analyzedConnector.driver?.name as string;

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${table}_model`, allNames);
  const newModelPath = `models/${newModelName}.yaml`;

  // Get sufficiently qualified table name
  const sufficientlyQualifiedTableName = makeSufficientlyQualifiedTableName(
    driverName,
    database,
    databaseSchema,
    table,
  );

  // Use the sufficiently qualified table name directly
  const selectStatement = `select * from ${sufficientlyQualifiedTableName}`;

  // Get default OLAP connector for the output
  const runtimeInstance = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetInstanceQueryKey(client.instanceId, {}),
    queryFn: () => runtimeServiceGetInstance(client, { sensitive: true }),
  });
  if (!runtimeInstance) {
    throw new Error(`Could not find runtime instance ${client.instanceId}`);
  }
  const defaultOLAP = runtimeInstance?.instance?.olapConnector || "duckdb";

  // NOTE: Redshift does not support LIMIT clauses in its UNLOAD data exports.
  const shouldIncludeDevSection = driverName !== "redshift";
  const devSection = shouldIncludeDevSection
    ? `\n\ndev:\n  sql: ${selectStatement} limit 10000`
    : "";

  const yamlContent = yamlModelTemplate(driverName)
    .replace("{{ connector }}", connector)
    .replace(/{{ sql }}/g, selectStatement)
    .replace("{{ dev_section }}", devSection)
    .replace("{{ output_connector }}", defaultOLAP);

  // Write the YAML file
  await runtimeServicePutFile(client, {
    path: newModelPath,
    blob: yamlContent,
    createOnly: true,
  });

  // Invalidate relevant queries
  await queryClient.invalidateQueries({
    queryKey: ["runtimeServiceListFiles", client.instanceId],
  });

  return ["/" + newModelPath, newModelName];
}

export async function createSqlModelFromTable(
  client: RuntimeClient,
  queryClient: QueryClient,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
  addDevLimit: boolean = true,
): Promise<[string, string]> {
  // Get driver name
  const analyzeConnectorsQueryKey = getRuntimeServiceAnalyzeConnectorsQueryKey(
    client.instanceId,
    {},
  );
  const analyzeConnectorsQueryFn = async () =>
    runtimeServiceAnalyzeConnectors(client, {});
  const connectors = await queryClient.fetchQuery({
    queryKey: analyzeConnectorsQueryKey,
    queryFn: analyzeConnectorsQueryFn,
  });
  const analyzedConnector = connectors?.connectors?.find(
    (c) => c.name === connector,
  );
  if (!analyzedConnector) {
    throw new Error(`Could not find connector ${connector}`);
  }
  const driverName = analyzedConnector.driver?.name as string;

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${table}_model`, allNames);
  const newModelPath = `models/${newModelName}.sql`;

  // Get sufficiently qualified table name
  const sufficientlyQualifiedTableName = makeSufficientlyQualifiedTableName(
    driverName,
    database,
    databaseSchema,
    table,
  );

  // Create model — OLAP models use the same connector for both source and output
  const topComments = `-- Model SQL\n-- Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/${driverName}`;
  const connectorLine = `-- @connector: ${connector}`;
  const outputConnectorLine = `-- @output.connector: ${connector}`;
  const selectStatement = isNonStandardIdentifier(
    sufficientlyQualifiedTableName,
  )
    ? `select * from "${sufficientlyQualifiedTableName}"`
    : `select * from ${sufficientlyQualifiedTableName}`;
  const devLimit = "{{ if dev }} limit 100000 {{ end}}";

  let modelSQL = `${topComments}\n`;
  modelSQL += `${connectorLine}\n`;
  modelSQL += `${outputConnectorLine}\n`;

  modelSQL += `\n${selectStatement}`;

  if (addDevLimit) {
    modelSQL += `\n${devLimit}`;
  }

  await runtimeServicePutFile(client, {
    path: newModelPath,
    blob: modelSQL,
    createOnly: true,
  });

  eventBus.emit("notification", {
    message: `Queried ${table} in workspace`,
  });

  // Done
  return ["/" + newModelPath, newModelName];
}
