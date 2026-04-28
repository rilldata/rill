import type { QueryClient } from "@tanstack/query-core";
import type { RuntimeClient } from "../../../runtime-client/v2";
import {
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGenerateFile,
  runtimeServiceGetFile,
} from "../../../runtime-client/v2/gen/runtime-service";
import { replaceOrAddEnvVariable } from "../../connectors/code-utils";
import { OLAP_ENGINES } from "./constants";

const OLAP_SET = new Set<string>(OLAP_ENGINES);

// OLAP per instance, populated by AddDataModal when the instance OLAP is known.
// Avoids a redundant GetInstance round-trip on first generateTemplate invocation.
const olapCache = new Map<string, string>();

/** Set the cached OLAP value for an instance. */
export function setOlapCache(instanceId: string, olap: string) {
  olapCache.set(instanceId, olap);
}

/** Test seam: clear the OLAP cache between tests. */
export function _clearOlapCache() {
  olapCache.clear();
}

/**
 * Resolve the template name from (driver, olap).
 * OLAP engine drivers have standalone templates (e.g. "clickhouse").
 * Source drivers use combined templates (e.g. "s3-duckdb", "postgres-clickhouse"),
 * regardless of whether we're rendering the connector or model output.
 */
function resolveTemplateName(driver: string, olap: string): string {
  if (OLAP_SET.has(driver)) return driver;
  return `${driver}-${olap}`;
}

/**
 * Call the GenerateFile RPC to produce YAML and env-var names from
 * structured form data. The backend handles env-var naming, conflict
 * suffixes, and YAML formatting via declarative templates.
 *
 * Always uses preview mode so the server renders without writing files;
 * the caller is responsible for persisting the YAML and `.env`.
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
  // Resolve OLAP from cache (populated when the modal mounts).
  // Falls back to "duckdb" if the cache is empty (shouldn't happen in practice).
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

  return {
    blob: response.files?.[0]?.blob ?? "",
    envVars: response.envVars ?? {},
  };
}

/**
 * Merge env vars returned by GenerateFile into the existing `.env` file.
 * The backend has already resolved names and conflict suffixes, so this
 * is a straight key=value merge.
 *
 * Returns the updated blob and the original blob (for rollback).
 */
export async function mergeEnvVars(
  client: RuntimeClient,
  queryClient: QueryClient,
  envVars: Record<string, string>,
): Promise<{ newBlob: string; originalBlob: string }> {
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
      (
        error as {
          message?: string;
          response?: { data?: { message?: string } };
        }
      )?.message ??
      (
        error as {
          message?: string;
          response?: { data?: { message?: string } };
        }
      )?.response?.data?.message ??
      "";
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
