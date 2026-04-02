import type { QueryClient } from "@tanstack/query-core";
import {
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGenerateFile,
  runtimeServiceGetFile,
} from "../../../runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "../../../runtime-client/v2";
import { replaceOrAddEnvVariable } from "../../connectors/code-utils";
import { OLAP_ENGINES } from "./constants";

const OLAP_SET = new Set(OLAP_ENGINES);

// OLAP per instance, populated by createConnectorSchemas() when schemas load.
// Avoids a redundant GetInstance call on first generateTemplate invocation.
const olapCache = new Map<string, string>();

/** Set the cached OLAP value for an instance. Called by connector-schemas.ts. */
export function setOlapCache(instanceId: string, olap: string) {
  olapCache.set(instanceId, olap);
}

/**
 * Resolve the template name from (driver, olap).
 * OLAP engine connectors have standalone templates (e.g., "clickhouse").
 * Source connectors use combined templates (e.g., "s3-duckdb", "athena-duckdb")
 * regardless of whether we're rendering the connector or model output.
 */
function resolveTemplateName(driver: string, olap: string): string {
  if (OLAP_SET.has(driver)) return driver;
  return `${driver}-${olap}`;
}

/**
 * Call the GenerateFile RPC to produce YAML + env var names from
 * structured form data. The backend handles rewrites, env var
 * naming/conflict resolution, and YAML formatting via declarative templates.
 *
 * Uses preview mode so the server renders without writing files;
 * the caller handles file writing and reconciliation.
 */
export async function generateTemplate(
  client: RuntimeClient,
  opts: {
    resourceType: string;
    driver: string;
    properties: Record<string, unknown>;
    connectorName?: string;
  },
): Promise<{ blob: string; envVars: Record<string, string> }> {
  // Resolve OLAP from cache (populated by createConnectorSchemas on modal open).
  // Falls back to "duckdb" if the cache is empty (shouldn't happen in normal flow).
  const olap = OLAP_SET.has(opts.driver)
    ? opts.driver
    : (olapCache.get(client.instanceId) ?? "duckdb");

  const templateName = resolveTemplateName(opts.driver, olap);

  const response = await runtimeServiceGenerateFile(client, {
    templateName,
    output: opts.resourceType,
    properties: opts.properties,
    connectorName: opts.connectorName,
    preview: true,
  });

  // Flatten to match the interface callers expect
  return {
    blob: response.files?.[0]?.blob ?? "",
    envVars: response.envVars ?? {},
  };
}

/**
 * Merge env vars returned by GenerateFile into the existing `.env` file.
 * The backend already resolved names and conflict suffixes, so this is
 * a straight key=value merge.
 *
 * Returns the updated blob and the original blob (for rollback).
 */
export async function mergeEnvVars(
  client: RuntimeClient,
  queryClient: QueryClient,
  envVars: Record<string, string>,
): Promise<{ newBlob: string; originalBlob: string }> {
  // Invalidate cache to get fresh .env content
  await queryClient.invalidateQueries({
    queryKey: getRuntimeServiceGetFileQueryKey(client.instanceId, {
      path: ".env",
    }),
  });

  let blob: string;
  let originalBlob: string;
  try {
    const file = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(client.instanceId, {
        path: ".env",
      }),
      queryFn: () => runtimeServiceGetFile(client, { path: ".env" }),
    });
    blob = file.blob || "";
    originalBlob = blob;
  } catch (error) {
    const msg =
      (error as any)?.message ?? (error as any)?.response?.data?.message ?? "";
    if (msg.includes("no such file")) {
      blob = "";
      originalBlob = "";
    } else {
      throw error;
    }
  }

  for (const [key, value] of Object.entries(envVars)) {
    if (!key || !value) continue;
    blob = replaceOrAddEnvVariable(blob, key, value);
  }

  return { newBlob: blob, originalBlob };
}
